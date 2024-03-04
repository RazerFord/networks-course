package main

import (
	"bufio"
	"bytes"
	"io"
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

	for {
		tcp, err := server.AcceptTCP()

		if err != nil {
			panic("connection error")
		}

		logger.Printf("[ INFO ] connection accepted\n")

		reader := bufio.NewReader(tcp)
		req, err := http.ReadRequest(reader)

		if err != nil {
			logger.Printf("[ ERROR ] %v\n", err)
			tcp.Close()
			continue
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

		if !values.Has(reqParam) {
			logger.Printf("[ ERROR ] %v parameter not found\n", reqParam)
			tcp.Close()
			continue
		} else {
			logger.Printf("[ INFO ] %v = %v\n", reqParam, file)
		}

		fbytes, err := os.ReadFile(file)

		resp := http.Response{
			Proto:      req.Proto,
			Request:    req,
			ProtoMajor: req.ProtoMajor,
			ProtoMinor: req.ProtoMinor,
			Header:     make(http.Header),
		}

		if os.IsNotExist(err) {
			logger.Printf("[ ERROR ] file not found\n")

			resp.Status = http.StatusText(http.StatusNotFound)
			resp.StatusCode = http.StatusNotFound
		} else if err != nil {
			logger.Printf("[ ERROR ] file opening error\n")

			resp.Status = http.StatusText(http.StatusNotFound)
			resp.StatusCode = http.StatusNotFound
		} else {
			resp.Body = io.NopCloser(bytes.NewBuffer(fbytes))
			resp.ContentLength = int64(len(fbytes))
			resp.Status = http.StatusText(http.StatusAccepted)
			resp.StatusCode = http.StatusAccepted
			resp.Body.Close()
		}

		var byteBuff bytes.Buffer
		w := bufio.NewWriter(&byteBuff)

		err = resp.Write(w)

		if err != nil {
			logger.Printf("[ ERROR ] error writing HTTP response. %v\n", err)
			tcp.Close()
			continue
		}

		w.Flush()
		tcp.Write(byteBuff.Bytes())

		logger.Printf("[ INFO ] response sent successfully: %v\n", resp.ContentLength)
		tcp.Close()
	}
}
