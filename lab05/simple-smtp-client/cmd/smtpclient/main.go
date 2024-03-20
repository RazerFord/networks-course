package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"io/ioutil"
	"smtpclient/internal/app/smtpclient"
)

var required = []string{"to", "from", "password", "smtp", "port", "message"}

func main() {
	client, message := create()

	client.SendMail(*message)
}

func create() (*smtpclient.Client, *smtpclient.Message) {
	to := flag.String("to", "", "mail recipient address")
	from := flag.String("from", "", "mail sender address")
	password := flag.String("password", "", "sender's email password")
	smtpaddress := flag.String("smtp", "", "smtp server address")
	smtpport := flag.Int("port", 0, "smtp server port")
	message := flag.String("message", "", "path to sent message")
	flag.Parse()

	received := map[string]bool{}
	flag.Visit(func(f *flag.Flag) {
		received[f.Name] = true
	})

	for _, v := range required {
		if _, ok := received[v]; !ok {
			fmt.Printf("argument \"%s\" not passed\n", v)
			os.Exit(1)
		}
	}

	mime := "text/plain"
	lastIdx := strings.LastIndex(*message, ".") 
	if lastIdx < len(*message) && (*message)[lastIdx:] == ".html" {
		mime = "text/html"
	}
	if msg, err := ioutil.ReadFile(*message); err != nil {
		fmt.Printf("message file not found")
		os.Exit(1)
	} else {
		*message = string(msg)
	}
	

	return &smtpclient.Client{
		To:          *to,
		From:        *from,
		Pass:        *password,
		Smtpaddress: *smtpaddress,
		Smtpport:    *smtpport,
	},
	&smtpclient.Message{
		Body: *message,
		Mime: mime,
	}
}
