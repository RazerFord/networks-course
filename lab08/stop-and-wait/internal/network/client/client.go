package client

import (
	"fmt"
	"net"
	"stop-and-wait/internal/network/common"
	"time"
)

////////////////////////////// Client //////////////////////////////

type Client struct {
	udp     *net.UDPConn
	timeout time.Duration
	sender  *common.Sender
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

	return &Client{conn, timeout, common.NewSender(conn, timeout)}, nil
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
