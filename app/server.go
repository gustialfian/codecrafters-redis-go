package main

import (
	"bufio"
	"fmt"
	"io"
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
			fmt.Println("Error accepting connection:", err.Error())
			os.Exit(1)
		}

		go HandleCon(conn)
	}
}

func HandleCon(conn net.Conn) {
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	for {
		l, _, err := r.ReadLine()
		if len(l) < 3 {
			continue
		}
		if err == io.EOF {
			break
		}

		fmt.Println("Received:", string(l))
		_, err = conn.Write([]byte("+PONG\r\n"))
		if err != nil {
			w.Flush()
		}
	}
}
