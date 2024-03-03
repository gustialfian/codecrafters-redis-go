package main

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

// func TestParse(t *testing.T) {
// 	Parse("/home/highbits/code/sandbox/codecrafters-redis-go/dump/simple.rdb")
// }

func TestParseV2(t *testing.T) {
	rdb := ParseV2("/home/highbits/code/sandbox/codecrafters-redis-go/dump/dump.rdb")
	// fmt.Println(string(rdb.MagicString[:]), string(rdb.RDBVerNum[:]), rdb.AuxField, rdb.Databases)
	fmt.Printf("%+v\n", rdb)
}

func TestParseAux(t *testing.T) {
	str := "\x09\x72\x65\x64\x69\x73\x2d\x76\x65\x72\x05\x37\x2e\x32\x2e\x34\xfa\x0a\x72\x65\x64\x69\x73\x2d\x62\x69\x74\x73\xc0\x40"
	r := bufio.NewReader(strings.NewReader(str))
	key, val, err := parseAux(r)
	fmt.Println(key, val, err)
}
