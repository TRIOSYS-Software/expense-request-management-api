package helper

import (
	"net/smtp"
)

func SendEmail(to []string, subject string, body string) error {
	auth := smtp.PlainAuth("", "thawthuhan9@gmail.com", "ozedxidfiyjgmjpp", "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, "thawthuhan9@gmail.com", to, []byte("Subject: "+subject+"\n\n"+body))
	return err
}
