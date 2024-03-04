package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"syscall"
)

var (
	logger       = log.Default()
	errFileParam = errors.New("file parameter not passed")
)

const (
	DEBUG      = false // set true to enable logging
	reqParam   = "file"
	fileSystem = "filesystem"
)

func init() {
	if !DEBUG {
		logger.SetOutput(io.Discard)
	}
}

func getPath(p string) string {
	return path.Join(".", fileSystem, p)
}

func handleRequest(tcp *net.TCPConn) {
	tid := syscall.Gettid()

	logger.Printf("[ INFO ] thread id: %v\n", tid)

	reader := bufio.NewReader(tcp)
	req, err := http.ReadRequest(reader)

	if err != nil {
		logger.Printf("[ ERROR ] %v\n", err)
		tcp.Close()
		return
	} else {
		logger.Printf("[ INFO ] message read\n")
	}
	req.Body.Close()

	params, err := req.URL.Parse(req.RequestURI)

	if err != nil {
		logger.Printf("[ ERROR ] %v\n", err)
	}

	values := params.Query()
	file := values.Get(reqParam)

	resp := http.Response{
		Proto:      req.Proto,
		Request:    req,
		ProtoMajor: req.ProtoMajor,
		ProtoMinor: req.ProtoMinor,
		Header:     make(http.Header),
	}

	fbytes := make([]byte, 0)
	if !values.Has(reqParam) {
		logger.Printf("[ ERROR ] %v parameter not found: %v\n", reqParam, err)
		err = errFileParam
	} else {
		logger.Printf("[ INFO ] %v = %v\n", reqParam, file)
		fbytes, err = os.ReadFile(getPath(file))
	}

	if os.IsNotExist(err) {
		logger.Printf("[ ERROR ] file not found: %v\n", err)

		resp.Status = http.StatusText(http.StatusNotFound)
		resp.StatusCode = http.StatusNotFound
	} else if err != nil {
		logger.Printf("[ ERROR ] file opening error: %v\n", err)

		resp.Status = http.StatusText(http.StatusNotFound)
		resp.StatusCode = http.StatusNotFound
	} else {
		resp.Body = io.NopCloser(bytes.NewBuffer(fbytes))
		resp.ContentLength = int64(len(fbytes))
		resp.Status = http.StatusText(http.StatusOK)
		resp.StatusCode = http.StatusOK
		resp.Body.Close()
	}

	var byteBuff bytes.Buffer
	w := bufio.NewWriter(&byteBuff)

	err = resp.Write(w)

	if err != nil {
		logger.Printf("[ ERROR ] error writing HTTP response: %v\n", err)
		tcp.Close()
		return
	}

	w.Flush()
	tcp.Write(byteBuff.Bytes())

	logger.Printf("[ INFO ] response sent successfully: %v\n", resp.ContentLength)
	tcp.Close()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("pass the port")
		os.Exit(1)
	}

	port, err := strconv.Atoi(os.Args[1])

	if err != nil {
		fmt.Println("port parsing error")
		os.Exit(1)
	}

	logger.Printf("[ INFO ] port %d received\n", port)

	addr := net.TCPAddr{}
	addr.IP = net.ParseIP("127.0.0.1")
	addr.Port = port

	server, err := net.ListenTCP("tcp", &addr)

	if err != nil {
		fmt.Println("listening error")
		os.Exit(1)
	}

	logger.Printf("[ INFO ] server created: http://%v:%v \n", addr.IP, addr.Port)

	for {
		tcp, err := server.AcceptTCP()

		if err != nil {
			logger.Printf("[ INFO ] accepting error: %v\n", err)
			continue
		}

		logger.Printf("[ INFO ] connection accepted\n")

		go handleRequest(tcp)
	}
}
