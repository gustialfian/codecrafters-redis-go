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
		fmt.Println("net.Listen:", err.Error())
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("l.Accept:", err.Error())
			os.Exit(1)
		}

		go HandleCon(conn)
	}
}

func HandleCon(conn net.Conn) {
	for {
		b := make([]byte, 1024)
		n, err := conn.Read(b)
		if err != nil {
			fmt.Println("conn.Read:", err.Error())
			os.Exit(1)
		}

		msg := string(b[:n])
		fmt.Println("Received:", msg)
		if msg == "*1\r\n$4\r\nping\r\n" {
			conn.Write([]byte("+PONG\r\n"))
		}
	}
}
