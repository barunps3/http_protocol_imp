package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	lineChan := make(chan string)
	go func() {
		defer f.Close()
		defer close(lineChan)
		buffer := make([]byte, 8)
		var lineBegin string = ""
		for {
			n, err := f.Read(buffer)
			if err != nil {
				fmt.Println("Error:", err)
				if lineBegin != "" {
					lineChan <- lineBegin
					lineBegin = ""
				}
				if err == io.EOF {
					break
				}
				break
			}
			str := string(buffer[:n])
			parts := strings.Split(str, "\n")

			for i := range len(parts) - 1 {
				lineChan <- lineBegin + parts[i]
				lineBegin = ""
			}
			lineBegin += parts[len(parts)-1]
		}
	}()

	return lineChan
}

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

		linesChan := getLinesChannel(conn)

		for line := range linesChan {
			fmt.Println(line)
		}
		fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
	}

}
