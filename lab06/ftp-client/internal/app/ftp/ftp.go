package ftp

import (
	"bufio"
	"errors"
	"fmt"
	"ftpclient/internal/app/command"
	"net"
	"os"
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

	fmt.Println("Authorized")
	return nil
}

func (ser *Server) Run() error {
	r := bufio.NewReader(os.Stdin)
	for {
		cmd, err := r.ReadString('\n')
		if err != nil {
			return err
		}
		cmd = strings.Trim(cmd, " \n\r")
		switch strings.ToLower(cmd) {
		case "list":
			{
				list := command.List{}
				err := list.Do(ser.w, ser.r)
				if err != nil {
					fmt.Println(err.Error())
				}
			}
		case "retr":
			{
				fmt.Println("Select source:")
				source, _ := r.ReadString('\n')

				fmt.Println("Select target:")
				target, _ := r.ReadString('\n')

				retr := command.Retr{Source: source, Target: target}
				err := retr.Do(ser.w, ser.r)
				if err != nil {
					fmt.Println(err.Error())
				} else {
					fmt.Println("File downloaded")
				}
			}
		case "stor":
			{
				fmt.Println("Select source:")
				source, _ := r.ReadString('\n')

				fmt.Println("Select target:")
				target, _ := r.ReadString('\n')

				stor := command.Stor{Source: source, Target: target}
				err := stor.Do(ser.w, ser.r)
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println("File uploaded")
				}
			}
		case "quit":
			{
				quit := command.Quit{}
				err := quit.Do(ser.w, ser.r)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println("Goodbye")
				return nil
			}
		default:
			{
				fmt.Println("Unknown command")
			}
		}
	}
}

func (ser *Server) Close() {
	if ser.conn != nil {
		(*ser.conn).Close()
	}
}
