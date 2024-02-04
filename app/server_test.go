package main

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestStartServer(t *testing.T) {
	go startServer(serverOpt{dir: "/tmp/redis-files", dbfilename: "dump.rdb"})
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
			input:  makeCmd([]string{"ping"}),
			expect: "+PONG\r\n",
		},
		{
			name:   "echo",
			input:  makeCmd([]string{"echo", "foobarbaz"}),
			expect: "+foobarbaz\r\n",
		},
		{
			name:   "set",
			input:  makeCmd([]string{"set", "hello", "world"}),
			expect: "+OK\r\n",
		},
		{
			name:   "get",
			input:  makeCmd([]string{"get", "hello"}),
			expect: "+world\r\n",
		},
		{
			name:   "get_not_found",
			input:  makeCmd([]string{"get", "get_not_found"}),
			expect: "$-1\r\n",
		},
		{
			name:   "set_with_expiry",
			input:  makeCmd([]string{"set", "expiry", "123", "px", "10"}),
			expect: "+OK\r\n",
		},
		{
			name:   "get_with_expiry",
			input:  makeCmd([]string{"get", "expiry"}),
			expect: "+123\r\n",
			wait:   11 * time.Millisecond,
		},
		{
			name:   "get_with_expiry_not_found",
			input:  makeCmd([]string{"get", "expiry"}),
			expect: "$-1\r\n",
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
				t.Fatalf("expected %v got %v", string(res[:n]), tt.expect)
			}
		})
		<-time.After(tt.wait)
	}
}

func makeCmd(s []string) string {
	var result strings.Builder
	result.WriteString("*0\r\n")
	for _, v := range s {
		result.WriteString(fmt.Sprintf("$0\r\n%s\r\n", v))
	}

	return result.String()
}
