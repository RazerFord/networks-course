package main

import (
	"stop-and-wait/internal/network/client"
	"time"
)

func main() {
	cl, err := client.Connect("localhost", 9999, 1000*time.Millisecond)
	if err != nil {
		panic(err)
	}
	msg := "Hello world"
	cl.Write([]byte(msg))
}
