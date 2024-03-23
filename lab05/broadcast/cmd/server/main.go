package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	addr := net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: 8080,
	}

	conn, err := net.DialUDP("udp", nil, &addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	writer := bufio.NewWriter(conn)
	for {
		time.Sleep(time.Second)
		now := time.Now()

		if _, err := writer.WriteString(now.GoString() + "\n"); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		writer.Flush()

		fmt.Printf("[ INFO ] %s\n", now.GoString())
	}
}
