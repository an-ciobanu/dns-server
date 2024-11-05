package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
)

const (
	UDP  = "udp"
	port = ":53"
)

func main() {
	udpServer, err := net.ListenPacket(UDP, port)
	if err != nil {
		log.Fatal(fmt.Errorf("Fatal Error: %v", err))
	}

	defer udpServer.Close() //inchidem serverul cu defer

	for {
		buff := make([]byte, 1024) //creeam bufferul in care primim requestul la server
		n, addr, err := udpServer.ReadFrom(buff)
		if err != nil {
			fmt.Println(fmt.Errorf("Error in processing a read: %v", err))
			continue
		}
		//creem un go routine care sa prelucreze separat requestul
		go func(addr net.Addr, buff []byte) {
			domain, qType, qClass, err := handleRequest(buff[:n]) // prelucrăm doar partea utilizată din buffer
			if err != nil {
				fmt.Println(fmt.Errorf("Error in handling q reqeust: %v", err))
			}
			sendResponse(udpServer, addr, domain, qType, qClass)
		}(addr, buff)
	}
}

/*
*	handleRequest: functie care va prelucra reqeustul primit la serverul DNS
 */
func handleRequest(buff []byte) (string, uint16, uint16, error) {

	if len(buff) < 12 {
		//cererea e invalida, nu avem nici macar headerul corect
		return "", 0, 0, errors.New("Invalid DNS query")
	}

	//sarim peste header
	offset := 12

	domainParts := []string{}

	for {
		//preluam prima lungime de parte din domeniu
		length := int(buff[offset])
		if length == 0 {
			offset++
			break
		}
		offset++
		//am gasit u bucata de text din domeniu
		domainParts = append(domainParts, string(buff[offset:offset+length]))
		offset += length
	}

	domain := strings.Join(domainParts, ".")

	// Tipul și clasa cererii
	qType := binary.BigEndian.Uint16(buff[offset : offset+2])
	qClass := binary.BigEndian.Uint16(buff[offset+2 : offset+4])

	return domain, qType, qClass, nil
}

func searchDomain(filename string, domain string) (string, error) {
	return "192.168.100.101", nil
}

func sendResponse(udpServer net.PacketConn, addr net.Addr, domain string, qType uint16, qClass uint16) {
}
