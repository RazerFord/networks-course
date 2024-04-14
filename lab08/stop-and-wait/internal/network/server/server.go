package server

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"stop-and-wait/internal/network/common"
	"time"
)

////////////////////////////// Server //////////////////////////////

type Server struct {
	udp      *net.UDPConn
	duration time.Duration
	reader   *common.Reader
	sender   *Sender
}

func Connect(address string, port int, duration time.Duration) (*Server, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	return &Server{
			udp:      conn,
			duration: duration,
			reader:   common.NewReader(conn),
			sender:   NewSender(conn, duration),
		},
		nil
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
	fmt.Printf("[ INFO ] number of packets received %d\n", s.reader.CurSeqNum)
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
	fmt.Printf("[ INFO ] number of packets sent %d\n", s.sender.CurSeqNum)
	return n, err
}

type Sender struct {
	udp       *net.UDPConn
	timeout   time.Duration
	CurAckNum uint16
	CurSeqNum uint16
}

func NewSender(udp *net.UDPConn, d time.Duration) *Sender {
	return &Sender{udp, d, 0, 0}
}

func (s *Sender) Write(p []byte, fin byte, addr net.Addr) (int, error) {
	s.next()
	for {
		msg := common.NewMessage(s.CurAckNum, s.CurSeqNum, 0, fin, p)

		p1, err := common.ToBytes(msg)
		if err != nil {
			panic(err)
		}

		n, err := s.internalWrite(p1, addr)
		if errors.Is(err, common.ErrHeader) {
			continue
		}

		n1, err := s.internalRead(p1)
		if err != nil {
			if errors.Is(err, common.ErrHeader) {
				continue
			}
			s.udp.SetDeadline(time.Time{})
			return 0, err
		}

		m, err := common.FromBytes(p1[:n1])
		if err != nil {
			continue
		}

		if m.AckNum == msg.AckNum && m.SeqNum == s.CurSeqNum {
			fmt.Printf("[ INFO ] received Ack %d\n", msg.AckNum)
			s.udp.SetDeadline(time.Time{})
			return toRealS(n), nil
		}
		fmt.Printf("[ ERROR ] expected Ack %d, but actual Ack %d\n", msg.AckNum, s.CurAckNum)
	}
}

func (s *Sender) next() {
	s.CurAckNum = common.NextNum(s.CurAckNum)
	s.CurSeqNum++
}

func toRealS(s int) int {
	return max(common.ToRealSize(s), 0)
}

func (s *Sender) internalWrite(p []byte, addr net.Addr) (int, error) {
	if rand.Float32() < common.PacketLoss {
		fmt.Printf("[ INFO ] lost Ack %d\n", s.CurAckNum)
		return len(p), nil
	}
	s.udp.SetDeadline(time.Now().Add(s.timeout))
	n, err := s.udp.WriteTo(p, addr)
	if n < common.HeaderSize {
		return n, fmt.Errorf("%w: %w", common.ErrHeader, err)
	}
	fmt.Printf("[ INFO ] sent Ack %d\n", s.CurAckNum)
	return n, err
}

func (s *Sender) internalRead(p []byte) (int, error) {
	s.udp.SetDeadline(time.Now().Add(s.timeout))
	n, err := s.udp.Read(p)
	if errors.Is(err, os.ErrDeadlineExceeded) {
		fmt.Println("[ ERROR ] timeout")
		return n, fmt.Errorf("%w: %v", common.ErrHeader, err)
	}
	return n, err
}
