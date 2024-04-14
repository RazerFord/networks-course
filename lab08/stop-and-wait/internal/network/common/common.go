package common

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
)

const (
	PacketSize = 1024 // 1024 bytes
	PacketLoss = 0.3  // 0 <= probability <= 1
	HeaderSize = 9    // 9 bytes
)

var (
	ErrHeader    = errors.New("error header")
	errPacketNum = errors.New("prev packet number must be 0 or 1")
)

type Message struct {
	AckNum   uint16 // 2 bytes
	SeqNum   uint16 // 2 bytes
	Checksum uint16 // 2 bytes
	Length   uint16 // 2 bytes
	Fin      byte   // 1 byte
	Payload  []byte
}

func ToRealSize(s int) int {
	return s - HeaderSize
}

func NewMessage(a, s, c uint16, f byte, p []byte) *Message {
	return &Message{
		AckNum:   a,
		SeqNum:   s,
		Checksum: c,
		Length:   uint16(len(p)),
		Fin:      f,
		Payload:  p,
	}
}

func ToBytes(m *Message) ([]byte, error) {
	buff := []byte{}
	buff = binary.BigEndian.AppendUint16(buff, m.AckNum)
	buff = binary.BigEndian.AppendUint16(buff, m.SeqNum)
	buff = binary.BigEndian.AppendUint16(buff, m.Checksum)
	buff = binary.BigEndian.AppendUint16(buff, m.Length)
	buff = append(buff, m.Fin)
	buff = append(buff, m.Payload...)
	return buff, nil
}

func FromBytes(p []byte) (*Message, error) {
	if len(p) < HeaderSize {
		return nil, ErrHeader
	}
	ackNum := binary.BigEndian.Uint16(p)
	p = p[2:]
	seqNum := binary.BigEndian.Uint16(p)
	p = p[2:]
	checksum := binary.BigEndian.Uint16(p)
	p = p[2:]
	length := binary.BigEndian.Uint16(p)
	p = p[2:]
	fin := p[0]
	p = p[1:]
	return &Message{
			AckNum:   ackNum,
			SeqNum:   seqNum,
			Checksum: checksum,
			Length:   length,
			Fin:      fin,
			Payload:  p,
		},
		nil
}

type number interface {
	uint16
}

func NextNum[T number](i T) T {
	Require(i == 0 || i == 1, errPacketNum)
	return 1 - i
}

func Require(c bool, err error) {
	if !c {
		fmt.Println(err)
		os.Exit(1)
	}
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
		msg := NewMessage(s.CurAckNum, s.CurSeqNum, 0, fin, p)

		p1, err := ToBytes(msg)
		if err != nil {
			panic(err)
		}

		n, err := s.internalWrite(p1)
		if errors.Is(err, ErrHeader) {
			continue
		}

		n1, err := s.internalRead(p1)
		if err != nil {
			if errors.Is(err, ErrHeader) {
				continue
			}
			s.udp.SetDeadline(time.Time{})
			return 0, err
		}

		m, err := FromBytes(p1[:n1])
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
	s.CurAckNum = NextNum(s.CurAckNum)
	s.CurSeqNum++
}

func toRealS(s int) int {
	return max(ToRealSize(s), 0)
}

func (s *Sender) internalWrite(p []byte) (int, error) {
	if rand.Float32() < PacketLoss {
		fmt.Printf("[ INFO ] lost Ack %d\n", s.CurAckNum)
		return len(p), nil
	}
	s.udp.SetWriteDeadline(time.Now().Add(s.timeout))
	n, err := s.udp.Write(p)
	if n < HeaderSize {
		return n, fmt.Errorf("%w: %w", ErrHeader, err)
	}
	fmt.Printf("[ INFO ] sent Ack %d\n", s.CurAckNum)
	return n, err
}

func (s *Sender) internalRead(p []byte) (int, error) {
	s.udp.SetReadDeadline(time.Now().Add(s.timeout))
	n, err := s.udp.Read(p)
	if errors.Is(err, os.ErrDeadlineExceeded) {
		fmt.Println("[ ERROR ] timeout")
		return n, fmt.Errorf("%w: %v", ErrHeader, err)
	}
	return n, err
}

////////////////////////////// reader //////////////////////////////

type Reader struct {
	udp       *net.UDPConn
	CurAckNum uint16
	CurSeqNum uint16
}

func NewReader(udp *net.UDPConn) *Reader {
	return &Reader{udp, 0, 0}
}

func (r *Reader) Read(p []byte) (int, byte, net.Addr, error) {
	tmpBuff := make([]byte, HeaderSize+PacketSize)
	for {
		n, addr, err := r.internalRead(tmpBuff)
		if err != nil {
			r.internalWriteAck(addr)
			continue
		}

		m, err := FromBytes(tmpBuff[:n])
		if err != nil {
			r.internalWriteAck(addr)
			continue
		}

		if NextNum(r.CurAckNum) == m.AckNum && r.CurSeqNum+1 == m.SeqNum {
			fmt.Printf("[ INFO ] received Ack %d\n", m.AckNum)
			r.next()
			r.internalWriteAck(addr)
			n = int(m.Length)
			for i := range m.Length {
				p[i] = m.Payload[i]
			}
			return n, m.Fin, addr, nil
		}

		r.internalWriteAck(addr)

		fmt.Printf("[ ERROR ] expected Ack %d, but actual Ack %d\n", NextNum(r.CurAckNum), m.AckNum)
	}
}

func (r *Reader) next() {
	r.CurAckNum = NextNum(r.CurAckNum)
	r.CurSeqNum++
}

func (r *Reader) internalWriteAck(addr net.Addr) (int, error) {
	msg := NewMessage(r.CurAckNum, r.CurSeqNum, 0, 0, []byte{})
	b, err := ToBytes(msg)
	if err != nil {
		panic(err)
	}
	if rand.Float32() < PacketLoss {
		fmt.Printf("[ INFO ] lost Ack %d\n", r.CurAckNum)
		return len(b), nil
	}
	fmt.Printf("[ INFO ] sent Ack %d\n", r.CurAckNum)
	return r.udp.WriteTo(b, addr)
}

func (r *Reader) internalRead(p []byte) (int, net.Addr, error) {
	n, addr, err := r.udp.ReadFrom(p)
	if n < HeaderSize {
		return n, addr, fmt.Errorf("%w: %w", ErrHeader, err)
	}
	return n, addr, err
}
