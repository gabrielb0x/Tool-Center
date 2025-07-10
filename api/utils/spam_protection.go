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

func ApplySpamSanction(userID string) {
	db, err := config.OpenDB()
	if err != nil {
		return
	}
	defer db.Close()
	cfg := config.Get()
	prevStatus, newStatus, err := DecreaseStatus(db, userID, cfg.AntiSpam.StatusDecrease)
	if err != nil {
		return
	}
	reason := "Abus de l'API (spam)"
	var end *time.Time
	if newStatus == "Banned" {
		t := time.Now().Add(time.Duration(cfg.AntiSpam.BanHours) * time.Hour)
		end = &t
	}
	recordSanction(db, userID, prevStatus, reason, end)
	var email string
	_ = db.QueryRow("SELECT email FROM users WHERE user_id=?", userID).Scan(&email)
	if email != "" {
		year := time.Now().Year()
		body := fmt.Sprintf(`<!DOCTYPE html>
<html lang="fr">
<head>
<meta charset="UTF-8">
<title>Sanction de votre compte Tool Center</title>
<link href="https://fonts.googleapis.com/css2?family=Poppins:wght@300;400;500;600;700&display=swap" rel="stylesheet">
<style>
body{
	margin:0;
	padding:0;
	background-color:#121212;
	font-family:'Poppins',sans-serif;
	color:#e0e0e0;
}
.container{
	max-width:600px;
	margin:30px auto;
	background:#1e1e1e;
	border-radius:12px;
	overflow:hidden;
	box-shadow:0 10px 30px rgba(0,0,0,0.3);
	border:1px solid #333;
}
.header{
	padding:30px;
	text-align:center;
	background:linear-gradient(135deg,#3000FF 0%%,#6200EA 100%%);
}
.logo{
	max-width:180px;
	height:auto;
	transition:transform .3s ease;
}
.logo:hover{transform:scale(1.05);}
.content{padding:30px;}
h1{
	color:#fff;
	font-size:28px;
	margin-bottom:20px;
	font-weight:600;
	text-align:center;
}
p{
	font-size:16px;
	line-height:1.6;
	margin-bottom:20px;
	color:#b0b0b0;
}
.highlight{color:#fff;font-weight:500;}
.button{
	display:inline-block;
	background:linear-gradient(135deg,#3000FF 0%%,#6200EA 100%%);
	color:#fff !important;
	text-decoration:none;
	padding:14px 28px;
	border-radius:8px;
	font-weight:600;
	font-size:16px;
	margin:25px auto;
	text-align:center;
	transition:all .3s ease;
	box-shadow:0 4px 15px rgba(48,0,255,.3);
	border:none;
	cursor:pointer;
}
.button:hover{
	transform:translateY(-2px);
	box-shadow:0 6px 20px rgba(48,0,255,.4);
}
.expire-note{
	font-size:14px;
	color:#888;
	text-align:center;
	margin-top:10px;
}
.footer{
	padding:20px;
	text-align:center;
	background:#121212;
	border-top:1px solid #333;
	font-size:12px;
	color:#666;
}
.divider{
	height:1px;
	background:linear-gradient(90deg,transparent,#333,transparent);
	margin:20px 0;
}
.security-note{
	background:rgba(48,0,255,.1);
	padding:15px;
	border-radius:8px;
	border-left:4px solid #3000FF;
	font-size:14px;
	color:#fff;
	margin-top:25px;
}
@media only screen and (max-width:600px){
	.container{margin:0;border-radius:0;}
	.content{padding:25px;}
}
</style>
</head>
<body>
<div class="container">
	<div class="header">
	<img src="https://tool-center.fr/assets/tc_logo.webp" alt="Tool Center Logo"
		style="border-radius:9999px;display:block;margin:auto;" class="logo">
	</div>
	<div class="content">
	<h1>Sanction de votre compte</h1>
	<p>Bonjour,</p>
	<p>Votre compte a été <span class="highlight">sanctionné pour spam</span> sur Tool Center.</p>
	<p>Votre nouveau statut est&nbsp;: <span class="highlight">%s</span></p>
	<div class="divider"></div>
	<p>Si vous pensez qu'il s'agit d'une erreur, veuillez contacter le support.</p>
	<div class="security-note">
		<strong>Conseil de sécurité&nbsp;:</strong> Ne partagez jamais vos informations de connexion. Tool Center ne vous demandera jamais vos identifiants par email.
	</div>
	</div>
	<div class="footer">
	<p>© Tool Center %d. Tous droits réservés.</p>
	</div>
</div>
</body>
</html>`, newStatus, year)
		_ = QueueEmail(db, email, "Sanction ToolCenter", body)
	}
}

func recordSanction(db *sql.DB, userID, prevStatus, reason string, end *time.Time) {
	if end != nil {
		_, _ = db.Exec(`INSERT INTO moderation_actions (user_id, action_type, reason, previous_status, start_date, end_date) VALUES (?, 'Ban', ?, ?, NOW(), ?)`, userID, reason, prevStatus, *end)
	} else {
		_, _ = db.Exec(`INSERT INTO moderation_actions (user_id, action_type, reason, previous_status, start_date) VALUES (?, 'Warn', ?, ?, NOW())`, userID, reason, prevStatus)
	}
}

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
			if cfg.AntiSpam.ProxyMultiplier > 1 {
				interval /= time.Duration(cfg.AntiSpam.ProxyMultiplier)
			} else {
				interval /= 2
			}
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
				m := cfg.AntiSpam.ProxyMultiplier
				if m == 0 {
					m = 2
				}
				block *= time.Duration(m)
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
