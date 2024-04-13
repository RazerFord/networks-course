package main

import (
	"fmt"
	"stop-and-wait/internal/network/server"
)

func main() {
	s, err := server.Connect("localhost", 9999)
	if err != nil {
		panic(err)
	}
	p := make([]byte, 1000)
	n, err := s.Read(p)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(p[:n]))
}
