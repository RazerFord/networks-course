package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"stop-and-wait/internal/network/server"
)

func main() {
	addr := flag.String("address", "localhost", "server address")
	port := flag.Int("port", 8888, "server port")
	file := flag.String("file", "", "file name")
	flag.Parse()

	s, err := server.Connect(*addr, *port)
	exitIfNil(err)

	buff := make([]byte, 4)
	s.Read(buff)

	l := binary.BigEndian.Uint32(buff)
	buff = make([]byte, l)

	s.Read(buff)
	os.WriteFile(*file, buff, fs.FileMode(777))
}

func exitIfNil(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
