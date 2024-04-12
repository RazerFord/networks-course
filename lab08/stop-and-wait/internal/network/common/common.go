package common

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
)

const (
	PacketSize = 1024 // 1024 bytes
	PacketLoss = 0.3  // 0 <= probability <= 1
)

var (
	errPacketNum = errors.New("prev packet number must be 0 or 1")
)

type Message struct {
	AckNum   uint16 // 2 byte
	SeqNum   uint16 // 2 byte
	Checksum uint16 // 2 byte
	Length   uint16 // 2 byte
	Payload  []byte
}

func ToBytes(m *Message) ([]byte, error) {
	buff := bytes.Buffer{}
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(m)
	if err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func FromBytes(p []byte) (*Message, error) {
	enc := gob.NewDecoder(bytes.NewReader(p))
	m := Message{}
	err := enc.Decode(&m)
	if err != nil {
		return nil, err
	}
	return &m, nil
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
