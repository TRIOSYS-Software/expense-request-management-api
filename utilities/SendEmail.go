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
	mailcowServer := configs.Envs.SMTP_HOST         // Replace with your Mailcow server address
	port, _ := strconv.Atoi(configs.Envs.SMTP_PORT) // 587 (STARTTLS) or 465 (SSL/TLS)
	username := configs.Envs.EMAIL_USERNAME         // Mailcow email address
	password := configs.Envs.EMAIL_PASSWORD         // Mailcow application password

	// Authentication
	auth := smtp.PlainAuth("", username, password, mailcowServer)

	// TLS Configuration
	tlsConfig := &tls.Config{
		ServerName:         mailcowServer,
		InsecureSkipVerify: false, // Set to false if using valid certificate
	}
	// Connect to SMTP server
	conn, err := smtp.Dial(fmt.Sprintf("%s:%d", mailcowServer, port))
	if err != nil {
		return err
	}
	defer conn.Close()

	// Start TLS (for port 587)
	if ok, _ := conn.Extension("STARTTLS"); ok {
		if err = conn.StartTLS(tlsConfig); err != nil {
			return err
		}
	}

	// Authenticate
	if err = conn.Auth(auth); err != nil {
		return err
	}

	// Set sender and recipient
	if err = conn.Mail(username); err != nil {
		return err
	}

	for _, recipient := range to {
		if err = conn.Rcpt(recipient); err != nil {
			return err
		}
	}

	// Send email body
	w, err := conn.Data()
	if err != nil {
		return err
	}
	defer w.Close()

	fromHeader := "From: Admin\r\n"
	toHeader := fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")) // Convert []string to, ", "))
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
			"\r\n" + // Empty line between headers and body
			body,
	)

	if _, err = w.Write(msg); err != nil {
		return err
	}

	conn.Quit()
	return nil
}
