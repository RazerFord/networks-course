package smtpclient

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

const (
	local = "127.0.0.1"
)

var (
	errCode = errors.New("wrong code")
)

////////////////////////////// Client //////////////////////////////

type Client struct {
	To          string // recipient
	From        string // sender
	Pass        string // sender password
	Smtpaddress string // smtp server address
	Smtpport    int    // smtp server port
}

func NewClient(to, from, pass, smtpaddress string, smtpport int) (*Client, error) {
	return &Client{
		To:          to,
		From:        from,
		Pass:        pass,
		Smtpaddress: smtpaddress,
		Smtpport:    smtpport,
	}, nil
}

func (c *Client) SendMail(msg Message) error {
	conn, err := createConn(c.Smtpaddress, c.Smtpport)
	if err != nil {
		return err
	}
	defer conn.Close()

	connect := Connect{
		r: bufio.NewReader(conn),
		w: bufio.NewWriter(conn),
	}

	s, err := connect.r.ReadString('\n')

	if err != nil || !strings.HasPrefix(s, "220") {
		return err
	}

	if err := connect.send(fmt.Sprintf("HELO %s\r\n", local), "250"); err != nil {
		return err
	}

	if err := connect.send("AUTH LOGIN\r\n", "334"); err != nil {
		return err
	}

	if err := connect.send(fmt.Sprintf("%s\r\n", base64.StdEncoding.EncodeToString([]byte(c.From))), "334"); err != nil {
		return err
	}

	if err := connect.send(fmt.Sprintf("%s\r\n", base64.StdEncoding.EncodeToString([]byte(c.Pass))), "235"); err != nil {
		return err
	}

	if err := connect.send(fmt.Sprintf("MAIL FROM: <%s>\r\n", c.From), "250"); err != nil {
		return err
	}

	if err := connect.send(fmt.Sprintf("RCPT TO: <%s>\r\n", c.To), "250"); err != nil {
		return err
	}

	if err := connect.send("DATA\r\n", "354"); err != nil {
		return err
	}

	buff := bytes.Buffer{}
	setHeader(&buff, "From", c.From)
	setHeader(&buff, "To", c.To)
	setHeader(&buff, "Content-Type", msg.Mime)
	buff.WriteString(fmt.Sprintf("%s\r\n", msg.Body))
	buff.WriteString(".\r\n")

	if err := connect.send(buff.String(), "250"); err != nil {
		return err
	}

	connect.w.WriteString("QUIT\r\n")

	return nil
}

////////////////////////////// Message //////////////////////////////

type Message struct {
	Body string
	Mime string
}

////////////////////////////// Connect //////////////////////////////

type Connect struct {
	w *bufio.Writer
	r *bufio.Reader
}

func (c *Connect) send(msg, code string) error {
	c.w.WriteString(msg)
	c.w.Flush()
	s, err := c.r.ReadString('\n')

	if err != nil {
		return err
	}
	if !strings.HasPrefix(s, code) {
		return errCode
	}

	return nil
}

////////////////////////////// Utility //////////////////////////////

func setHeader(w *bytes.Buffer, key, value string) {
	w.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
}

func createConn(smtpaddress string, smtpport int) (*tls.Conn, error) {
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpaddress,
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", smtpaddress, smtpport), tlsconfig)
	if err != nil {
		return nil, err
	}
	return conn, err
}
