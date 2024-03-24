package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

var (
	errConnServer = errors.New("server connection error")
	errAuth       = errors.New("user authorization error")
)

func main() {
	conn, err := net.Dial("tcp", "ftp.dlptest.com:21")
	check(err)
	defer conn.Close()

	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	s, err := r.ReadString('\n')
	check(err)
	require(!strings.HasPrefix(s, "220"), fmt.Errorf("%s: %w", s[:len(s)-2], errConnServer))

	w.WriteString("USER dlpuser\r\n")
	w.Flush()

	s, err = r.ReadString('\n')
	check(err)
	require(!strings.HasPrefix(s, "331"), fmt.Errorf("%s: %w", s[:len(s)-2], errAuth))

	w.WriteString("PASS rNrKYTX9g7z3RgJRmxWuGHbeu\r\n")
	w.Flush()

	s, err = r.ReadString('\n')
	check(err)
	require(!strings.HasPrefix(s, "230"), fmt.Errorf("%s: %w", s[:len(s)-2], errAuth))

	w.WriteString("PASV\r\n")
	w.Flush()

	s, err = r.ReadString('\n')
	check(err)
	require(!strings.HasPrefix(s, "227"), fmt.Errorf("%s", s[:len(s)-2]))

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
		check(err)

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
		printed <- struct{}{}
	}()

	w.WriteString("LIST\r\n")
	w.Flush()

	<-printed
	s, e := r.ReadString('\n')
	check(e)
	fmt.Println(s)
}

func require(b bool, err error) {
	if b {
		fmt.Println(err)
		os.Exit(1)
	}
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
