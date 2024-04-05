package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"

	"echo/internal/app/logging"
)

const (
	boundary   = 0.2
	packetSize = 1024
)

type Server struct {
	conn *net.UDPConn
}

func NewServer(port int) *Server {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))

	if err != nil {
		logging.Warn(err.Error())
		os.Exit(1)
	}

	conn, err := net.ListenUDP("udp", addr)

	if err != nil {
		logging.Warn(err.Error())
		os.Exit(1)
	}

	logging.Info("server runs on address: %v:%v", net.IPv4bcast, port)
	return &Server{conn: conn}
}

func (s *Server) Read() (string, net.Addr, error) {
	buff := make([]byte, packetSize)
	n, addr, err := (*s.conn).ReadFrom(buff)
	return string(buff[:n]), addr, err
}

func (s *Server) Write(addr net.Addr, msg string) (int, error) {
	return (*s.conn).WriteTo([]byte(msg), addr)
}

func (s *Server) Close() {
	(*s.conn).Close()
}

func main() {
	port := flag.Int("port", 8080, "server port")
	flag.Parse()

	s := NewServer(*port)
	defer s.Close()

	for {
		msg, addr, err := s.Read()

		if err != nil {
			logging.Warn(err.Error())
		}

		logging.Info("received message: %s", msg)

		if rand.Float32() > boundary {
			_, err = s.Write(addr, strings.ToUpper(msg))
			if err != nil {
				logging.Warn(err.Error())
			}
		}
	}
}
