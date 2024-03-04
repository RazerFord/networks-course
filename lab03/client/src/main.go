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
	LOG      = false
	reqParam = "file"
	httpVer  = "HTTP/1.1"
)

func init() {
	if !LOG {
		logger.SetOutput(io.Discard)
	}
}

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

type client struct {
	conn *net.Conn
}

func NewClient(serverName string, port int) (*client, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%v:%v", serverName, port))

	if err != nil {
		return nil, err
	}

	logger.Println("[ INFO ] connection established")

	return &client{conn: &conn}, nil
}

func requestToBytes(req *http.Request) ([]byte, error) {
	var byteBuff bytes.Buffer
	w := bufio.NewWriter(&byteBuff)

	err := req.Write(w)
	w.Flush()

	return byteBuff.Bytes(), err
}

func (c *client) sendRequest(req *http.Request) error {
	buff, err := requestToBytes(req)

	if err != nil {
		logger.Printf("[ ERROR ] error converting requests to bytes: %v\n", err)
		return err
	}
	_, err = (*c.conn).Write(buff)

	if err != nil {
		logger.Printf("[ ERROR ] error writing HTTP response: %v\n", err)
		(*c.conn).Close()
		return err
	}
	logger.Println("[ INFO ] request has been sent.")

	return nil
}

func (c *client) readResponse(req *http.Request) ([]byte, error) {
	resp, err := http.ReadResponse(bufio.NewReader(*c.conn), req)
	defer func() { resp.Body.Close() }()

	if err != nil {
		logger.Printf("[ ERROR ] error reading HTTP request: %v\n", err)
		return nil, err
	}
	logger.Println("[ INFO ] response to request received.")

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Printf("[ ERROR ] body reading error: %v\n", err)
		return nil, err
	}
	return b, err
}

func (c *client) RequestFile(fileName string) ([]byte, error) {
	req := createRequest((*c.conn).RemoteAddr().Network(), fileName)

	if err := c.sendRequest(req); err != nil {
		logger.Printf("[ ERROR ] error sending request: %v\n", err)
		return nil, err
	}

	buff, err := c.readResponse(req)
	if err != nil {
		logger.Printf("[ ERROR ] error reading response: %v\n", err)
		return nil, err
	}

	return buff, nil
}

func (c *client) Close() {
	(*c.conn).Close()
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("pass server name, port, file name")
		os.Exit(1)
	}

	port, err := strconv.Atoi(os.Args[2])

	if err != nil {
		fmt.Println("port parsing error")
		os.Exit(1)
	}

	serverName := os.Args[1]
	fileName := os.Args[3]

	logger.Printf("[ INFO ] server name \"%v\" received\n", serverName)
	logger.Printf("[ INFO ] port \"%v\" received\n", port)
	logger.Printf("[ INFO ] file name \"%v\" received\n", fileName)

	c, err := NewClient(serverName, port)
	defer func() { c.Close() }()

	if err != nil {
		logger.Printf("[ ERROR ] client creation error: %V\n", err)
		return
	}

	buff, err := c.RequestFile(fileName)

	if err != nil {
		logger.Printf("[ ERROR ] file request error: %V\n", err)
		return
	}

	fmt.Println(string(buff))
}
