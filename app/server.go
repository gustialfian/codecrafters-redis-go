package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	StartServer()
}

func StartServer() {
	fmt.Println("StartServer...")
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
	args []string
}

func parse(conn net.Conn) (message, error) {
	b := make([]byte, 1024)
	n, err := conn.Read(b)
	if err != nil {
		return message{}, err
	}
	s := string(b[:n])

	lines := strings.Split(s, "\r\n")
	cmd := lines[2]
	args := lines[3:]

	return message{
		cmd:  cmd,
		args: args,
	}, nil
}

func runMessage(conn net.Conn, m message) error {
	if m.cmd == "ping" {
		val := "PONG"

		res := fmt.Sprintf("+%v\r\n", val)
		conn.Write([]byte(res))
		return nil
	}
	if m.cmd == "echo" {
		res := fmt.Sprintf("+%v\r\n", m.args[1])
		conn.Write([]byte(res))
		return nil
	}
	if m.cmd == "set" {
		onSet(m.args)
		val := "OK"

		res := fmt.Sprintf("+%v\r\n", val)
		conn.Write([]byte(res))
		return nil
	}
	if m.cmd == "get" {
		val := onGet(m.args)

		res := fmt.Sprintf("+%v\r\n", val)
		conn.Write([]byte(res))
		return nil
	}
	return fmt.Errorf("unknown command")
}

var data = make(map[string]string)

func onSet(args []string) {
	if len(args) == 5 {
		data[args[1]] = args[3]
		return
	}
	if len(args) == 9 {
		data[args[1]] = args[3]

		ttl, err := strconv.ParseInt(args[7], 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			<-time.After(time.Duration(ttl) * time.Millisecond)
			delete(data, args[1])
		}()
	}
}

func onGet(args []string) string {
	return data[args[1]]
}
