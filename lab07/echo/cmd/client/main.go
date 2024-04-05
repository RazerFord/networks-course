package main

import (
	"bufio"
	"errors"
	"flag"
	"net"
	"os"
	"time"

	"echo/internal/app/logging"
)

type Client struct {
	conn *net.UDPConn
	rw   *bufio.ReadWriter
}

func NewClient(port int) *Client {
	conn, err := net.DialUDP(
		"udp",
		nil,
		&net.UDPAddr{
			IP:   net.IPv4bcast,
			Port: port,
		},
	)

	if err != nil {
		logging.Warn(err.Error())
		os.Exit(1)
	}

	return &Client{
		conn: conn,
		rw:   bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
	}
}

func (c *Client) Read() (string, error) {
	return c.rw.Reader.ReadString('\n')
}

func (c *Client) ReadTimeout(timeout time.Duration) (string, error) {
	c.conn.SetDeadline(time.Now().Add(timeout))
	return c.Read()
}

func (c *Client) Write(msg string) (int, error) {
	n, err := c.rw.Writer.WriteString(msg)
	c.rw.Writer.Flush()
	return n, err
}

func main() {
	port := flag.Int("port", 8080, "client port")
	flag.Parse()

	client := NewClient(*port)

	for range 10 {
		client.Write("Hello world\n")
		msg, err := client.ReadTimeout(time.Second)

		if errors.Is(err, os.ErrDeadlineExceeded) {
			logging.Warn("timeout")
		} else {
			logging.Info("received message: %s", msg)
		}
	}
}
