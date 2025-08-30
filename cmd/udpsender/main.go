package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	serverAddr := "localhost:42069"
	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		log.Fatal(err)
	}
	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer udpConn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input %v\n", err)
			os.Exit(1)
		}

		_, err = udpConn.Write([]byte(message))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Message sent: %s", message)
	}
}
