package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("net.Listen:", err.Error())
		os.Exit(1)
	}
	defer l.Close()

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
		m, err := parse(conn)
		if err != nil {
			fmt.Println(err.Error())
			break
		}

		err = runMessage(conn, m)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
	}
}

type message struct {
	cmd  string
	args string
}

func parse(conn net.Conn) (message, error) {
	b := make([]byte, 1024)
	n, err := conn.Read(b)
	if err != nil {
		return message{}, err
	}
	s := string(b[:n])

	fmt.Println("Received:", s)

	lines := strings.Split(s, "\r\n")
	cmd := lines[2]
	args := lines[3]

	return message{
		cmd:  cmd,
		args: args,
	}, nil
}

func runMessage(conn net.Conn, m message) error {
	if m.cmd == "ping" {
		conn.Write([]byte("+PONG\r\n"))
		return nil
	}
	if m.cmd == "echo" {
		res := fmt.Sprintf("+%v\r\n", m.args)
		conn.Write([]byte(res))
		return nil
	}
	return fmt.Errorf("unknown command")
}
