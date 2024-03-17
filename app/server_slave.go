package main

import (
	"fmt"
	"net"
)

type SlaveServer struct {
	host string
	conn net.Conn
}

func (ss *SlaveServer) Connect() error {
	conn, err := net.Dial("tcp", ss.host)
	if err != nil {
		return fmt.Errorf("Connect: %w", err)
	}
	ss.conn = conn
	// fmt.Println("Connect")
	return nil
}

func (ss *SlaveServer) Send(b string) (Message, error) {
	_, err := ss.conn.Write([]byte(b))
	if err != nil {
		return Message{}, fmt.Errorf("Send: %w", err)
	}

	m, err := ParseRESP(ss.conn)
	if err != nil {
		return Message{}, fmt.Errorf("Send: %w", err)
	}
	// fmt.Println("Send")
	return m, nil
}
