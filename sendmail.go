package common

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/google/uuid"
	"mime/quotedprintable"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
	"time"
)

type Sendmail struct {
	Host     string
	Port     int
	Username string
	Password string
	CAFile   string
	c        *smtp.Client
}

func NewSendmail(hostname string, port int, username, password, CAFile string) (*Sendmail, error) {

	c := Sendmail{
		Host:     hostname,
		Port:     port,
		Username: username,
		Password: password,
		CAFile:   CAFile,
	}
	if c.Port == 0 {
		c.Port = 465
	}
	caCertPool := x509.NewCertPool()
	if c.CAFile != "" {
		caCert, err := os.ReadFile(c.CAFile)
		if err != nil {
			return nil, Fatal(err)
		}
		ok := caCertPool.AppendCertsFromPEM(caCert)
		if !ok {
			return nil, Fatalf("failed appending CA cert: %s", c.CAFile)
		}
	} else {
		var err error
		caCertPool, err = x509.SystemCertPool()
		if err != nil {
			return nil, Fatal(err)
		}
	}

	tlsConfig := &tls.Config{
		RootCAs:    caCertPool,
		ServerName: c.Host,
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port), tlsConfig)
	if err != nil {
		return nil, Fatal(err)
	}

	c.c, err = smtp.NewClient(conn, c.Host)
	if err != nil {
		return nil, Fatal(err)
	}

	_, err = c.readPassword()
	if err != nil {
		return nil, Fatal(err)
	}
	return &c, nil
}

func (c *Sendmail) readPassword() (string, error) {
	password := c.Password
	if strings.HasPrefix(password, "@") {
		data, err := os.ReadFile(password[1:])
		if err != nil {
			return "", Fatal(err)
		}
		password = strings.TrimSpace(string(data))
	}

	return password, nil
}

func (c *Sendmail) Send(to, from, subject string, body []byte) error {

	password, err := c.readPassword()
	if err != nil {
		return Fatal(err)
	}

	err = c.c.Auth(smtp.PlainAuth("", c.Username, password, c.Host))
	if err != nil {
		return Fatal(err)
	}

	err = c.c.Mail(from)
	if err != nil {
		return Fatal(err)
	}
	err = c.c.Rcpt(to)
	if err != nil {
		return Fatal(err)
	}

	fp, err := c.c.Data()
	if err != nil {
		return Fatal(err)
	}

	data, err := formatMessage(to, from, subject, body)
	_, err = fp.Write(data)
	if err != nil {
		return Fatal(err)
	}

	err = fp.Close()
	if err != nil {
		return Fatal(err)
	}

	c.c.Quit()
	return nil
}

func extractDomain(email string) string {
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return "unknown_domain"
	}
	return addr.Address[strings.Index(addr.Address, "@")+1:]
}

func generateMessageID(from string) string {
	now := time.Now().UnixNano()
	uuid := uuid.New().String()
	mid := fmt.Sprintf("%d.%s@%s", now, uuid, extractDomain(from))
	return mid
}

func formatMessage(to, from, subject string, body []byte) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", to))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	buf.WriteString(fmt.Sprintf("Message-ID: %s\r\n", generateMessageID(from)))
	buf.WriteString("Content-Type: text/plain; charset=\"us-ascii\"\r\n")
	buf.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	buf.WriteString("\r\n")
	writer := quotedprintable.NewWriter(&buf)
	_, err := writer.Write(body)
	if err != nil {
		return nil, Fatal(err)
	}
	writer.Close()
	return buf.Bytes(), nil
}
