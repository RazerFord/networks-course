package message

import (
	"encoding/json"
	"time"
)

type Message struct {
	Seq  int       `json:"seq"`
	Time time.Time `json:"time"`
}

func NewMessage(seq int, time time.Time) *Message {
	return &Message{
		Seq:  seq,
		Time: time,
	}
}

func FromBytes(bs []byte) *Message {
	m := &Message{}
	json.Unmarshal(bs, m)
	return m
}

func ToBytes(m *Message) ([]byte, error) {
	return json.Marshal(m)
}
