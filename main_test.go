package main

import (
	"bytes"
	"encoding/binary"
	"net"
	"os"
	"testing"
	"time"
)

func TestHandleRequest(t *testing.T) {
	request := []byte{
		0x12, 0x34, 0x81, 0x80, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, // Header
		0x06, 'g', 'o', 'o', 'g', 'l', 'e', 0x03, 'c', 'o', 'm', 0x00, // Question
		0x00, 0x01, 0x00, 0x01, // Type A, Class IN
	}

	domain, qType, qClass, err := handleRequest(request)

	if err != nil {
		t.Fatalf("Error in handling request: %v", err)
	}

	expectedDomain := "google.com"

	if domain != expectedDomain {
		t.Errorf("Did not get expected domain. Got: %v, expected: %v", domain, expectedDomain)
	}

	expectedQType := uint16(1)
	if qType != expectedQType {
		t.Errorf("Did not get expected question type. Got: %v, expected: %v", qType, expectedQType)
	}

	expectedQClass := uint16(1)
	if qClass != expectedQClass {
		t.Errorf("Did not get expected question class. Got: %v, expected: %v", qClass, expectedQClass)
	}
}

func TestSearchDomain(t *testing.T) {
	//continutul fisierului meu de zona
	zoneFileContent := `$TTL 86400
@    IN  SOA   ns1.dns-server.com. admin.dns-server.com. (
               2024103001
               3600
               1800
               1209600
               86400
               )

     IN  NS    ns1.dns-server.com.
     IN  NS    ns2.dns-server.com.

@    IN  A     192.0.2.1
www  IN  A     192.0.2.2
mail IN  A     192.0.2.3
`

	//creem un fisier de zona temporar
	tempFile, err := os.CreateTemp("", "zone.txt")
	if err != nil {
		t.Fatalf("Error in creating temp zone file: %v", err)
	}

	//stergem la final fisierul
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(zoneFileContent)
	if err != nil {
		t.Fatalf("Eroare la scrierea în fișierul temporar: %v", err)
	}
	tempFile.Close()

	// testam cautarea domeniului principal "@"
	ip, err := searchDomain(tempFile.Name(), "dns-server.com")
	if err != nil {
		t.Errorf("Error in looking for domain: %v", err)
	}
	if ip != "192.0.2.1" {
		t.Errorf("Error in getting correct ip address: got %s, want %s", ip, "192.0.2.1")
	}

	// testam cautarea subdomeniului "www"
	ip, err = searchDomain(tempFile.Name(), "www")
	if err != nil {
		t.Errorf("Error in looking for domain www: %v", err)
	}
	if ip != "192.0.2.2" {
		t.Errorf("Error in getting correct ip address www: got %s, want %s", ip, "192.0.2.2")
	}

	// testam cautarea subdomeniului "mail"
	ip, err = searchDomain(tempFile.Name(), "mail")
	if err != nil {
		t.Errorf("Error in looking for domain mail: %v", err)
	}
	if ip != "192.0.2.3" {
		t.Errorf("Error in getting correct ip address mail: got %s, want %s", ip, "192.0.2.3")
	}

	// testam cautarea unui subdomeniu inexistent
	_, err = searchDomain(tempFile.Name(), "nonexistent")
	if err == nil {
		t.Error("Should have thrown an error here")
	}
}

func TestSendResponse(t *testing.T) {

	domain := "google.com"
	ip := "142.250.201.206"
	addr := &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 53}

	request := []byte{
		0x12, 0x34, 0x81, 0x80, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, // Header
		0x06, 'g', 'o', 'o', 'g', 'l', 'e', 0x03, 'c', 'o', 'm', 0x00, // Question
		0x00, 0x01, 0x00, 0x01, // Type A, Class IN
	}

	mockConn := &mockPacketConn{}

	err := sendResponse(mockConn, addr, domain, ip, request)
	if err != nil {
		t.Fatalf("Error in sending the response: %v", err)
	}

	response := mockConn.buffer.Bytes()

	// verificam header-ul
	if len(response) < 12 {
		t.Fatalf("Response sent is too short: %v", response)
	}
	id := binary.BigEndian.Uint16(response[0:2])
	flags := binary.BigEndian.Uint16(response[2:4])
	qdCount := binary.BigEndian.Uint16(response[4:6])
	anCount := binary.BigEndian.Uint16(response[6:8])

	if id != 0x1234 {
		t.Errorf("Incorrect ID: got %x, want %x", id, 0x1234)
	}
	if flags != 0x8180 {
		t.Errorf("Incorrect Flags: got %x, want %x", flags, 0x8180)
	}
	if qdCount != 1 {
		t.Errorf("Incorrect question count: got %d, want %d", qdCount, 1)
	}
	if anCount != 1 {
		t.Errorf("Incorrect answer number: got %d, want %d", anCount, 1)
	}

	// sarim peste header
	offset := 12
	expectedQuestion := request[12:]
	if !bytes.Equal(response[offset:offset+len(expectedQuestion)], expectedQuestion) {
		t.Errorf("Malformed question section: got %v, want %v", response[offset:offset+len(expectedQuestion)], expectedQuestion)
	}

	// Verificăm secțiunea Answer
	offset += len(expectedQuestion)
	// Check domain name
	expectedDomain := []byte{0x06, 'g', 'o', 'o', 'g', 'l', 'e', 0x03, 'c', 'o', 'm', 0x00}
	if !bytes.Equal(response[offset:offset+len(expectedDomain)], expectedDomain) {
		t.Errorf("Incorrec domain name: got %v, want %v", response[offset:offset+len(expectedDomain)], expectedDomain)
	}
	offset += len(expectedDomain)

	// Check type and class
	expectedType := uint16(1)  // Type A
	expectedClass := uint16(1) // Class IN
	if binary.BigEndian.Uint16(response[offset:offset+2]) != expectedType {
		t.Errorf("Incorrect type: got %x, want %x", binary.BigEndian.Uint16(response[offset:offset+2]), expectedType)
	}
	if binary.BigEndian.Uint16(response[offset+2:offset+4]) != expectedClass {
		t.Errorf("Incorrect class: got %x, want %x", binary.BigEndian.Uint16(response[offset+2:offset+4]), expectedClass)
	}
	offset += 4

	// Check TTL
	expectedTTL := uint32(300)
	expectedDataLen := uint16(4)
	if binary.BigEndian.Uint32(response[offset:offset+4]) != expectedTTL {
		t.Errorf("Incorrect TTL: got %x, want %x", binary.BigEndian.Uint32(response[offset:offset+4]), expectedTTL)
	}
	if binary.BigEndian.Uint16(response[offset+4:offset+6]) != expectedDataLen {
		t.Errorf("Incorrect length of answers: got %x, want %x", binary.BigEndian.Uint16(response[offset+4:offset+6]), expectedDataLen)
	}
	offset += 6

	// Check IP
	expectedIP := net.ParseIP(ip).To4()
	if !bytes.Equal(response[offset:offset+4], expectedIP) {
		t.Errorf("Adresa IP incorectă în secțiunea Answer: got %v, want %v", response[offset:offset+4], expectedIP)
	}
}

// mockPacketConn este un mock pentru net.PacketConn
type mockPacketConn struct {
	buffer bytes.Buffer
}

func (m *mockPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	return 0, nil, nil
}

func (m *mockPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	return m.buffer.Write(p)
}

func (m *mockPacketConn) Close() error                       { return nil }
func (m *mockPacketConn) LocalAddr() net.Addr                { return &net.UDPAddr{} }
func (m *mockPacketConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockPacketConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockPacketConn) SetWriteDeadline(t time.Time) error { return nil }
