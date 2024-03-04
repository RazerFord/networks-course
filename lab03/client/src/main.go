package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

var logger = log.Default()

const reqParam = "file"

func createURL(host, fileName string) *url.URL {
	urlAddr := url.URL{}
	urlAddr.Host = host
	values := url.Values{}
	values.Add(reqParam, fileName)
	urlAddr.RawQuery = values.Encode()
	return &urlAddr
}

func main() {
	if len(os.Args) < 4 {
		panic("pass server name, port, file name")
	}

	port, err := strconv.Atoi(os.Args[2])

	if err != nil {
		panic("port parsing error")
	}

	serverName := os.Args[1]
	fileName := os.Args[3]

	logger.Printf("[ INFO ] server name %v received\n", serverName)
	logger.Printf("[ INFO ] port %v received\n", port)
	logger.Printf("[ INFO ] file name %v received\n", fileName)

	conn, err := net.Dial("tcp", fmt.Sprintf("%v:%v", serverName, port))

	if err != nil {
		panic("connection error")
	}

	logger.Println("[ INFO ] connection established")

	urlAddr := createURL(conn.RemoteAddr().Network(), fileName)

	resp := http.Request{
		Proto:      "HTTP/1.1",
		Host:       conn.RemoteAddr().Network(),
		Method:     http.MethodGet,
		RequestURI: urlAddr.RequestURI(),
		Header:     make(http.Header),
		URL:        urlAddr,
		Body:       nil,
	}

	var byteBuff bytes.Buffer
	w := bufio.NewWriter(&byteBuff)

	err = resp.Write(w)
	w.Flush()

	if err != nil {
		logger.Printf("[ ERROR ] error writing HTTP response. %v\n", err)
		conn.Close()
		return
	}
	conn.Write(byteBuff.Bytes())

	logger.Println("[ INFO ] request has been sent.")

	conn.Close()
}
