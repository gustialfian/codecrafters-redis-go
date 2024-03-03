package main

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

func Test_decodeInt(t *testing.T) {
	str := "\x0F"
	r := bufio.NewReader(strings.NewReader(str))

	got, err := decodeInt(r, 8)
	fmt.Println(got, err)

	str = "\x00\xFF"
	r = bufio.NewReader(strings.NewReader(str))
	got, err = decodeInt(r, 16)
	fmt.Println(got, err)

	str = "\x00\xFF\xFF\xFF"
	r = bufio.NewReader(strings.NewReader(str))
	got, err = decodeInt(r, 32)
	fmt.Println(got, err)
}
