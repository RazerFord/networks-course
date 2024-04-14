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

	return &Client{
			udp:     conn,
			timeout: timeout,
			sender:  NewSender(conn, timeout),
			reader:  NewReader(conn),
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
	fmt.Printf("[ INFO ] number of packets sent %d\n", c.sender.CurSeqNum)
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
	fmt.Printf("[ INFO ] number of packets received %d\n", s.reader.CurSeqNum)
	return n, err
}

////////////////////////////// Sender //////////////////////////////

type Sender struct {
	udp       *net.UDPConn
	timeout   time.Duration
	CurAckNum uint16
	CurSeqNum uint16
}

func NewSender(udp *net.UDPConn, d time.Duration) *Sender {
	return &Sender{udp, d, 0, 0}
}

func (s *Sender) Write(p []byte, fin byte) (int, error) {
	s.next()
	for {
		msg := common.NewMessage(s.CurAckNum, s.CurSeqNum, 0, fin, p)

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
			s.udp.SetDeadline(time.Time{})
			return 0, err
		}

		m, err := common.FromBytes(p1[:n1])
		if err != nil {
			continue
		}

		if m.AckNum == msg.AckNum && m.SeqNum == s.CurSeqNum {
			fmt.Printf("[ INFO ] received Ack %d SeqNum %d\n", msg.AckNum, msg.SeqNum)
			s.udp.SetDeadline(time.Time{})
			return toRealS(n), nil
		}

		if s.CurAckNum != m.AckNum {
			fmt.Printf("[ ERROR ] expected Ack %d, but actual Ack %d\n", s.CurAckNum, m.AckNum)
		}
		if s.CurSeqNum != m.SeqNum {
			fmt.Printf("[ ERROR ] expected SeqNum %d, but actual SeqNum %d\n", s.CurSeqNum, m.SeqNum)
		}
	}
}

func (s *Sender) next() {
	s.CurAckNum = common.NextNum(s.CurAckNum)
	s.CurSeqNum++
}

func toRealS(s int) int {
	return max(common.ToRealSize(s), 0)
}

func (s *Sender) internalWrite(p []byte) (int, error) {
	if rand.Float32() < common.PacketLoss {
		fmt.Printf("[ INFO ] lost Ack %d SeqNum %d\n", s.CurAckNum, s.CurSeqNum)
		return len(p), nil
	}
	s.udp.SetWriteDeadline(time.Now().Add(s.timeout))
	n, err := s.udp.Write(p)
	if n < common.HeaderSize {
		return n, fmt.Errorf("%w: %w", common.ErrHeader, err)
	}
	fmt.Printf("[ INFO ] sent Ack %d SeqNum %d\n", s.CurAckNum, s.CurSeqNum)
	return n, err
}

func (s *Sender) internalRead(p []byte) (int, error) {
	s.udp.SetReadDeadline(time.Now().Add(s.timeout))
	n, err := s.udp.Read(p)
	if errors.Is(err, os.ErrDeadlineExceeded) {
		fmt.Println("[ ERROR ] timeout")
		return n, fmt.Errorf("%w: %v", common.ErrHeader, err)
	}
	return n, err
}

////////////////////////////// Reader //////////////////////////////

type Reader struct {
	udp       *net.UDPConn
	CurAckNum uint16
	CurSeqNum uint16
}

func NewReader(udp *net.UDPConn) *Reader {
	return &Reader{udp, 0, 0}
}

func (r *Reader) Read(p []byte) (int, byte, error) {
	tmpBuff := make([]byte, common.HeaderSize+common.PacketSize)
	for {
		n, err := r.internalRead(tmpBuff)
		if err != nil {
			r.internalWriteAck()
			continue
		}

		m, err := common.FromBytes(tmpBuff[:n])
		if err != nil {
			r.internalWriteAck()
			continue
		}

		expAck := common.NextNum(r.CurAckNum)
		expSeq := r.CurSeqNum + 1
		if expAck == m.AckNum && expSeq == m.SeqNum {
			fmt.Printf("[ INFO ] received Ack %d SeqNum %d\n", m.AckNum, m.SeqNum)
			r.next()
			r.internalWriteAck()
			n = int(m.Length)
			for i := range m.Length {
				p[i] = m.Payload[i]
			}
			return n, m.Fin, nil
		}
		if expAck != m.AckNum {
			fmt.Printf("[ ERROR ] expected Ack %d, but actual Ack %d\n", expAck, m.AckNum)
		}
		if expSeq != m.SeqNum {
			fmt.Printf("[ ERROR ] expected SeqNum %d, but actual SeqNum %d\n", expSeq, m.SeqNum)
		}
		r.internalWriteAck()
	}
}

func (r *Reader) next() {
	r.CurAckNum = common.NextNum(r.CurAckNum)
	r.CurSeqNum++
}

func (r *Reader) internalWriteAck() (int, error) {
	msg := common.NewMessage(r.CurAckNum, r.CurSeqNum, 0, 0, []byte{})
	b, err := common.ToBytes(msg)
	if err != nil {
		panic(err)
	}
	if rand.Float32() < common.PacketLoss {
		fmt.Printf("[ INFO ] lost Ack %d SeqNum %d\n", r.CurAckNum, r.CurSeqNum)
		return len(b), nil
	}
	fmt.Printf("[ INFO ] sent Ack %d SeqNum %d\n", r.CurAckNum, r.CurSeqNum)
	return r.udp.Write(b)
}

func (r *Reader) internalRead(p []byte) (int, error) {
	n, err := r.udp.Read(p)
	if n < common.HeaderSize {
		return n, fmt.Errorf("%w: %w", common.ErrHeader, err)
	}
	return n, err
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
