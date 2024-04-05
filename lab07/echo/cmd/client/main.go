package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"math"
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
	length := flag.Int("len", 63, "message length")
	n := flag.Int("n", 10, "number of messages")
	flag.Parse()

	client := NewClient(*port)

	min := math.MaxFloat64
	max := -1.0
	avg := 0.0
	sum := 0.0
	c := 0
	s := time.Now()
	for range *n {
		start := time.Now()
		client.Write(genString(*length))
		msg, err := client.ReadTimeout(time.Second)
		end := time.Now()

		if errors.Is(err, os.ErrDeadlineExceeded) {
			fmt.Printf("request timed out\n")
		} else {
			l := len([]byte(msg))
			addr := client.conn.RemoteAddr().String()
			t := float64(end.Sub(start)) / 1e6

			sum += t
			c++
			min = math.Min(min, t)
			max = math.Max(max, t)
			avg = sum / float64(c)

			fmt.Printf("%d bytes from %v: time=%.2f ms\n", l, addr, t)
		}
	}
	e := time.Now()
	m := float64(*n)
	fmt.Printf("\n--- %s ping statistics ---\n", client.conn.RemoteAddr().String())
	fmt.Printf("%d packets transmitted, %d received, %d%% packet loss, time %0.fms\n", *n, c, int((m-float64(c))/m*100), float64(e.Sub(s))/1e6)
	fmt.Printf("rtt min/avg/max = %.2f/%.2f/%.2f ms\n", min, avg, max)
}

func genString(n int) string {
	rs := make([]rune, n)
	l := len(letters)
	for i := range n {
		rs[i] = letters[rand.Intn(l)]
	}
	return string(rs)
}
