package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/buildkite/shellwords"
	"net"
	"os"
	"os/exec"
	"strings"
)

func main() {
	ip := flag.String("ip", "127.0.0.1", "server IP")
	port := flag.Int("port", 8080, "server port")
	flag.Parse()

	addr := net.TCPAddr{
		IP:   net.ParseIP(*ip),
		Port: *port,
	}

	conn, err := net.ListenTCP("tcp", &addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for {
		tcp, err := conn.AcceptTCP()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		go func() {
			defer tcp.Close()
			r := bufio.NewReader(tcp)
			w := bufio.NewWriter(tcp)

			s, err := r.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			slice, err := shellwords.SplitPosix(strings.Trim(s, " \r\n"))

			if err != nil {
				fmt.Println(err)
				w.WriteString(err.Error())
				w.Flush()
				return
			}

			name := slice[0]
			args := slice[1:]

			cmd := exec.Command(name, args...)

			out, _ := cmd.CombinedOutput()

			w.Write(out)
			w.Flush()
		}()
	}
}
