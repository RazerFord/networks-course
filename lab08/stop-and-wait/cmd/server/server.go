package main

import (
	"fmt"
	"stop-and-wait/internal/network/server"
	"time"
)

func main() {
	s, err := server.Connect("localhost", 9999, 1000*time.Millisecond)
	if err != nil {
		panic(err)
	}
	p := make([]byte, 1000)
	n, err := s.Read(p)
	if err != nil {
		panic(err)
	}
	fmt.Println(p[:n])
}
