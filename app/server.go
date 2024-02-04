package main

import (
	"flag"
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

	dir := flag.String("dir", "", "The directory where RDB files are stored")
	dbfilename := flag.String("dbfilename", "", "The name of the RDB file")
	flag.Parse()

	startServer(serverOpt{dir: *dir, dbfilename: *dbfilename})
}

type serverOpt struct {
	dir        string
	dbfilename string
}

var cfg = make(map[string]string)

func startServer(opt serverOpt) {
	log.Println("StartServer...")

	cfg["dir"] = opt.dir
	cfg["dbfilename"] = opt.dbfilename

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
	switch m.cmd {
	case "ping":
		conn.Write([]byte("+PONG\r\n"))
		return nil

	case "echo":
		res := fmt.Sprintf("+%v\r\n", m.args[1])
		conn.Write([]byte(res))
		return nil

	case "set":
		res := onSet(m.args)
		conn.Write([]byte(res))
		return nil

	case "get":
		res := onGet(m.args)
		conn.Write([]byte(res))
		return nil

	case "config":
		res := onConfig(m.args)
		conn.Write([]byte(res))
		return nil

	default:
		return fmt.Errorf("unknown command")

	}
}

var data = make(map[string]string)

func onSet(args []string) string {
	if len(args) == 5 {
		data[args[1]] = args[3]
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
	return "+OK\r\n"
}

func onGet(args []string) string {
	val := data[args[1]]

	if len(val) == 0 {
		return "$-1\r\n"
	}

	return fmt.Sprintf("+%v\r\n", val)
}

func onConfig(args []string) string {
	key := args[1]
	val := cfg[args[3]]

	if len(val) == 0 {
		return "$-1\r\n"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*2\r\n$%d\r\n%s\r\n", len(key), key))
	sb.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val))
	return sb.String()

}
