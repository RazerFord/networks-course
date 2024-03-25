package main

import "ftpclient/internal/app/ftp"

func main() {
	s, _ := ftp.NewServer("localhost", 21)
	s.Auth("testftp", "1234")
	s.Run()
}
