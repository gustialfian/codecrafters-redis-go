package main

import (
	"bufio"
	"errors"
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
var rdb RDB

func startServer(opt serverOpt) {
	log.Println("StartServer...")

	cfg["dir"] = opt.dir
	cfg["dbfilename"] = opt.dbfilename
	path := fmt.Sprintf("./%v/%v", opt.dir, opt.dbfilename)
	rdb = ParseV2(path)

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
	r := bufio.NewReader(conn)
	b, err := r.ReadBytes('\n')
	if err != nil {
		return message{}, err
	}

	if len(b) < 1 {
		return message{}, errors.New("empty line")
	}

	if b[0] != '*' {
		return message{}, errors.New("not impl first command not array")
	}

	readUntilCRLF := func(r *bufio.Reader) ([]byte, error) {
		b, err := r.ReadBytes('\n')
		if err != nil {
			return b, err
		}

		if len(b) < 1 {
			return b, errors.New("empty line")
		}

		if !strings.HasSuffix(string(b), "\r\n") {
			return b, errors.New("not ended with CRLF")
		}

		length := len(b)
		return b[:length-2], nil
	}

	lengthBytes := b[1:]
	lengthStr := string(lengthBytes[:len(lengthBytes)-2]) // remove the CRLF
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return message{}, fmt.Errorf("invalid ararys length: %w", err)
	}

	if length < 1 {
		return message{}, errors.New("empty command")
	}

	msg := message{
		args: make([]string, length-1),
	}

	for i := 0; i < length; i++ {
		var data string
		b, err := r.ReadByte()
		if err != nil {
			return message{}, err
		}

		switch b {
		case '$': // bulk string
			lengthByte, err := readUntilCRLF(r)
			if err != nil {
				return message{}, fmt.Errorf("failed reading bulk string length: %w", err)
			}
			length, err := strconv.Atoi(string(lengthByte))
			if err != nil {
				return message{}, fmt.Errorf("invalid bulk string length: %w", err)
			}

			data, err = readBulkString(r, length)
			if err != nil {
				return message{}, fmt.Errorf("failed to read bulkstring: %w", err)
			}
		}

		if i == 0 {
			msg.cmd = data
			continue
		}

		msg.args[i-1] = data
	}

	return msg, nil
}

func readBulkString(r *bufio.Reader, length int) (string, error) {
	buf := make([]byte, length)

	if _, err := r.Read(buf); err != nil {
		return "", err
	}

	if _, err := r.Read(make([]byte, 2)); err != nil {
		return "", err
	}

	return string(buf), nil
}

func runMessage(conn net.Conn, m message) error {
	var resp string
	switch m.cmd {
	case "ping":
		resp = "+PONG\r\n"
	case "echo":
		resp = fmt.Sprintf("+%v\r\n", m.args[0])
	case "set":
		resp = onSet(m.args)
	case "get":
		resp = onGet(m.args)
	case "config":
		resp = onConfig(m.args)
	case "keys":
		// redis-cli keys "*"
		resp = onKeys(m.args)
	default:
		return fmt.Errorf("unknown command")
	}

	_, err := conn.Write([]byte(resp))
	return err
}

var data = make(map[string]string)

func onSet(args []string) string {
	if len(args) == 2 {
		data[args[0]] = args[1]
	}
	if len(args) == 4 {
		data[args[0]] = args[1]

		ttl, err := strconv.ParseInt(args[3], 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			<-time.After(time.Duration(ttl) * time.Millisecond)
			delete(data, args[0])
		}()
	}
	return "+OK\r\n"
}

func onGet(args []string) string {
	val := data[args[0]]

	if len(val) == 0 {
		return "$-1\r\n"
	}

	return fmt.Sprintf("+%v\r\n", val)
}

func onConfig(args []string) string {
	key := args[1]
	val := cfg[args[1]]

	if len(val) == 0 {
		return "$-1\r\n"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*2\r\n$%d\r\n%s\r\n", len(key), key))
	sb.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val))
	return sb.String()
}

func onKeys(args []string) string {
	switch args[0] {
	case "*":
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("*%d\r\n", len(rdb.Databases[0].Fields)))
		for _, f := range rdb.Databases[0].Fields {
			sb.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(f.Key), f.Key))
		}
		return sb.String()
	}
	return "*0"
}
