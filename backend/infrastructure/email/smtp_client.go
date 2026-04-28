package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strings"
)

type SMTPClient struct {
	host     string
	port     string
	username string
	password string
}

func NewSMTPClient() *SMTPClient {
	return &SMTPClient{
		host:     strings.TrimSpace(os.Getenv("SMTP_HOST")),
		port:     strings.TrimSpace(os.Getenv("SMTP_PORT")),
		username: strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		password: strings.TrimSpace(os.Getenv("SMTP_PASSWORD")),
	}
}



func (c *SMTPClient) SendMail(to, subject, body string) error {
	addr := net.JoinHostPort(c.host, c.port)

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		c.username, to, subject, body,
	)

	auth := smtp.PlainAuth("", c.username, c.password, c.host)

	tlsConfig := &tls.Config{
		ServerName:         c.host,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true,
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}

	client, err := smtp.NewClient(conn, c.host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("smtp starttls: %w", err)
	}

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	if err := client.Mail(c.username); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	if _, err := fmt.Fprint(w, msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}

	return w.Close()
}

