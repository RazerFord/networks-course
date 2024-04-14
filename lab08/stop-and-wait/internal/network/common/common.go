package common

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"sync"
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

type Body struct {
	Addr    net.Addr
	Length  uint16 // 2 bytes
	Fin     byte   // 1 byte
	Payload []byte
}

func NewBody(a net.Addr, p []byte, fin byte) *Body {
	return &Body{
		Addr:    a,
		Length:  uint16(len(p)),
		Fin:     fin,
		Payload: p,
	}
}

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

type Command int

const (
	Send = Command(iota)
	Read
)

type entity struct {
	Command
	Body Body
	Read func([]byte) (int, net.Addr, error)
	Send func([]byte, net.Addr) (int, error)
}

func (e *entity) sendAck(ack, seq uint16, addr net.Addr) {
	fmt.Printf("[ INFO ] send Ack %d SeqNum %d\n", ack, seq)
	msg := NewMessage(ack, seq, 0, 0, []byte{})
	b, err := ToBytes(msg)
	exitIfNotNil(err)
	e.Send(b, addr)
}

type Response struct {
	Body  Body
	Addr  net.Addr
	Bytes int
	Fin   byte
	Err   error
}

type SAW struct {
	curAckNum uint16
	curSeqNum uint16
	mtx       sync.Mutex
	cache     sync.Map
	cmd       chan entity
	res       chan *Response
}

func NewSAW(read func([]byte) (int, net.Addr, error), send func([]byte, net.Addr) (int, error)) *SAW {
	saw := &SAW{0, 0, sync.Mutex{}, sync.Map{}, make(chan entity, 1), make(chan *Response, 1)}

	go func(saw *SAW) {
		for {
			saw.mtx.Lock()
			tmpBuff := make([]byte, HeaderSize+PacketSize)
			n, addr, _ := read(tmpBuff)
			if n == 0 {
				saw.mtx.Unlock()
				continue
			}
			msg, err := FromBytes(tmpBuff[:n])
			if err != nil {
				saw.mtx.Unlock()
				continue
			}

			if ack, ok := saw.cache.Load(msg.SeqNum); ok {
				fmt.Printf("[ INFO ] send Ack %d SeqNum %d\n", ack, msg.SeqNum)
				msg := NewMessage(ack.(uint16), msg.SeqNum, 0, 0, []byte{})
				b, err := ToBytes(msg)
				exitIfNotNil(err)
				send(b, addr)
			}
			saw.mtx.Unlock()
		}
	}(saw)
	go func(saw *SAW) {
		for {
			cmd := <-saw.cmd
			saw.mtx.Lock()
			var resp *Response
			switch cmd.Command {
			case Send:
				{
					for {
						// send message //
						msg := NewMessage(saw.curAckNum, saw.curSeqNum, 0, cmd.Body.Fin, cmd.Body.Payload)
						buff, err := ToBytes(msg)
						exitIfNotNil(err)

						_, err = cmd.Send(buff, cmd.Body.Addr)
						if err != nil {
							if errors.Is(err, os.ErrDeadlineExceeded) {
								continue
							}
							resp = &Response{cmd.Body, nil, 0, 0, err}
							break
						}

						// read acknowledge //
						ack := make([]byte, HeaderSize)
						_, _, err = cmd.Read(ack)
						if err != nil {
							if errors.Is(err, os.ErrDeadlineExceeded) {
								continue
							}
							resp = &Response{cmd.Body, nil, 0, 0, err}
							break
						}
						msg, err = FromBytes(ack)
						exitIfNotNil(err)
						fmt.Printf("[ INFO ] actual {Ack: %d, SeqNum: %d}; expected: {Ack: %d, SeqNum: %d}\n", msg.AckNum, msg.SeqNum, saw.curAckNum, saw.curSeqNum)

						// check ackNum and seqNum
						if msg.AckNum == saw.curAckNum && msg.SeqNum == saw.curSeqNum {
							saw.curAckNum = NextNum(saw.curAckNum)
							saw.curSeqNum = saw.curSeqNum + 1
							resp = &Response{cmd.Body, nil, int(cmd.Body.Length), cmd.Body.Fin, nil}
							break
						}
						if ack, ok := saw.cache.Load(msg.SeqNum); ok {
							cmd.sendAck(ack.(uint16), msg.SeqNum, cmd.Body.Addr)
						}
					}
				}
			case Read:
				{
					tmpBuff := make([]byte, HeaderSize+cmd.Body.Length)
					for {
						expAck := saw.curAckNum
						expSeq := saw.curSeqNum

						n, addr, err := cmd.Read(tmpBuff)
						if err != nil {
							if errors.Is(err, os.ErrDeadlineExceeded) {
								continue
							}
							resp = &Response{cmd.Body, nil, 0, 0, err}
							break
						}
						if n < HeaderSize {
							cmd.sendAck(expAck, expSeq, addr)
							continue
						}

						msg, err := FromBytes(tmpBuff[:n])
						exitIfNotNil(err)

						if expAck == msg.AckNum && expSeq == msg.SeqNum {
							cmd.sendAck(expAck, expSeq, addr)
							saw.cache.Store(expSeq, expAck)
							saw.curAckNum = NextNum(expAck)
							saw.curSeqNum = expSeq + 1
							n = int(msg.Length)
							for i := range n {
								cmd.Body.Payload[i] = msg.Payload[i]
							}
							resp = &Response{cmd.Body, addr, n, msg.Fin, nil}
							break
						}
						if ack, ok := saw.cache.Load(msg.SeqNum); ok {
							cmd.sendAck(ack.(uint16), msg.SeqNum, addr)
						} else {
							cmd.sendAck(expAck, expSeq, addr)
						}
					}
				}
			}
			saw.res <- resp
			saw.mtx.Unlock()
		}
	}(saw)
	return saw
}

func (saw *SAW) Put(
	cmd Command,
	body *Body,
	read func([]byte) (int, net.Addr, error),
	send func([]byte, net.Addr) (int, error),
) {
	saw.cmd <- entity{cmd, *body, read, send}
}

func (saw *SAW) Get() *Response {
	return <-saw.res
}

func exitIfNotNil(err error) {
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}

func ToChecksum(data []byte) uint16 {
	var cs uint16 = 0
	for len(data) != 0 {
		if len(data) > 1 {
			cs += binary.BigEndian.Uint16(data)
		} else {
			cs += uint16(data[0])
		}
	}
	return cs ^ math.MaxUint16
}

func CheckChecksum(data []byte, checksum uint16) bool {
	return (ToChecksum(data) | checksum) == math.MaxUint16
}
