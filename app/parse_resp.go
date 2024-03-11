package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type Message struct {
	cmd  string
	args []string
}

func ParseRESP(conn net.Conn) (Message, error) {
	r := bufio.NewReader(conn)
	b, err := r.ReadBytes('\n')
	if err != nil {
		return Message{}, err
	}

	if len(b) < 1 {
		return Message{}, errors.New("empty line")
	}

	if b[0] != '*' {
		return Message{}, errors.New("not impl first command not array")
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
		return Message{}, fmt.Errorf("invalid ararys length: %w", err)
	}

	if length < 1 {
		return Message{}, errors.New("empty command")
	}

	msg := Message{
		args: make([]string, length-1),
	}

	for i := 0; i < length; i++ {
		var data string
		b, err := r.ReadByte()
		if err != nil {
			return Message{}, err
		}

		switch b {
		case '$': // bulk string
			lengthByte, err := readUntilCRLF(r)
			if err != nil {
				return Message{}, fmt.Errorf("failed reading bulk string length: %w", err)
			}
			length, err := strconv.Atoi(string(lengthByte))
			if err != nil {
				return Message{}, fmt.Errorf("invalid bulk string length: %w", err)
			}

			data, err = readBulkString(r, length)
			if err != nil {
				return Message{}, fmt.Errorf("failed to read bulkstring: %w", err)
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

func makeArrayBulkString(s []string) string {
	var result strings.Builder
	result.WriteString(fmt.Sprintf("*%d\r\n", len(s)))
	for _, v := range s {
		result.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
	}
	return result.String()
}
