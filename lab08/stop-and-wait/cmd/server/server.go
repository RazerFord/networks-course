package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"stop-and-wait/internal/network/server"
	"time"
)

func main() {
	addr := flag.String("address", "localhost", "server address")
	port := flag.Int("port", 8888, "server port")
	timeout := flag.Int64("timeout", 1000, "time-out")
	file := flag.String("file", "", "file name")
	flag.Parse()

	s, err := server.Connect(*addr, *port, time.Duration(*timeout)*time.Millisecond)
	exitIfNotNil(err)

	// receiving a message
	fmt.Println("Begin receiving")

	buff := make([]byte, 4)
	_, _, err = s.Read(buff)
	exitIfNotNil(err)

	l := binary.BigEndian.Uint32(buff)
	buff = make([]byte, l)
	_, _, err = s.Read(buff)
	exitIfNotNil(err)

	os.WriteFile(*file, buff, fs.FileMode(0777))

	// sending a message
	// fmt.Println("Begin sending")

	// size := binary.BigEndian.AppendUint32(nil, uint32(len(buff)))

	// _, err = s.Write(size, a)
	// exitIfNotNil(err)

	// _, err = s.Write(buff, a)
	// exitIfNotNil(err)
}

func exitIfNotNil(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
