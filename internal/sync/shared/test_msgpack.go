//go:build ignore

package main

import (
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
)

func main() {
	// Test 1: Direct type
	data := []string{"test1", "test2"}
	_, err := msgpack.Marshal(data)
	if err != nil {
		fmt.Printf("ERROR (direct): %v\n", err)
	} else {
		fmt.Println("SUCCESS (direct)")
	}

	// Test 2: Through interface{}
	var data2 interface{} = []string{"test1", "test2"}
	_, err = msgpack.Marshal(data2)
	if err != nil {
		fmt.Printf("ERROR (interface{}): %v\n", err)
	} else {
		fmt.Println("SUCCESS (interface{})")
	}
}
