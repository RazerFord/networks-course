package client

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"stop-and-wait/internal/network/common"
	"time"
)

////////////////////////////// Client //////////////////////////////

type Client struct {
	udp       *net.UDPConn
	timeout   time.Duration
	curAckNum uint16
	curSeqNum uint16
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

	return &Client{conn, timeout, 0, 0}, nil
}

func (c *Client) Write(p []byte) (n int, err error) {
	s := newSender(c)
	for count, n1 := len(p), 0; count != 0; count = len(p) {
		count = min(count, common.PacketSize)
		packet := p[:count]

		n1, err = s.write(packet, 0)
		n += n1
		if err != nil {
			break
		}

		p = p[n1:]
	}
	// send fin byte
	s.write([]byte{}, 1)

	return n, err
}

////////////////////////////// sender //////////////////////////////

type sender struct {
	*Client
}

func newSender(c *Client) *sender {
	return &sender{c}
}

func (s *sender) write(p []byte, fin byte) (int, error) {
	s.next()
	for {
		msg := common.NewMessage(s.curAckNum, s.curSeqNum, 0, fin, p)

		p1, err := common.ToBytes(msg)
		if err != nil {
			panic(err)
		}

		n, err := s.internalWrite(p1)
		if errors.Is(err, common.ErrHeader) {
			continue
		}

		n1, err := s.internalRead(p1)
		if err != nil {
			if errors.Is(err, common.ErrHeader) {
				continue
			}
			return 0, err
		}

		m, err := common.FromBytes(p1[:n1])
		if err != nil {
			continue
		}

		if m.AckNum == msg.AckNum && m.SeqNum == s.curSeqNum {
			return toRealS(n), nil
		}
	}
}

func (s *sender) next() {
	s.curAckNum = common.NextNum(s.curAckNum)
	s.curSeqNum++
}

func toRealS(s int) int {
	return max(common.ToRealSize(s), 0)
}

func (s *sender) internalWrite(p []byte) (int, error) {
	if rand.Float32() < common.PacketLoss {
		return len(p), nil
	}
	s.udp.SetDeadline(time.Now().Add(s.timeout))
	n, err := s.udp.Write(p)
	if n < common.HeaderSize {
		return n, fmt.Errorf("%w: %w", common.ErrHeader, err)
	}
	return n, err
}

func (s *sender) internalRead(p []byte) (int, error) {
	s.udp.SetDeadline(time.Now().Add(s.timeout))
	n, err := s.udp.Read(p)
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return n, fmt.Errorf("%w: %v", common.ErrHeader, err)
	}
	return n, err
}

////////////////////////////// sender //////////////////////////////

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
