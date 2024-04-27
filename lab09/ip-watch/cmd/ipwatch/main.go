package main

import (
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/google/gopacket/routing"
)

const GoogleIP = "8.8.8.8"

var errNotFound = errors.New("nof found")

func main() {
	router, err := routing.New()
	exitIfNotNil(err)

	googleIP := net.ParseIP(GoogleIP)
	_, _, p, err := router.Route(googleIP)
	exitIfNotNil(err)

	addrs, err := net.InterfaceAddrs()
	exitIfNotNil(err)

	for _, addr := range addrs {
		if addr, ok := addr.(*net.IPNet); ok && !addr.IP.IsLoopback() {
			if addr.IP.Equal(p) {
				fmt.Printf("ip   %s\n", p)
				fmt.Printf("mask %s\n", addr.Mask.String())
				return
			}
		}
	}
	
	exitIfNotNil(errNotFound)
}

func exitIfNotNil(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
