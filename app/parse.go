package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
)

func Parse(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	r := bufio.NewReader(file)

	var keyVal []byte

	var state int
	for {
		b, err := r.ReadByte()
		// handle error
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		// handle header
		if b == 250 {
			state = 250
			continue
		} else if b == 254 {
			state = 254
			continue
		} else if b == 251 {
			state = 251
			continue
		} else if b == 255 {
			state = 255
			continue
		}

		// handle data
		if state == 250 {
			continue
		} else if state == 254 {
			continue
		} else if state == 251 {
			keyVal = append(keyVal, b)
			continue
		}
	}
	// fmt.Println("250:")
	// for _, v := range auxiliaryField {
	// 	fmt.Printf("%+q\n", v)
	// }
	// fmt.Println("254:")
	// for _, v := range dbInfo {
	// 	fmt.Printf("%q\n", v)
	// }
	// fmt.Println("251:")
	// for _, v := range keyVal {
	// 	fmt.Printf("%q ", v)
	// }
	// fmt.Println()

	db := make(map[string]string)
	mapingKV(keyVal, db)
}

func readNBytes(r io.Reader, n int) ([]byte, error) {
	result := make([]byte, n)
	n, err := r.Read(result)
	if err != nil {
		return []byte{}, err
	}

	return result[:n], nil
}

func mapingKV(b []byte, data map[string]string) {
	state := "start" // key, val
	var curKey []byte
	var curVal []byte
	for i := 3; i < len(b); i++ {
		fmt.Printf("%q -> %q: %q\n", b[i], curKey, curVal)

		// finite state
		if state == "start" && b[i] == 5 {
			state = "key"
			continue
		} else if state == "key" && b[i] == 5 {
			state = "val"
			continue
		}

		// act
		if state == "key" {
			curKey = append(curKey, b[i])
			continue
		} else if state == "val" {
			curVal = append(curVal, b[i])
			continue
		}
	}
}
