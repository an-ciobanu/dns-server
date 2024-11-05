package main

import (
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

// mockPacketConn este un mock pentru net.PacketConn
type mockPacketConn struct{}

func (m *mockPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	return 0, nil, nil
}

func (m *mockPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	return len(p), nil
}

func (m *mockPacketConn) Close() error                       { return nil }
func (m *mockPacketConn) LocalAddr() net.Addr                { return &net.UDPAddr{} }
func (m *mockPacketConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockPacketConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockPacketConn) SetWriteDeadline(t time.Time) error { return nil }
