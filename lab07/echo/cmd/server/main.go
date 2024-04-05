package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"

	"echo/internal/app/logging"
)

const boundary = 0.2

type Server struct {
	conn *net.UDPConn
	rw   *bufio.ReadWriter
}

func NewServer(port int) *Server {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4bcast, Port: port})

	if err != nil {
		logging.Warn(err.Error())
		os.Exit(1)
	}

	logging.Info("server runs on address: %v:%v", net.IPv4bcast, port)
	return &Server{
		conn: conn,
		rw:   bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
	}
}

func (s *Server) Read() (string, error) {
	return s.rw.Reader.ReadString('\n')
}

func (s *Server) Write(msg string) (int, error) {
	n, err := s.rw.Writer.WriteString(msg)
	s.rw.Writer.Flush()
	return n, err
}

func (s *Server) Close() {
	s.conn.Close()
}

func main() {
	port := flag.Int("port", 8080, "server port")
	flag.Parse()

	s := NewServer(*port)
	defer s.Close()

	for {
		msg, err := s.Read()

		if err != nil {
			logging.Warn(err.Error())
		}

		logging.Info("received message: %s", msg)

		if rand.Float32() > boundary {
			fmt.Println("BLOCKING")
			_, err = s.Write(strings.ToUpper(msg))
			fmt.Println(strings.ToUpper(msg)+"\n")
			fmt.Println("UNBLOCKING")
			if err != nil {
				logging.Warn(err.Error())
			}
		}
	}
}
