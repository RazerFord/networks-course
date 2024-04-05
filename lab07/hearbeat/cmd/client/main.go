package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"

	"echo/internal/app/logging"
	"echo/internal/app/message"
)

const border = 0.2

type Client struct {
	conn *net.UDPConn
	rw   bufio.ReadWriter
}

func NewClient(port int) *Client {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))

	if err != nil {
		logging.Warn(err.Error())
		os.Exit(1)
	}

	conn, err := net.DialUDP(
		"udp",
		nil,
		addr,
	)

	if err != nil {
		logging.Warn(err.Error())
		os.Exit(1)
	}

	return &Client{
		conn: conn,
		rw:   *bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
	}
}

func (c *Client) Read() (string, error) {
	return c.rw.ReadString('\n')
}

func (c *Client) ReadTimeout(timeout time.Duration) (string, error) {
	c.conn.SetReadDeadline(time.Now().Add(timeout))
	return c.Read()
}

func (c *Client) Write(msg []byte) (int, error) {
	n, err := c.rw.Write(msg)
	c.rw.Flush()
	return n, err
}

func main() {
	port := flag.Int("port", 8080, "client port")
	timeout := flag.Int("timeout", 600, "timeout in milliseconds")
	flag.Parse()

	client := NewClient(*port)

	for c := 0; ; c++ {
		if rand.Float64() > border {
			m := message.NewMessage(c, time.Now())
			bs, err := message.ToBytes(m)
			if err != nil {
				logging.Warn(err.Error())
			}
			client.Write(bs)
		}
		time.Sleep(time.Millisecond * time.Duration(*timeout))
	}
}
