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
	sender  *common.Sender
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
			sender:  common.NewSender(conn, timeout),
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

////////////////////////////// reader //////////////////////////////

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

		if common.NextNum(r.CurAckNum) == m.AckNum && r.CurSeqNum+1 == m.SeqNum {
			fmt.Printf("[ INFO ] received Ack %d\n", m.AckNum)
			r.next()
			r.internalWriteAck()
			n = int(m.Length)
			for i := range m.Length {
				p[i] = m.Payload[i]
			}
			return n, m.Fin, nil
		}

		r.internalWriteAck()

		fmt.Printf("[ ERROR ] expected Ack %d, but actual Ack %d\n", common.NextNum(r.CurAckNum), m.AckNum)
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
		fmt.Printf("[ INFO ] lost Ack %d\n", r.CurAckNum)
		return len(b), nil
	}
	fmt.Printf("[ INFO ] sent Ack %d\n", r.CurAckNum)
	return r.udp.Write(b)
}

func (r *Reader) internalRead(p []byte) (int, error) {
	fmt.Println(r.udp.LocalAddr())
	fmt.Println(r.udp.RemoteAddr())
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
