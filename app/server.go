package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go HandleCon(conn)
	}
}

func HandleCon(conn net.Conn) {
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading connection: ", err.Error())
			return
		}

		if string(buf[:n]) == "*1\r\n$4\r\nping\r\n" {
			conn.Write([]byte("+PONG\r\n"))
			continue
		}

		conn.Write([]byte("unknown msg\r\n"))
	}
}
