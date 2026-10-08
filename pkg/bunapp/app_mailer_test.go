package bunapp

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/uptrace/uptrace/pkg/bunconf"
	mail "github.com/wneessen/go-mail"
)

// NewMailer must not force implicit TLS on the SMTP connection.
// mail.WithSSLPort(false) misleadingly enables implicit TLS (it calls
// SetSSLPort(true, false)), which breaks STARTTLS ports such as 587 with
// "tls: first record does not look like a TLS handshake". See issues #427, #614.
func TestNewMailerDoesNotUseImplicitTLS(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %s", err)
	}
	defer ln.Close()

	got := make(chan []byte, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(10 * time.Second))

		// Send a plaintext greeting like any STARTTLS-capable server would.
		// A client with implicit TLS enabled never reads it and starts the
		// TLS handshake instead; a correct client answers with EHLO.
		_, _ = io.WriteString(conn, "220 localhost ESMTP fake\r\n")

		buf := make([]byte, 5)
		n, _ := io.ReadFull(conn, buf)
		got <- buf[:n]
	}()

	conf := &bunconf.Config{}
	conf.SMTPMailer.Enabled = true
	conf.SMTPMailer.Host = "127.0.0.1"
	conf.SMTPMailer.Port = ln.Addr().(*net.TCPAddr).Port
	conf.SMTPMailer.AuthType = mail.SMTPAuthPlain

	client, err := NewMailer(conf)
	if err != nil {
		t.Fatalf("NewMailer: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = client.DialWithContext(ctx) // expected to fail; we only care about the first bytes

	select {
	case b := <-got:
		if len(b) >= 2 && b[0] == 0x16 && b[1] == 0x03 {
			t.Fatalf("client started an implicit TLS handshake (first bytes % x); want plaintext EHLO", b)
		}
		if len(b) < 4 || string(b[:4]) != "EHLO" {
			t.Fatalf("client sent %q; want plaintext EHLO first", b)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for the client to speak SMTP")
	}
}
