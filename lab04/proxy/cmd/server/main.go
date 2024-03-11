package main

import (
	"flag"
	"fmt"
	"os"
	"razer-ford/proxy-server/internal/pkg/server"
)

func main() {
	port := flag.Int("p", 8080, "Provide a port number")
	address := flag.String("s", "localhost", "Provide host address")

	flag.Parse()

	ps := server.NewProxyServer(*port, *address)

	if nil != ps.Run() {
		fmt.Println("server error")
		os.Exit(1)
	}
}
