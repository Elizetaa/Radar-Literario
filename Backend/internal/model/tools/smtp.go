package tools

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"radar-literario/backend/internal/config"
	"radar-literario/backend/internal/model/interfaces"
	"strconv"
	"strings"
	"time"
)

type smtpMailer struct{ c config.Config }

func NewSMTPMailer(c config.Config) interfaces.Mailer { return &smtpMailer{c: c} }
func (m *smtpMailer) Send(ctx context.Context, to, subject, body string) error {
	if strings.ContainsAny(to+subject+m.c.SMTPFrom, "\r\n") {
		return fmt.Errorf("invalid mail headers")
	}
	dial := net.Dialer{Timeout: 5 * time.Second}
	conn, e := dial.DialContext(ctx, "tcp", net.JoinHostPort(m.c.SMTPHost, strconv.Itoa(m.c.SMTPPort)))
	if e != nil {
		return e
	}
	defer conn.Close()
	deadline := time.Now().Add(10 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	if e = conn.SetDeadline(deadline); e != nil {
		return e
	}
	client, e := smtp.NewClient(conn, m.c.SMTPHost)
	if e != nil {
		return e
	}
	defer client.Close()
	if m.c.SMTPTLS {
		if e = client.StartTLS(&tls.Config{ServerName: m.c.SMTPHost, MinVersion: tls.VersionTLS12}); e != nil {
			return e
		}
	}
	if m.c.SMTPUser != "" {
		if e = client.Auth(smtp.PlainAuth("", m.c.SMTPUser, m.c.SMTPPassword, m.c.SMTPHost)); e != nil {
			return e
		}
	}
	if e = client.Mail(m.c.SMTPFrom); e != nil {
		return e
	}
	if e = client.Rcpt(to); e != nil {
		return e
	}
	w, e := client.Data()
	if e != nil {
		return e
	}
	message := "From: " + m.c.SMTPFrom + "\r\nTo: " + to + "\r\nSubject: " + subject + "\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" + body + "\r\n"
	if _, e = w.Write([]byte(message)); e != nil {
		return e
	}
	if e = w.Close(); e != nil {
		return e
	}
	return client.Quit()
}
