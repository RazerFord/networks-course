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
	reader   *Reader
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
			reader:   NewReader(conn),
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

////////////////////////////// Reader //////////////////////////////

type Reader struct {
	udp       *net.UDPConn
	CurAckNum uint16
	CurSeqNum uint16
}

func NewReader(udp *net.UDPConn) *Reader {
	return &Reader{udp, 0, 0}
}

func (r *Reader) Read(p []byte) (int, byte, net.Addr, error) {
	tmpBuff := make([]byte, common.HeaderSize+common.PacketSize)
	for {
		n, addr, err := r.internalRead(tmpBuff)
		if err != nil {
			r.internalWriteAck(addr)
			continue
		}

		m, err := common.FromBytes(tmpBuff[:n])
		if err != nil {
			r.internalWriteAck(addr)
			continue
		}

		expAck := common.NextNum(r.CurAckNum)
		expSeq := r.CurSeqNum + 1
		if expAck == m.AckNum && expSeq == m.SeqNum {
			fmt.Printf("[ INFO ] received Ack %d SeqNum %d\n", m.AckNum, m.SeqNum)
			r.next()
			r.internalWriteAck(addr)
			n = int(m.Length)
			for i := range m.Length {
				p[i] = m.Payload[i]
			}
			return n, m.Fin, addr, nil
		}
		if expAck != m.AckNum {
			fmt.Printf("[ ERROR ] expected Ack %d, but actual Ack %d\n", expAck, m.AckNum)
		}
		if expSeq != m.SeqNum {
			fmt.Printf("[ ERROR ] expected SeqNum %d, but actual SeqNum %d\n", expSeq, m.SeqNum)
		}
		r.internalWriteAck(addr)
	}
}

func (r *Reader) next() {
	r.CurAckNum = common.NextNum(r.CurAckNum)
	r.CurSeqNum++
}

func (r *Reader) internalWriteAck(addr net.Addr) (int, error) {
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
	return r.udp.WriteTo(b, addr)
}

func (r *Reader) internalRead(p []byte) (int, net.Addr, error) {
	n, addr, err := r.udp.ReadFrom(p)
	if n < common.HeaderSize {
		return n, addr, fmt.Errorf("%w: %w", common.ErrHeader, err)
	}
	return n, addr, err
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

		if m.AckNum == s.CurAckNum && m.SeqNum == s.CurSeqNum {
			fmt.Printf("[ INFO ] received Ack %d SeqNum %d\n", m.AckNum, m.SeqNum)
			s.udp.SetDeadline(time.Time{})
			return toRealS(n), nil
		}
		if m.AckNum != s.CurAckNum {
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

func (s *Sender) internalWrite(p []byte, addr net.Addr) (int, error) {
	if rand.Float32() < common.PacketLoss {
		fmt.Printf("[ INFO ] lost Ack %d SeqNum %d\n", s.CurAckNum, s.CurSeqNum)
		return len(p), nil
	}
	s.udp.SetDeadline(time.Now().Add(s.timeout))
	n, err := s.udp.WriteTo(p, addr)
	if n < common.HeaderSize {
		return n, fmt.Errorf("%w: %w", common.ErrHeader, err)
	}
	fmt.Printf("[ INFO ] sent Ack %d SeqNum %d\n", s.CurAckNum, s.CurSeqNum)
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
