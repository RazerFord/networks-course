package common

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

const (
	PacketSize = 8   // 1024 bytes
	PacketLoss = 0.0 // 0 <= probability <= 1
	HeaderSize = 9   // 9 bytes
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
