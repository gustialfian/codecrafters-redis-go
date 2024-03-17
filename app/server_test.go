package main

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestStartServer(t *testing.T) {
	go startServer(ServerOpt{
		port:       "6379",
		dir:        "/home/highbits/code/sandbox/codecrafters-redis-go/dump",
		dbfilename: "dump.rdb",
	})
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
			input:  makeArrayBulkString([]string{"ping"}),
			expect: "+PONG\r\n",
		},
		{
			name:   "echo",
			input:  makeArrayBulkString([]string{"echo", "foobarbaz"}),
			expect: "+foobarbaz\r\n",
		},
		{
			name:   "set",
			input:  makeArrayBulkString([]string{"set", "hello", "world"}),
			expect: "+OK\r\n",
		},
		{
			name:   "get",
			input:  makeArrayBulkString([]string{"get", "hello"}),
			expect: "+world\r\n",
		},
		{
			name:   "get_not_found",
			input:  makeArrayBulkString([]string{"get", "get_not_found"}),
			expect: "$-1\r\n",
		},
		{
			name:   "set_with_expiry",
			input:  makeArrayBulkString([]string{"set", "expiry", "123", "px", "10"}),
			expect: "+OK\r\n",
		},
		{
			name:   "get_with_expiry",
			input:  makeArrayBulkString([]string{"get", "expiry"}),
			expect: "+123\r\n",
			wait:   11 * time.Millisecond,
		},
		{
			name:   "get_with_expiry_not_found",
			input:  makeArrayBulkString([]string{"get", "expiry"}),
			expect: "$-1\r\n",
		},
		{
			name:   "get_config",
			input:  makeArrayBulkString([]string{"config", "get", "dir"}),
			expect: makeArrayBulkString([]string{"dir", "/home/highbits/code/sandbox/codecrafters-redis-go/dump"}),
		},
		{
			name:   "get_keys",
			input:  makeArrayBulkString([]string{"keys", "*"}),
			expect: makeArrayBulkString([]string{"foo", "bar"}),
		},
		{
			name:   "info_replication",
			input:  makeArrayBulkString([]string{"info", "replication"}),
			expect: makeBulkString("# Replication\nrole:master\nmaster_replid:8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb\nmaster_repl_offset:0"),
		},
		{
			name:   "replconf_listening_port",
			input:  makeArrayBulkString([]string{"REPLCONF", "listening-port", "6380"}),
			expect: "+OK\r\n",
		},
		{
			name:   "replconf_capa",
			input:  makeArrayBulkString([]string{"REPLCONF", "capa", "eof", "capa", "psync2"}),
			expect: "+OK\r\n",
		},
		{
			name:   "psync_init",
			input:  makeArrayBulkString([]string{"PSYNC", "?", "-1"}),
			expect: "+FULLRESYNC 8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb 0\r\n",
			// wait:   20 * time.Millisecond,
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

func makeBulkString(s string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(s), s)
}
