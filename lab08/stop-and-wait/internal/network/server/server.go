package server

import (
	"fmt"
	"math/rand"
	"net"
	"stop-and-wait/internal/network/common"
)

////////////////////////////// Server //////////////////////////////

type Server struct {
	udp       *net.UDPConn
	curAckNum uint16
	curSeqNum uint16
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

	return &Server{conn, 0, 0}, nil
}

func (s *Server) Read(p []byte) (n int, err error) {
	r := newReader(s)
	for len(p) != 0 {
		n1, fin, err := r.read(p[:])
		n += n1
		if err != nil || fin == 1 {
			return n, err
		}
		p = p[n1:]
	}
	return n, err
}

////////////////////////////// reader //////////////////////////////

type reader struct {
	*Server
}

func newReader(s *Server) *reader {
	return &reader{s}
}

func (r *reader) read(p []byte) (int, byte, error) {
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

		if common.NextNum(r.curAckNum) == m.AckNum && r.curSeqNum+1 == m.SeqNum {
			r.next()
			r.internalWriteAck(addr)
			n = int(m.Length)
			for i := range m.Length {
				p[i] = m.Payload[i]
			}
			return n, m.Fin, nil
		}

		r.internalWriteAck(addr)
	}
}

func (r *reader) next() {
	r.curAckNum = common.NextNum(r.curAckNum)
	r.curSeqNum++
}

func (r *reader) internalWriteAck(addr net.Addr) (int, error) {
	msg := common.NewMessage(r.curAckNum, r.curSeqNum, 0, 0, []byte{})
	b, err := common.ToBytes(msg)
	if err != nil {
		panic(err)
	}
	if rand.Float32() < common.PacketLoss {
		return len(b), nil
	}
	return r.udp.WriteTo(b, addr)
}

func (r *reader) internalRead(p []byte) (int, net.Addr, error) {
	n, addr, err := r.udp.ReadFrom(p)
	if n < common.HeaderSize {
		return n, addr, fmt.Errorf("%w: %w", common.ErrHeader, err)
	}
	return n, addr, err
}
