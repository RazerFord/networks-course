package client

import (
	"fmt"
	"math/rand"
	"net"
	"stop-and-wait/internal/network/common"
	"time"
)

////////////////////////////// Client //////////////////////////////

type Client struct {
	udp     *net.UDPConn
	timeout time.Duration
	sender  *Sender
	reader  *Reader
}

func Connect(address string, port int, timeout time.Duration) (*Client, error) {
	raddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, err
	}

	saw := common.NewSAW()

	return &Client{
			udp:     conn,
			timeout: timeout,
			sender:  NewSender(conn, saw, timeout),
			reader:  NewReader(conn, saw),
		},
		nil
}

func (c *Client) Write(p []byte) (n int, err error) {
	for count, n1 := len(p), 0; count != 0; count = len(p) {
		count = min(count, common.PacketSize)
		packet := p[:count]

		n1, err = c.sender.Write(packet, 0)
		n += n1
		if err != nil {
			break
		}

		p = p[n1:]
	}
	// send fin byte
	c.sender.Write([]byte{}, 1)
	return n, err
}

func (s *Client) Read(p []byte) (n int, err error) {
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
	// read fin bit
	s.reader.Read(p[:])
	return n, err
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

func (s *Sender) Write(p []byte, fin byte) (int, error) {
	body := common.NewBody(nil, p[:], fin)
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

func (s *Sender) send(b []byte, _ net.Addr) (int, error) {
	if rand.Float32() < common.PacketLoss {
		return len(b), nil
	}
	return s.udp.Write(b)
}

////////////////////////////// Reader //////////////////////////////

type Reader struct {
	udp *net.UDPConn
	saw *common.SAW
}

func NewReader(udp *net.UDPConn, saw *common.SAW) *Reader {
	return &Reader{udp, saw}
}

func (r *Reader) Read(p []byte) (int, byte, error) {
	body := common.NewBody(nil, p[:], 0)
	r.saw.Put(common.Read, body, r.read, r.send)
	resp := r.saw.Get()
	return resp.Bytes, resp.Fin, resp.Err
}

func (r *Reader) read(b []byte) (int, net.Addr, error) {
	n, err := r.udp.Read(b)
	return n, nil, err
}
func (r *Reader) send(b []byte, _ net.Addr) (int, error) {
	if rand.Float32() < common.PacketLoss {
		return len(b), nil
	}
	return r.udp.Write(b)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
