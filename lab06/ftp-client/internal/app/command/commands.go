package command

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

var (
	errAuth = errors.New("user authorization error")
)

type Query interface {
	Do(*bufio.Writer, *bufio.Reader) error
}

type User struct {
	Name string
}

func (u *User) Do(w *bufio.Writer, r *bufio.Reader) error {
	w.WriteString(fmt.Sprintf("USER %s\r\n", u.Name))
	w.Flush()

	s, err := r.ReadString('\n')

	if err != nil {
		return err
	}

	if !strings.HasPrefix(s, "331") {
		return fmt.Errorf("%s: %w", s[:len(s)-2], errAuth)
	}

	return nil
}

type Pass struct {
	Pass string
}

func (p *Pass) Do(w *bufio.Writer, r *bufio.Reader) error {
	w.WriteString(fmt.Sprintf("PASS %s\r\n", p.Pass))
	w.Flush()

	s, err := r.ReadString('\n')

	if err != nil {
		return err
	}

	if !strings.HasPrefix(s, "230") {
		return fmt.Errorf("%s: %w", s[:len(s)-2], errAuth)
	}

	return nil
}

type Pasv struct{}

func (p *Pasv) Do(w *bufio.Writer, _ *bufio.Reader) {
	w.WriteString("PASV\r\n")
	w.Flush()
}

type List struct{}

func (l *List) Do(w *bufio.Writer, r *bufio.Reader) error {
	pasv := Pasv{}
	pasv.Do(w, r)
	s, err := r.ReadString('\n')

	if err != nil {
		return err
	}

	if !strings.HasPrefix(s, "227") {
		return fmt.Errorf("%s", s[:len(s)-2])
	}

	addr := parseAddress(s)

	printed := make(chan struct{})
	go func() {
		conn, err := net.Dial("tcp", addr)

		if err != nil {
			return
		}

		r := bufio.NewReader(conn)

		for {
			s, e := r.ReadString('\n')
			if e != nil {
				fmt.Println(s)
				break
			}
			fmt.Print(s)
		}
		printed <- struct{}{} // double kik
	}()

	w.WriteString("LIST\r\n")
	w.Flush()

	<-printed
	s, e := r.ReadString('\n')
	fmt.Println(s)

	return e
}

func parseAddress(s string) string {
	lb := strings.Index(s, "(") + 1
	rb := strings.Index(s, ")")

	ipslice := strings.Split(s[lb:rb], ",")
	ip := net.ParseIP(strings.Join(ipslice[0:4], "."))

	port := 0
	step := 256
	for _, v := range ipslice[4:] {
		n, _ := strconv.Atoi(v)
		port += n * step
		step /= 256
	}

	return fmt.Sprintf("%s:%d", ip, port)
}
