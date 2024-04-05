package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"

	"echo/internal/app/logging"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyz")

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

func (c *Client) Write(msg string) (int, error) {
	n, err := c.rw.WriteString(msg + "\n")
	c.rw.Flush()
	return n, err
}

func main() {
	port := flag.Int("port", 8080, "client port")
	length := flag.Int("len", 10, "message length")
	n := flag.Int("n", 10, "number of messages")
	flag.Parse()

	client := NewClient(*port)

	for i := range *n {
		start := time.Now()
		client.Write(genString(*length))
		msg, err := client.ReadTimeout(time.Second)
		end := time.Now()

		if errors.Is(err, os.ErrDeadlineExceeded) {
			logging.Warn("Request timed out %d", i + 1)
		} else {
			logging.Info("received message: %s", msg)
			logging.Info("Ping %d %v", i+1, end.Sub(start).Seconds())
		}
	}
}

func genString(n int) string {
	rs := make([]rune, n)
	l := len(letters)
	for i := range n {
		rs[i] = letters[rand.Intn(l)]
	}
	return string(rs)
}
