package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

var logger = log.Default()

const (
	reqParam = "file"
	httpVer  = "HTTP/1.1"
)

func createURL(host, fileName string) *url.URL {
	urlAddr := url.URL{}
	urlAddr.Host = host
	values := url.Values{}
	values.Add(reqParam, fileName)
	urlAddr.RawQuery = values.Encode()
	return &urlAddr
}

func createRequest(host, fileName string) *http.Request {
	urlAddr := createURL(host, fileName)

	return &http.Request{
		Proto:      httpVer,
		Host:       host,
		Method:     http.MethodGet,
		RequestURI: urlAddr.RequestURI(),
		Header:     make(http.Header),
		URL:        urlAddr,
		Body:       nil,
	}
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

	req := createRequest(conn.RemoteAddr().Network(), fileName)

	var byteBuff bytes.Buffer
	w := bufio.NewWriter(&byteBuff)

	err = req.Write(w)
	w.Flush()
	conn.Write(byteBuff.Bytes())

	if err != nil {
		logger.Printf("[ ERROR ] error writing HTTP response. %v\n", err)
		conn.Close()
		return
	}
	logger.Println("[ INFO ] request has been sent.")

	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	defer func() { resp.Body.Close() }()

	if err != nil {
		logger.Printf("[ ERROR ] error reading HTTP request: %v\n", err)
		return
	}
	logger.Println("[ INFO ] response to request received.")

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Printf("[ ERROR ] body reading error: %v\n", err)
	}

	fmt.Println(string(b))
	conn.Close()
}
