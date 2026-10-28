package tools_test

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"radar-literario/backend/internal/config"
	"radar-literario/backend/internal/model/tools"
)

type smtpCapture struct {
	recipient string
	body      string
	err       error
}

// testSMTPServer is a loopback SMTP peer, so the adapter really sends a message
// through net/smtp instead of replacing the transport with a fake.
func testSMTPServer(t *testing.T, rejectRecipient bool) (config.Config, <-chan smtpCapture) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	results := make(chan smtpCapture, 1)
	go func() {
		result := smtpCapture{}
		defer func() { results <- result }()
		conn, err := listener.Accept()
		if err != nil {
			result.err = err
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
		reader := textproto.NewReader(bufio.NewReader(conn))
		write := func(line string) bool { _, e := fmt.Fprint(conn, line+"\r\n"); result.err = e; return e == nil }
		if !write("220 localhost test SMTP") {
			return
		}
		for {
			line, e := reader.ReadLine()
			if e != nil {
				return
			}
			switch {
			case strings.HasPrefix(line, "EHLO "):
				if !write("250-localhost\r\n250 SIZE 1048576") {
					return
				}
			case strings.HasPrefix(line, "HELO "):
				if !write("250 localhost") {
					return
				}
			case strings.HasPrefix(line, "MAIL FROM:"):
				if !write("250 sender accepted") {
					return
				}
			case strings.HasPrefix(line, "RCPT TO:"):
				result.recipient = line
				if rejectRecipient {
					write("550 recipient refused")
					return
				}
				if !write("250 recipient accepted") {
					return
				}
			case line == "DATA":
				if !write("354 finish with dot") {
					return
				}
				body, e := reader.ReadDotBytes()
				if e != nil {
					result.err = e
					return
				}
				result.body = string(body)
				if !write("250 queued") {
					return
				}
			case line == "QUIT":
				write("221 goodbye")
				return
			default:
				if !write("500 command unsupported") {
					return
				}
			}
		}
	}()
	address := listener.Addr().(*net.TCPAddr)
	return config.Config{SMTPHost: "127.0.0.1", SMTPPort: address.Port, SMTPFrom: "sender@example.test"}, results
}

func TestSMTPAdapterDeliversThroughLocalSMTP(t *testing.T) {
	cfg, received := testSMTPServer(t, false)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := tools.NewSMTPMailer(cfg).Send(ctx, "reader@example.test", "Recuperacao de senha", "Mensagem de teste com link local."); err != nil {
		t.Fatal(err)
	}
	select {
	case message := <-received:
		if message.err != nil {
			t.Fatal(message.err)
		}
		if !strings.Contains(message.recipient, "reader@example.test") || !strings.Contains(message.body, "Mensagem de teste com link local.") || !strings.Contains(message.body, "Content-Type: text/plain; charset=UTF-8") {
			t.Fatal("SMTP envelope/body missing")
		}
	case <-ctx.Done():
		t.Fatal("SMTP capture did not finish")
	}
}

func TestSMTPAdapterRejectsInjectionAndDependencyFailures(t *testing.T) {
	if err := tools.NewSMTPMailer(config.Config{SMTPFrom: "sender@example.test"}).Send(context.Background(), "reader@example.test\r\nBcc: other@example.test", "Subject", "Body"); err == nil {
		t.Fatal("accepted header injection")
	}
	cfg, received := testSMTPServer(t, true)
	if err := tools.NewSMTPMailer(cfg).Send(context.Background(), "reader@example.test", "Subject", "Body"); err == nil {
		t.Fatal("SMTP recipient rejection ignored")
	}
	<-received
	cfg, received = testSMTPServer(t, false)
	cfg.SMTPTLS = true
	if err := tools.NewSMTPMailer(cfg).Send(context.Background(), "reader@example.test", "Subject", "Body"); err == nil {
		t.Fatal("required STARTTLS silently bypassed")
	}
	<-received
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cfg.SMTPPort = 1
	if err := tools.NewSMTPMailer(cfg).Send(ctx, "reader@example.test", "Subject", "Body"); err == nil {
		t.Fatal("canceled SMTP operation succeeded")
	}
}
