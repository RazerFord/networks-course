package client

import (
	"fmt"
	"math/rand"
	"net"
	"stop-and-wait/internal/network/common"
	"time"
)

////////////////////////////// Connection //////////////////////////////

type Connection struct {
	udp     *net.UDPConn
	timeout time.Duration
}

func Connect(address string, port int, timeout time.Duration) (*Connection, error) {
	raddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, err
	}

	return &Connection{conn, timeout}, nil
}

func (c *Connection) Write(p []byte) (n int, err error) {
	s := newSender(c)
	for count := len(p); count != 0; count = len(p) {
		count = min(count, common.PacketSize)
		packet := p[:count]

		n1, err := s.internalWrite(packet)
		if err != nil {
			return n, err
		}

		p = p[n1:]
	}
	return n, err
}

func (c *Connection) Read(p []byte) (int, error) {
	return 0, nil
}

////////////////////////////// sender //////////////////////////////

type sender struct {
	*Connection
	curAckNum uint16
	curSeqNum uint16
}

func newSender(c *Connection) *sender {
	return &sender{c, 0, 0}
}

func (s *sender) internalWrite(p []byte) (int, error) {
	msg := common.Message{
		s.curAckNum,
		s.curSeqNum,
		0,
		uint16(len(p)),
		p,
	}

	p, err := common.ToBytes(&msg)
	if err != nil {
		panic(err)
	}

	n, err := s.writeLoss(p)
	if err != nil {
		return n, err
	}

	for {
		s.udp.SetDeadline(time.Now().Add(s.timeout))
		n, err = s.udp.Read(p)

		if err == nil {
			m, err := common.FromBytes(p[:n])
			if err != nil {
				panic(err)
			}
			if m.AckNum == msg.AckNum+1 && m.SeqNum == msg.SeqNum+1 {
				break
			}
		}
	}
	msg.AckNum = common.NextNum(msg.AckNum)
	s.curSeqNum++

	return n, nil
}

func (s *sender) writeLoss(p []byte) (int, error) {
	if rand.Float32() < common.PacketLoss {
		return len(p), nil
	}
	return s.udp.Write(p)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
