package ftp

import (
	"bufio"
	"errors"
	"fmt"
	"ftpclient/internal/app/command"
	"net"
	"strings"
)

var (
	errConnServer = errors.New("server connection error")
)

type Server struct {
	Address string
	Port    int
	conn    *net.Conn
	r       *bufio.Reader
	w       *bufio.Writer
}

func NewServer(address string, port int) (*Server, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", address, port))

	if err != nil {
		return nil, err
	}

	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	s, err := r.ReadString('\n')

	if err != nil {
		conn.Close()
		return nil, err
	}

	if !strings.HasPrefix(s, "220") {
		conn.Close()
		return nil, fmt.Errorf("%s: %w", s[:len(s)-2], errConnServer)
	}

	return &Server{
		Address: address,
		Port:    port,
		conn:    &conn,
		r:       r,
		w:       w,
	}, nil
}

func (ser *Server) Auth(name, pass string) error {
	cmds := []command.Query{
		&command.User{Name: name},
		&command.Pass{Pass: pass},
	}

	for _, c := range cmds {
		if err := c.Do(ser.w, ser.r); err != nil {
			return err
		}
	}

	return nil
}

func (ser *Server) Run() error {
	// list := command.List{}
	// list.Do(ser.w, ser.r)

	// retr := command.Retr{Source: "actorstoday.txt", Target: "./" + "actorstoday.txt"}
	// err := retr.Do(ser.w, ser.r)
	// fmt.Println(err)

	stor := command.Stor{Source: "./hello.txt", Target: "hello.txt"}
	err := stor.Do(ser.w, ser.r)
	fmt.Println(err)
	return nil
}

func (ser *Server) Close() {
	if ser.conn != nil {
		(*ser.conn).Close()
	}
}
