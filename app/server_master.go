package main

import (
	"fmt"
	"net"
)

type MasterServer struct {
	host string
	conn net.Conn
}

func (ms *MasterServer) Connect() error {
	conn, err := net.Dial("tcp", ms.host)
	if err != nil {
		return fmt.Errorf("Connect: %w", err)
	}
	ms.conn = conn
	// fmt.Println("Connect")
	return nil
}

func (ms *MasterServer) Send(b string) (Message, error) {
	_, err := ms.conn.Write([]byte(b))
	if err != nil {
		return Message{}, fmt.Errorf("Send: %w", err)
	}

	m, err := ParseRESP(ms.conn)
	if err != nil {
		return Message{}, fmt.Errorf("Send: %w", err)
	}
	// fmt.Println("Send")
	return m, nil
}
