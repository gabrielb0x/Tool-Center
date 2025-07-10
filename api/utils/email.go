package utils

import (
    "log"
    "toolcenter/config"
    "gopkg.in/gomail.v2"
)

// SendEmail sends a plain text email using SMTP settings from config.
func SendEmail(to, subject, body string) error {
    cfg := config.Get()
    msg := gomail.NewMessage()
    msg.SetHeader("From", cfg.Email.From)
    msg.SetHeader("To", to)
    msg.SetHeader("Subject", subject)
    msg.SetBody("text/plain", body)

    d := gomail.NewDialer(cfg.Email.Host, cfg.Email.Port, cfg.Email.Username, cfg.Email.Password)
    d.SSL = true
    if err := d.DialAndSend(msg); err != nil {
        log.Println("send email:", err)
        return err
    }
    return nil
}
