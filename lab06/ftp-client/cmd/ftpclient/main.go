package main

import (
	"flag"
	"fmt"
	"ftpclient/internal/app/ftp"
)

func main() {
	addr := flag.String("addr", "localhost", "ftp server address")
	port := flag.Int("port", 21, "ftp server port")
	user := flag.String("user", "anonymous", "username")
	pass := flag.String("pass", "anonymous@domain.com", "user password")
	flag.Parse()
	
	var s ftp.Server
	if s, err := ftp.NewServer(*addr, *port); err != nil {
		fmt.Println(err)
		return
	} else if err := s.Auth(*user, *pass); err != nil {
		fmt.Println(err)
	} else if err := s.Run(); err != nil {
		fmt.Println(err)
	}
	s.Close()
}
