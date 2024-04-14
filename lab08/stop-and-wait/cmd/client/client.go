package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"stop-and-wait/internal/network/client"
	"strings"
	"time"
)

func main() {
	addr := flag.String("address", "localhost", "server address")
	port := flag.Int("port", 8888, "server port")
	timeout := flag.Int64("timeout", 100, "time-out")
	file := flag.String("file", "", "file name")
	flag.Parse()

	p, err := os.ReadFile(*file)
	exitIfNotNil(err)

	c, err := client.Connect(*addr, *port, time.Duration(*timeout)*time.Millisecond)
	exitIfNotNil(err)

	// sending a message
	fmt.Println("Begin sending")

	s := binary.BigEndian.AppendUint32(nil, uint32(len(p)))

	_, err = c.Write(s)
	exitIfNotNil(err)

	_, err = c.Write(p)
	exitIfNotNil(err)

	// receiving a message
	fmt.Println("Begin receiving")

	p = make([]byte, 4)

	_, err = c.Read(p)
	exitIfNotNil(err)

	l := binary.BigEndian.Uint32(p)
	p = make([]byte, l)

	_, err = c.Read(p)
	exitIfNotNil(err)

	arr := strings.Split(*file, ".")
	os.WriteFile(fmt.Sprintf("received.%s", arr[len(arr)-1]), p, fs.FileMode(0777))
	time.Sleep(5 * time.Second)
}

func exitIfNotNil(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
