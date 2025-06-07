package worker

import (
	"log"
	"time"

	"toolcenter/config"

	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/gomail.v2"
)

// Start launches a periodic cleanup worker based on configuration values.
func Start() {
	cfg := config.Get()
	if cfg.Cleanup.CheckInterval == 0 {
		cfg.Cleanup.CheckInterval = 600
	}
	ticker := time.NewTicker(time.Duration(cfg.Cleanup.CheckInterval) * time.Second)
	go func() {
		for {
			run()
			<-ticker.C
		}
	}()
}

func run() {
	cfg := config.Get()
	db, err := config.OpenDB()
	if err != nil {
		log.Println("cleanup: open db:", err)
		return
	}
	defer db.Close()

	grace := cfg.Cleanup.GracePeriod
	if grace == 0 {
		grace = 10
	}

	rows, err := db.Query(`SELECT user_id,email FROM users WHERE email_verified_at IS NULL AND created_at < NOW() - INTERVAL ? MINUTE`, grace)
	if err != nil {
		log.Println("cleanup: select users:", err)
		return
	}
	for rows.Next() {
		var id, email string
		if err := rows.Scan(&id, &email); err != nil {
			log.Println("cleanup: scan user:", err)
			continue
		}
		sendEmail(email, "Compte supprimé", "Votre compte a été supprimé faute de vérification.")
		if _, err := db.Exec("DELETE FROM users WHERE user_id=?", id); err != nil {
			log.Println("cleanup: delete user:", err)
		}
	}
	rows.Close()

	rows2, err := db.Query("SELECT queue_id,to_email,subject,body FROM email_queue")
	if err != nil {
		log.Println("cleanup: select queue:", err)
		return
	}
	for rows2.Next() {
		var qid int
		var to, sub, body string
		if err := rows2.Scan(&qid, &to, &sub, &body); err != nil {
			log.Println("cleanup: scan queue:", err)
			continue
		}
		sendEmail(to, sub, body)
		if _, err := db.Exec("DELETE FROM email_queue WHERE queue_id=?", qid); err != nil {
			log.Println("cleanup: delete queue:", err)
		}
	}
	rows2.Close()
}

func sendEmail(to, subject, body string) {
	cfg := config.Get()
	msg := gomail.NewMessage()
	msg.SetHeader("From", cfg.Email.From)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", body)

	d := gomail.NewDialer(cfg.Email.Host, cfg.Email.Port, cfg.Email.Username, cfg.Email.Password)
	d.SSL = true
	if err := d.DialAndSend(msg); err != nil {
		log.Printf("cleanup: send email to %s failed: %v", to, err)
	}
}
