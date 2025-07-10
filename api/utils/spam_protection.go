package utils

import (
	"database/sql"
	"fmt"
	"net/http"
	"sync"
	"time"

	"toolcenter/config"

	"github.com/gin-gonic/gin"
)

type spamEntry struct {
	last         time.Time
	count        int
	strikes      int
	blockedUntil time.Time
}

var (
	spamMu  sync.Mutex
	spamMap = make(map[string]*spamEntry)
)

func isProxy(r *http.Request) bool {
	if r.Header.Get("X-Forwarded-For") != "" {
		return true
	}
	if r.Header.Get("Via") != "" || r.Header.Get("Forwarded") != "" {
		return true
	}
	return false
}

// ApplySpamSanction decreases the user status and logs the action.
func ApplySpamSanction(userID string) {
	db, err := config.OpenDB()
	if err != nil {
		return
	}
	defer db.Close()
	cfg := config.Get()
	newStatus, err := DecreaseStatus(db, userID, cfg.AntiSpam.StatusDecrease)
	if err != nil {
		return
	}
	reason := "Abus de l'API (spam)"
	var end *time.Time
	if newStatus == "Banned" {
		t := time.Now().Add(time.Duration(cfg.AntiSpam.BanHours) * time.Hour)
		end = &t
	}
	recordSanction(db, userID, reason, end)
	var email string
	_ = db.QueryRow("SELECT email FROM users WHERE user_id=?", userID).Scan(&email)
	if email != "" {
		body := "Votre compte a été sanctionné pour spam. Nouveau statut: " + newStatus
		_ = QueueEmail(db, email, "Sanction ToolCenter", body)
	}
}

func recordSanction(db *sql.DB, userID, reason string, end *time.Time) {
	if end != nil {
		_, _ = db.Exec(`INSERT INTO moderation_actions (user_id, action_type, reason, start_date, end_date) VALUES (?, 'Ban', ?, NOW(), ?)`, userID, reason, *end)
	} else {
		_, _ = db.Exec(`INSERT INTO moderation_actions (user_id, action_type, reason, start_date) VALUES (?, 'Warn', ?, NOW())`, userID, reason)
	}
}

// SpamProtectionMiddleware blocks aggressive clients and triggers sanctions on abuse.
func SpamProtectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Get()
		if !cfg.AntiSpam.Enabled {
			c.Next()
			return
		}

		ip := c.ClientIP()
		if ip == "" {
			ip = "unknown"
		}
		proxy := isProxy(c.Request)
		spamMu.Lock()
		ent, ok := spamMap[ip]
		if !ok {
			ent = &spamEntry{}
			spamMap[ip] = ent
		}
		now := time.Now()
		if ent.blockedUntil.After(now) {
			retryAfter := int(ent.blockedUntil.Sub(now).Seconds())
			spamMu.Unlock()
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			c.AbortWithStatusJSON(429, gin.H{
				"success":     false,
				"message":     "Trop de requêtes, veuillez patienter.",
				"retry_after": retryAfter,
			})
			return
		}
		interval := time.Duration(cfg.AntiSpam.IntervalSeconds) * time.Second
		if proxy {
			interval /= 2
		}
		if now.Sub(ent.last) < interval {
			ent.count++
		} else {
			ent.count = 1
		}
		ent.last = now
		if ent.count > cfg.AntiSpam.RequestThreshold {
			ent.strikes++
			block := time.Duration(cfg.AntiSpam.InitialBlockSeconds) * time.Second * time.Duration(ent.strikes)
			if proxy {
				block *= 2
			}
			ent.blockedUntil = now.Add(block)
			ent.count = 0
		}
		strikes := ent.strikes
		spamMu.Unlock()

		if strikes >= cfg.AntiSpam.MaxStrike {
			uid, _, _, _, err := Check(c, CheckOpts{RequireToken: true})
			if err == nil {
				ApplySpamSanction(uid)
			}
			spamMu.Lock()
			ent.strikes = 0
			spamMu.Unlock()
		}

		c.Next()
	}
}
