package server

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"stop-and-wait/internal/network/common"
	"time"
)

////////////////////////////// Server //////////////////////////////

type Server struct {
	udp       *net.UDPConn
	timeout   time.Duration
	curAckNum uint16
	curSeqNum uint16
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

	return &Server{conn, timeout, 0, 0}, nil
}

func (s *Server) Read(p []byte) (n int, err error) {
	r := newReader(s)
	for len(p) != 0 {
		n1, err := r.read(p)
		n += n1
		if err != nil {
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

func (r *reader) read(p []byte) (int, error) {
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
			io.CopyN(bytes.NewBuffer(p), bytes.NewBuffer(m.Payload), int64(n))
			return common.ToRealSize(n), nil
		}

		r.internalWriteAck(addr)
	}
}

func (r *reader) next() {
	r.curAckNum = common.NextNum(r.curAckNum)
	r.curSeqNum++
}

func (r *reader) internalWriteAck(addr net.Addr) (int, error) {
	msg := common.NewMessage(r.curAckNum, r.curSeqNum, 0, []byte{})
	b, err := common.ToBytes(msg)
	if err != nil {
		panic(err)
	}
	return r.udp.WriteTo(b, addr)
}

func (r *reader) internalRead(p []byte) (int, net.Addr, error) {
	r.udp.SetDeadline(time.Now().Add(r.timeout))
	n, addr, err := r.udp.ReadFrom(p)
	if n <= common.HeaderSize {
		return n, addr, fmt.Errorf("%w: %w", common.ErrHeader, err)
	}
	return n, addr, err
}
