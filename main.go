package main

import (
	"fmt"
	"log"
	"net"
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
		_, addr, err := udpServer.ReadFrom(buff)
		if err != nil {
			fmt.Println(fmt.Errorf("Error in processing a read: %v", err))
			continue
		}
		//creem un go routine care sa prelucreze separat requestul
		go handleRequest(udpServer, addr, buff)
	}
}

/*
*	handleRequest: functie care va prelucra reqeustul primit la serverul DNS
 */
func handleRequest(udpServer net.PacketConn, addr net.Addr, buff []byte) (string, uint16, uint16, error) {
	return "", uint16(1), uint16(1), nil
}
