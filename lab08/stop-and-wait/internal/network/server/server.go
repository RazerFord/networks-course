package server

import (
	"fmt"
	"math/rand"
	"net"
	"stop-and-wait/internal/network/common"
	"time"
)

////////////////////////////// Server //////////////////////////////

type Server struct {
	udp      *net.UDPConn
	timeout time.Duration
	reader   *Reader
	sender   *Sender
}

func Connect(address string, port int, timeout time.Duration) (*Server, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	s := &Server{
		udp:      conn,
		timeout: timeout,
	}
	saw := common.NewSAW(s.read, s.send)

	s.reader = NewReader(conn, saw)
	s.sender = NewSender(conn, saw, timeout)

	return s, nil
}

func (s *Server) read(b []byte) (int, net.Addr, error) {
	s.udp.SetReadDeadline(time.Now().Add(s.timeout))
	n, addr, err := s.udp.ReadFrom(b)
	s.udp.SetReadDeadline(time.Time{})
	return n, addr, err
}

func (s *Server) send(b []byte, addr net.Addr) (int, error) {
	if rand.Float32() < common.PacketLoss {
		return len(b), nil
	}
	return s.udp.WriteTo(b, addr)
}

func (s *Server) Read(p []byte) (n int, addr net.Addr, err error) {
	for len(p) != 0 {
		var n1 int
		var fin byte
		n1, fin, addr, err = s.reader.Read(p[:])
		n += n1
		if err != nil || fin == 1 {
			return n, addr, err
		}
		p = p[n1:]
	}
	// read fin bit
	s.reader.Read(p[:])
	return n, addr, err
}

func (s *Server) Write(p []byte, addr net.Addr) (n int, err error) {
	for count, n1 := len(p), 0; count != 0; count = len(p) {
		count = min(count, common.PacketSize)
		packet := p[:count]

		n1, err = s.sender.Write(packet, 0, addr)
		n += n1
		if err != nil {
			break
		}

		p = p[n1:]
	}
	// send fin byte
	s.sender.Write([]byte{}, 1, addr)
	return n, err
}

////////////////////////////// Reader //////////////////////////////

type Reader struct {
	udp *net.UDPConn
	saw *common.SAW
}

func NewReader(udp *net.UDPConn, saw *common.SAW) *Reader {
	return &Reader{udp, saw}
}

func (r *Reader) Read(p []byte) (int, byte, net.Addr, error) {
	body := common.NewBody(nil, p[:], 0)
	r.saw.Put(common.Read, body, r.read, r.send)
	resp := r.saw.Get()
	return resp.Bytes, resp.Fin, resp.Addr, resp.Err
}

func (r *Reader) read(b []byte) (int, net.Addr, error) {
	return r.udp.ReadFrom(b)
}

func (r *Reader) send(b []byte, addr net.Addr) (int, error) {
	if rand.Float32() < common.PacketLoss {
		return len(b), nil
	}
	return r.udp.WriteTo(b, addr)
}

////////////////////////////// Sender //////////////////////////////

type Sender struct {
	udp     *net.UDPConn
	saw     *common.SAW
	timeout time.Duration
}

func NewSender(udp *net.UDPConn, saw *common.SAW, d time.Duration) *Sender {
	return &Sender{udp, saw, d}
}

func (s *Sender) Write(p []byte, fin byte, addr net.Addr) (int, error) {
	body := common.NewBody(addr, p, fin)
	s.saw.Put(common.Send, body, s.read, s.send)
	resp := s.saw.Get()
	return resp.Bytes, resp.Err
}

func (s *Sender) read(b []byte) (int, net.Addr, error) {
	s.udp.SetReadDeadline(time.Now().Add(s.timeout))
	n, addr, err := s.udp.ReadFrom(b)
	s.udp.SetReadDeadline(time.Time{})
	return n, addr, err
}

func (s *Sender) send(b []byte, addr net.Addr) (int, error) {
	if rand.Float32() < common.PacketLoss {
		return len(b), nil
	}
	return s.udp.WriteTo(b, addr)
}
