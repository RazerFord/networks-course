package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"echo/internal/app/logging"
	"echo/internal/app/message"
)

const (
	boundary   = 0.2
	packetSize = 1024
)

var (
	clients map[string]*message.Message = make(map[string]*message.Message)
	mu                                  = sync.Mutex{}
)

func run(f func()) {
	mu.Lock()
	f()
	mu.Unlock()
}

type Client struct {
	Address  string
	Deadline time.Time
	*message.Message
}

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

	logging.Info("server runs on address: %v", conn.LocalAddr().String())
	return &Server{
		conn: conn,
	}
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

func startDispatcher(ch chan Client) {
	go func() {
		for {
			cl := <-ch
			time.Sleep(time.Until(cl.Deadline))

			run(func() {
				mnew, ok := clients[cl.Address]

				if ok && mnew.Seq <= cl.Seq {
					delete(clients, cl.Address)
					logging.Info("client %s disconnected", cl.Address)
				}
			})
		}
	}()
}

func main() {
	port := flag.Int("port", 8080, "server port")
	timeout := flag.Int("timeout", 2000, "timeout in milliseconds")
	flag.Parse()

	s := NewServer(*port)
	defer s.Close()

	ch := make(chan Client, 10)
	startDispatcher(ch)
	for {
		msg, addr, err := s.Read()

		if err != nil {
			logging.Warn(err.Error())
		}

		logging.Info("message received from %s: %s", addr.String(), msg)

		m := message.FromBytes([]byte(msg))
		run(func() {
			if mold, ok := clients[addr.String()]; ok {
				clients[addr.String()] = m
				if mold.Seq+1 != m.Seq {
					logging.Info("%s lost packets [%d, %d)\n", addr.String(), mold.Seq+1, m.Seq)
				}
			} else {
				clients[addr.String()] = m
			}
		})

		ch <- Client{Address: addr.String(), Deadline: time.Now().Add(time.Millisecond * time.Duration(*timeout)), Message: m}
	}
}
