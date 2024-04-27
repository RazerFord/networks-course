package main

import (
	"flag"
	"fmt"
	"net"
)

const maxPort = 1 << 16 // :)

func main() {
	ip := flag.String("ip", "0.0.0.0", "ip address")
	start := flag.Int("start", 0, "start of port ranges")
	end := flag.Int("end", maxPort, "end of port range (excluding)")
	flag.Parse()

	ports := []int{}
	for port := range *end - *start {
		port += *start

		addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", *ip, port))
		if err != nil {
			continue
		}

		conn, err := net.ListenTCP("tcp", addr)
		if err != nil {
			continue
		}
		conn.Close()

		ports = append(ports, port)
	}
	fmt.Printf("ip: %s\n", *ip)
	fmt.Printf("free ports: %v\n", ports)
}
