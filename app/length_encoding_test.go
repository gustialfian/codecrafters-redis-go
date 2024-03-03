package main

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

func TestDecodeLength(t *testing.T) {
	str := "\x41\x01"
	r := bufio.NewReader(strings.NewReader(str))
	got, err := DecodeLength(r)
	fmt.Println(got, err)
}
