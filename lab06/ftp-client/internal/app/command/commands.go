package command

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"strconv"
	"strings"
)

var (
	errAuth = errors.New("user authorization error")
)

type Query interface {
	Do(*bufio.Writer, *bufio.Reader) error
}

////////////////////////////// User //////////////////////////////

type User struct {
	Name string
}

func (u *User) Do(w *bufio.Writer, r *bufio.Reader) error {
	w.WriteString(fmt.Sprintf("USER %s\r\n", u.Name))
	w.Flush()

	s, err := r.ReadString('\n')

	if err = checkResponse(s, err, "331"); err != nil {
		return err
	}

	return nil
}

////////////////////////////// Pass //////////////////////////////

type Pass struct {
	Pass string
}

func (p *Pass) Do(w *bufio.Writer, r *bufio.Reader) error {
	w.WriteString(fmt.Sprintf("PASS %s\r\n", p.Pass))
	w.Flush()

	s, err := r.ReadString('\n')

	if err = checkResponse(s, err, "230"); err != nil {
		return err
	}

	return nil
}

////////////////////////////// Pasv //////////////////////////////

type Pasv struct{}

func (p *Pasv) Do(w *bufio.Writer, _ *bufio.Reader) {
	w.WriteString("PASV\r\n")
	w.Flush()
}

////////////////////////////// LIST //////////////////////////////

type List struct {
	Path string
}

func (l *List) Do(w *bufio.Writer, r *bufio.Reader) error {
	pasv := Pasv{}
	pasv.Do(w, r)
	s, err := r.ReadString('\n')

	if err = checkResponse(s, err, "227"); err != nil {
		return err
	}

	printer := Printer{make(chan struct{}, 1)}
	go printer.do(parseAddress(s))

	w.WriteString(fmt.Sprintf("LIST %s\r\n", l.Path))
	w.Flush()

	s, err = r.ReadString('\n')

	if err = checkResponse(s, err, "150"); err != nil {
		return err
	}

	<-printer.printed

	s, err = r.ReadString('\n')
	if err = checkResponse(s, err, "226"); err != nil {
		return err
	}

	return err
}

////////////////////////////// Retr //////////////////////////////

type Retr struct {
	Source string
	Target string
}

func (rtr *Retr) Do(w *bufio.Writer, r *bufio.Reader) error {
	pasv := Pasv{}
	pasv.Do(w, r)
	s, err := r.ReadString('\n')

	if err = checkResponse(s, err, "227"); err != nil {
		return err
	}

	d := Downloader{rtr.Target, make(chan struct{}, 1)}
	go d.do(parseAddress(s))

	w.WriteString(fmt.Sprintf("RETR %s\r\n", rtr.Source))
	w.Flush()

	s, err = r.ReadString('\n')

	if err = checkResponse(s, err, "150"); err != nil {
		return err
	}

	<-d.downloaded

	s, err = r.ReadString('\n')

	if err = checkResponse(s, err, "226"); err != nil {
		return err
	}

	return nil
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

func checkResponse(s string, e error, code string) error {
	if e != nil {
		return e
	}

	if !strings.HasPrefix(s, code) {
		return fmt.Errorf("%s", s[:len(s)-2])
	}

	return nil
}

////////////////////////////// Printer //////////////////////////////

type Printer struct {
	printed chan struct{}
}

func (p *Printer) do(addr string) {
	conn, err := net.Dial("tcp", addr)

	if err != nil {
		return
	}

	r := bufio.NewReader(conn)
	printResult(r)

	p.printed <- struct{}{}
}

func printResult(r *bufio.Reader) {
	for {
		s, e := r.ReadString('\n')
		if e != nil {
			break
		}
		fmt.Print(s)
	}
}

////////////////////////////// Downloader //////////////////////////////

type Downloader struct {
	dir        string
	downloaded chan struct{}
}

func (d *Downloader) do(addr string) {
	conn, err := net.Dial("tcp", addr)

	if err != nil {
		return
	}

	r := bufio.NewReader(conn)
	buff := bytes.Buffer{}
	for {
		b, err := r.ReadByte()
		if err != nil {
			break
		}
		buff.WriteByte(b)
	}
	os.WriteFile(d.dir, buff.Bytes(), fs.ModePerm)

	d.downloaded <- struct{}{}
}
