package main

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestStartServer(t *testing.T) {
	go startServer()
	time.Sleep(time.Millisecond)

	conn, err := net.Dial("tcp", "0.0.0.0:6379")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		input  string
		expect string
		wait   time.Duration
	}{
		{
			name:   "ping",
			input:  makeArrayString([]string{"ping"}),
			expect: "+PONG\r\n",
		},
		{
			name:   "echo",
			input:  makeArrayString([]string{"echo", "foobarbaz"}),
			expect: "+foobarbaz\r\n",
		},
		{
			name:   "set",
			input:  makeArrayString([]string{"set", "hello", "world"}),
			expect: "+OK\r\n",
		},
		{
			name:   "get",
			input:  makeArrayString([]string{"get", "hello"}),
			expect: "+world\r\n",
		},
		{
			name:   "get_not_found",
			input:  makeArrayString([]string{"get", "get_not_found"}),
			expect: "$-1\r\n",
		},
		{
			name:   "set_with_expiry",
			input:  makeArrayString([]string{"set", "expiry", "123", "px", "10"}),
			expect: "+OK\r\n",
		},
		{
			name:   "get_with_expiry",
			input:  makeArrayString([]string{"get", "expiry"}),
			expect: "+123\r\n",
			wait:   11 * time.Millisecond,
		},
		{
			name:   "get_with_expiry_not_found",
			input:  makeArrayString([]string{"get", "expiry"}),
			expect: "$-1\r\n",
		},
		{
			name:   "get_config",
			input:  makeArrayString([]string{"config", "get", "dir"}),
			expect: makeArrayString([]string{"dir", "/home/highbits/code/sandbox/codecrafters-redis-go/dump"}),
		},
		{
			name:   "get_keys",
			input:  makeArrayString([]string{"keys", "*"}),
			expect: makeArrayString([]string{"foo", "bar"}),
		},
		{
			name:   "info_replication",
			input:  makeArrayString([]string{"info", "replication"}),
			expect: makeArrayString([]string{"# Replication\nrole:master"}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err = conn.Write([]byte(tt.input))
			if err != nil {
				t.Fatal(err)
			}

			res := make([]byte, 1024)
			n, err := conn.Read(res)
			if err != nil {
				t.Fatal(err)
			}

			if string(res[:n]) != tt.expect {
				t.Fatalf("expected %q got %q", tt.expect, string(res[:n]))
			}
		})
		<-time.After(tt.wait)
	}
}

func makeArrayString(s []string) string {
	var result strings.Builder
	result.WriteString(fmt.Sprintf("*%d\r\n", len(s)))
	for _, v := range s {
		result.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
	}
	return result.String()
}
