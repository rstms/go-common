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

type SMTPServer struct {
	Host     string
	Port     int
	Username string
	Password string
	CAFile   string
}

func Sendmail(to, from, subject string, body []byte, smtpServer *SMTPServer) error {

	if smtpServer == nil {
		smtpServer = &SMTPServer{
			Host:     ViperGetString("smtp.host"),
			Port:     ViperGetInt("smtp.port"),
			Username: ViperGetString("smtp.username"),
			Password: ViperGetString("smtp.password"),
			CAFile:   ViperGetString("smtp.ca_file"),
		}
		if smtpServer.Port == 0 {
			smtpServer.Port = 465
		}
		passwordFile := ViperGetString("smtp.password_file")
		if passwordFile != "" {
			data, err := os.ReadFile(passwordFile)
			if err != nil {
				return Fatal(err)
			}
			smtpServer.Password = strings.TrimSpace(string(data))
		}
	}

	caCertPool := x509.NewCertPool()
	if smtpServer.CAFile != "" {
		caCert, err := os.ReadFile(smtpServer.CAFile)
		if err != nil {
			return Fatal(err)
		}
		ok := caCertPool.AppendCertsFromPEM(caCert)
		if !ok {
			return Fatalf("failed appending CA cert: %s", smtpServer.CAFile)
		}
	} else {
		var err error
		caCertPool, err = x509.SystemCertPool()
		if err != nil {
			return Fatal(err)
		}
	}

	tlsConfig := &tls.Config{
		RootCAs:    caCertPool,
		ServerName: smtpServer.Host,
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", smtpServer.Host, smtpServer.Port), tlsConfig)
	if err != nil {
		return Fatal(err)
	}

	c, err := smtp.NewClient(conn, smtpServer.Host)
	if err != nil {
		return Fatal(err)
	}

	err = c.Auth(smtp.PlainAuth("", smtpServer.Username, smtpServer.Password, smtpServer.Host))
	if err != nil {
		return Fatal(err)
	}

	err = c.Mail(from)
	if err != nil {
		return Fatal(err)
	}
	err = c.Rcpt(to)
	if err != nil {
		return Fatal(err)
	}

	fp, err := c.Data()
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

	c.Quit()
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
