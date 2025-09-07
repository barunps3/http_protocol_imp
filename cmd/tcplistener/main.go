package main

import (
	"fmt"
	"httpfromtcp/internal/requests"
	"log"
	"net"
)

const port = ":42069"

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("error: %s\n", err.Error())
		}
		fmt.Println("Accepted connection from", conn.RemoteAddr())
		req, err := requests.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("Error parsing request: %s\n", err.Error())
		}
		fmt.Println("Request Line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)

		fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
	}

}
