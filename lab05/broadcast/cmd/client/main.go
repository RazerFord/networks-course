package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	addr := net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: 8080,
	}
	for {
		conn, err := net.ListenUDP("udp", &addr)
		if err != nil {
			continue
		}
		reader := bufio.NewReader(conn)
		now, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Print(now)
	}
}
