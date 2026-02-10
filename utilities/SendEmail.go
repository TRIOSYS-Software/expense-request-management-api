package utilities

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"shwetaik-expense-management-api/configs"
	"strconv"
	"strings"
	"time"
)

func SendEmail(to []string, subject string, body string) error {
	auth := smtp.PlainAuth("", "thawthuhan9@gmail.com", "ozedxidfiyjgmjpp", "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, "thawthuhan9@gmail.com", to, []byte("Subject: "+subject+"\n\n"+body))
	return err
}

func SendMailViaMailcow(to []string, subject string, body string) error {
	// Mailcow SMTP Configuration
	mailcowServer := configs.Envs.SMTP_HOST
	portStr := configs.Envs.SMTP_PORT
	username := configs.Envs.EMAIL_USERNAME
	password := configs.Envs.EMAIL_PASSWORD

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid SMTP port: %v", err)
	}

	addr := fmt.Sprintf("%s:%d", mailcowServer, port)

	// Auth
	auth := smtp.PlainAuth("", username, password, mailcowServer)

	// TLS Config
	skipVerify := configs.Envs.Environment == "dev"
	tlsConfig := &tls.Config{
		ServerName:         mailcowServer,
		InsecureSkipVerify: skipVerify,
	}

	var conn *smtp.Client

	if port == 465 {
		tlsConn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to dial TLS: %v", err)
		}
		conn, err = smtp.NewClient(tlsConn, mailcowServer)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %v", err)
		}
	} else {
		conn, err = smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("failed to dial SMTP: %v", err)
		}

		if ok, _ := conn.Extension("STARTTLS"); ok {
			if err = conn.StartTLS(tlsConfig); err != nil {
				conn.Close()
				return fmt.Errorf("failed to start TLS: %v", err)
			}
		}
	}
	defer conn.Close()

	// Authenticate
	if err = conn.Auth(auth); err != nil {
		if strings.Contains(err.Error(), "unencrypted connection") {
			return fmt.Errorf("authentication failed: server requires TLS but connection is not encrypted (check STARTTLS support)")
		}
		return fmt.Errorf("failed to authenticate: %v", err)
	}

	// Set sender
	if err = conn.Mail(username); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}

	// Set recipients
	for _, recipient := range to {
		if err = conn.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %v", recipient, err)
		}
	}

	// Send Data
	w, err := conn.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %v", err)
	}

	// Headers
	fromHeader := fmt.Sprintf("From: %s\r\n", username)
	toHeader := fmt.Sprintf("To: %s\r\n", strings.Join(to, ", "))
	dateHeader := fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	subjectHeader := fmt.Sprintf("Subject: %s\r\n", subject)
	mimeHeader := "MIME-Version: 1.0\r\n"
	contentType := "Content-Type: text/plain; charset=UTF-8\r\n"

	msg := []byte(
		fromHeader +
			toHeader +
			dateHeader +
			subjectHeader +
			mimeHeader +
			contentType +
			"\r\n" +
			body,
	)

	if _, err = w.Write(msg); err != nil {
		w.Close()
		return fmt.Errorf("failed to write email body: %v", err)
	}

	if err = w.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %v", err)
	}

	return conn.Quit()
}
