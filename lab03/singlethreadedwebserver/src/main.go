package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
)

var logger = log.Default()

const reqParam = "file"

func main() {

	if len(os.Args) < 2 {
		panic("pass the port")
	}

	port, err := strconv.Atoi(os.Args[1])

	if err != nil {
		panic("port parsing error")
	}

	logger.Printf("[ INFO ] port %d received\n", port)

	addr := net.TCPAddr{}
	addr.IP = net.ParseIP("127.0.0.1")
	addr.Port = port

	server, err := net.ListenTCP("tcp", &addr)

	if err != nil {
		panic("connection error")
	}

	logger.Printf("[ INFO ] server created: http://%v:%v \n", addr.IP, addr.Port)

	tcp, err := server.AcceptTCP()

	if err != nil {
		panic("connection error")
	}

	logger.Printf("[ INFO ] connection accepted\n")

	reader := bufio.NewReader(tcp)
	req, err := http.ReadRequest(reader)

	if err != nil {
		logger.Printf("[ ERROR ] %v\n", err)
	} else {
		logger.Printf("[ INFO ] message read\n")
	}

	params, err := req.URL.Parse(req.RequestURI)

	if err != nil {
		logger.Printf("[ ERROR ] %v\n", err)
	}

	values := params.Query()
	file := values.Get(reqParam)

	if !values.Has(reqParam) {
		logger.Printf("[ ERROR ] %v parameter not found\n", reqParam)
		return
	} else {
		logger.Printf("[ INFO ] %v = %v\n", reqParam, file)
	}

	_, err = os.ReadFile(file)

	if err  != nil {
		logger.Printf("[ ERROR ] file opening error\n")
	}

	resp := http.Response{}

	fmt.Printf("resp: %v\n", resp)
}
