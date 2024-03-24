package ftp

import (
	"bufio"
	"errors"
	"fmt"
	"ftpclient/internal/app/command"
	"ftpclient/internal/app/common"
	"net"
	"strconv"
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
	pasv := command.Pasv{}
	pasv.Do(ser.w, ser.r)
	s, err := ser.r.ReadString('\n')
	common.Check(err)
	common.Require(!strings.HasPrefix(s, "227"), fmt.Errorf("%s", s[:len(s)-2]))

	fmt.Println(s)
	lb := strings.Index(s, "(") + 1
	rb := strings.Index(s, ")")
	ipslice := strings.Split(s[lb:rb], ",")
	ipstr := strings.Join(ipslice[0:4], ".")
	ip := net.ParseIP(ipstr)

	port := 0
	step := 256
	for _, v := range ipslice[4:] {
		n, _ := strconv.Atoi(v)
		port += n * step
		step /= 256
	}

	printed := make(chan struct{})
	go func() {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
		common.Check(err)

		r := bufio.NewReader(conn)

		for {
			s, e := r.ReadString('\n')
			fmt.Print(s)
			if e != nil {
				fmt.Println(e)
				fmt.Println(s)
				break
			}
		}
		printed <- struct{}{} // double kik
	}()

	ser.w.WriteString("LIST\r\n")
	ser.w.Flush()

	<-printed
	s, e := ser.r.ReadString('\n')
	common.Check(e)
	fmt.Println(s)
	return nil
}

func (ser *Server) Close() {
	if ser.conn != nil {
		(*ser.conn).Close()
	}
}
