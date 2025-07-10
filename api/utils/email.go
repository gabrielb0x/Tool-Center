package utils

import (
	"log"
	"strings"

	"toolcenter/config"

	gomail "gopkg.in/gomail.v2"
)

// SendEmail immediately sends an email using SMTP settings from the configuration.
func SendEmail(to, subject, body string) error {
	cfg := config.Get()
	msg := gomail.NewMessage()
	msg.SetHeader("From", cfg.Email.From)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	ctype := "text/plain"
	if strings.Contains(body, "<html") {
		ctype = "text/html"
	}
	msg.SetBody(ctype, body)

	d := gomail.NewDialer(cfg.Email.Host, cfg.Email.Port, cfg.Email.Username, cfg.Email.Password)
	d.SSL = true
	if err := d.DialAndSend(msg); err != nil {
		log.Printf("send email to %s failed: %v", to, err)
		return err
	}
	return nil
}
