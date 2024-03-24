package main

import "ftpclient/internal/app/ftp"

func main() {
	s, _ := ftp.NewServer("ftp.dlptest.com", 21)
	s.Auth("dlpuser", "rNrKYTX9g7z3RgJRmxWuGHbeu")
	s.Run()
}
