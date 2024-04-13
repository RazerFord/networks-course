package server

import (
	"fmt"
	"net"
	"stop-and-wait/internal/network/common"
)

////////////////////////////// Server //////////////////////////////

type Server struct {
	udp    *net.UDPConn
	reader *common.Reader
}

func Connect(address string, port int) (*Server, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	return &Server{udp: conn, reader: common.NewReader(conn)}, nil
}

func (s *Server) Read(p []byte) (n int, err error) {
	for len(p) != 0 {
		var n1 int
		var fin byte
		n1, fin, err = s.reader.Read(p[:])
		n += n1
		if err != nil || fin == 1 {
			return n, err
		}
		p = p[n1:]
	}
	// try to read fin bit
	s.reader.Read(p[:])
	fmt.Printf("[ INFO ] number of packets received %d\n", s.reader.CurSeqNum)
	return n, err
}
