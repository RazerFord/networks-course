package smtpclient

import (
	"fmt"
	"os"

	"gopkg.in/gomail.v2"
)

////////////////////////////// Client //////////////////////////////

type Client struct {
	To          string // recipient
	From        string // sender
	Pass        string // sender password
	Smtpaddress string // smtp server address
	Smtpport    int    // smtp server port
}

func NewClient(to, from, pass, smtpaddress string, smtpport int) *Client {
	return &Client{
		To:          to,
		From:        from,
		Pass:        pass,
		Smtpaddress: smtpaddress,
		Smtpport:    smtpport,
	}
}

func (c *Client) SendMail(mesg Message) {
	msg := gomail.NewMessage()

	msg.SetHeader("From", c.From)
	msg.SetHeader("To", c.To)
	msg.SetHeader("Subject", "")
	msg.SetBody(mesg.Mime, mesg.Body)

	d := gomail.NewDialer("smtp.gmail.com", 587, "desk10567@gmail.com", "mjrg mjza vwkm fknw")

	if err := d.DialAndSend(msg); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

////////////////////////////// Message //////////////////////////////

type Message struct {
	Body string
	Mime string
}