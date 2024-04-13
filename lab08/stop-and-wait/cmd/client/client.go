package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"stop-and-wait/internal/network/client"
	"time"
)

func main() {
	addr := flag.String("address", "localhost", "server address")
	port := flag.Int("port", 8888, "server port")
	timeout := flag.Int64("timeout", 100, "time-out")
	file := flag.String("file", "", "file name")
	flag.Parse()

	p, err := os.ReadFile(*file)
	exitIfNil(err)

	client, err := client.Connect(*addr, *port, time.Duration(*timeout)*time.Millisecond)
	exitIfNil(err)

	s := binary.BigEndian.AppendUint32(nil, uint32(len(p)))
	client.Write(s)
	client.Write(p)
}

func exitIfNil(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
