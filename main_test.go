package main

import (
	"net"
	"testing"
	"time"
)

func TestHandleRequest(t *testing.T) {
	request := []byte{
		0x12, 0x34, 0x81, 0x80, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, // Header
		0x06, 'g', 'o', 'o', 'g', 'l', 'e', 0x03, 'c', 'o', 'm', 0x00, // Question
		0x00, 0x01, 0x00, 0x01, // Type A, Class IN
	}

	udpServer := &mockPacketConn{}
	addr := &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 12345}

	domain, qType, qClass, err := handleRequest(udpServer, addr, request)

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
