package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strings"
	"sync"
	"time"
)

type SMTPClient struct {
	host     string
	port     string
	username string
	password string

	mu   sync.Mutex
	pool []*pooledConn
}

type pooledConn struct {
	client *smtp.Client
	idleAt time.Time
}

const (
	smtpIdleTimeout = 60 * time.Second
	smtpDialTimeout = 10 * time.Second
)

func NewSMTPClient() *SMTPClient {
	return &SMTPClient{
		host:     strings.TrimSpace(os.Getenv("SMTP_HOST")),
		port:     strings.TrimSpace(os.Getenv("SMTP_PORT")),
		username: strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		password: strings.TrimSpace(os.Getenv("SMTP_PASSWORD")),
	}
}

func (c *SMTPClient) tlsConfig() *tls.Config {
	return &tls.Config{
		ServerName:         c.host,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true,
	}
}

func (c *SMTPClient) dial() (*smtp.Client, error) {
	addr := net.JoinHostPort(c.host, c.port)
	conn, err := net.DialTimeout("tcp", addr, smtpDialTimeout)
	if err != nil {
		return nil, fmt.Errorf("smtp dial: %w", err)
	}
	client, err := smtp.NewClient(conn, c.host)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("smtp client: %w", err)
	}
	if err := client.StartTLS(c.tlsConfig()); err != nil {
		client.Close()
		return nil, fmt.Errorf("smtp starttls: %w", err)
	}
	if err := client.Auth(smtp.PlainAuth("", c.username, c.password, c.host)); err != nil {
		client.Close()
		return nil, fmt.Errorf("smtp auth: %w", err)
	}
	return client, nil
}

// acquire takes an idle pooled client whose age is within smtpIdleTimeout, or
// dials a fresh one. Stale entries are closed and discarded.
func (c *SMTPClient) acquire() (*smtp.Client, error) {
	c.mu.Lock()
	now := time.Now()
	for len(c.pool) > 0 {
		idx := len(c.pool) - 1
		pc := c.pool[idx]
		c.pool = c.pool[:idx]
		if now.Sub(pc.idleAt) > smtpIdleTimeout {
			pc.client.Close()
			continue
		}
		c.mu.Unlock()
		if err := pc.client.Reset(); err != nil {
			pc.client.Close()
			c.mu.Lock()
			continue
		}
		return pc.client, nil
	}
	c.mu.Unlock()
	return c.dial()
}

// release returns a healthy client to the pool. Callers pass nil to indicate a
// poisoned client; nil clients are not pooled.
func (c *SMTPClient) release(client *smtp.Client) {
	if client == nil {
		return
	}
	c.mu.Lock()
	c.pool = append(c.pool, &pooledConn{client: client, idleAt: time.Now()})
	c.mu.Unlock()
}

func (c *SMTPClient) SendMail(to, subject, body string) error {
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		c.username, to, subject, body,
	)

	client, err := c.acquire()
	if err != nil {
		return err
	}

	if err := client.Mail(c.username); err != nil {
		client.Close()
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		client.Close()
		return fmt.Errorf("smtp rcpt to: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		client.Close()
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := fmt.Fprint(w, msg); err != nil {
		client.Close()
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		client.Close()
		return fmt.Errorf("smtp body close: %w", err)
	}

	c.release(client)
	return nil
}
