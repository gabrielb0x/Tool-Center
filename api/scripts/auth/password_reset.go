package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"toolcenter/config"
	"toolcenter/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
)

type resetRequest struct {
	Email          string `json:"email"`
	TurnstileToken string `json:"turnstile_token"`
}

type resetConfirmRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func randToken(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func hashTok(t string) string {
	h := sha256.Sum256([]byte(t))
	return hex.EncodeToString(h[:])
}

func RequestPasswordResetHandler(c *gin.Context) {
	var req resetRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" || req.TurnstileToken == "" {
		utils.LogActivity(c, "", "pwd_reset_request", false, "bad request")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Requête invalide."})
		return
	}
	ok, err := utils.VerifyTurnstile(req.TurnstileToken, config.Get().Turnstile.SignInSecret, c.ClientIP())
	if err != nil || !ok {
		utils.LogActivity(c, "", "pwd_reset_request", false, "captcha invalid")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Captcha invalide"})
		return
	}
	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, "", "pwd_reset_request", false, "db error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erreur DB."})
		return
	}
	defer db.Close()

	var uid, username string
	err = db.QueryRow("SELECT user_id, username FROM users WHERE email=?", req.Email).Scan(&uid, &username)
	if err == sql.ErrNoRows {
		// avoid enumeration
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}
	if err != nil {
		utils.LogActivity(c, "", "pwd_reset_request", false, "query error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erreur interne."})
		return
	}

	token := randToken(32)
	tokenHash := hashTok(token)
	expires := time.Now().Add(time.Duration(config.Get().PasswordReset.TokenExpiryMinutes) * time.Minute)
	_, err = db.Exec(`INSERT INTO password_resets (user_id, token, expires_at) VALUES (?,?,?)`, uid, tokenHash, expires)
	if err != nil {
		utils.LogActivity(c, uid, "pwd_reset_request", false, "insert error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	go sendPasswordResetEmail(req.Email, username, token)
	utils.LogActivity(c, uid, "pwd_reset_request", true, "")
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func sendPasswordResetEmail(email, username, token string) {
	cfg := config.Get()
	link := fmt.Sprintf("%s/reset.html?token=%s", cfg.URLweb, token)
	body := fmt.Sprintf(`<!DOCTYPE html>
<html lang="fr">
<head>
<meta charset="UTF-8">
<title>Réinitialisation de votre mot de passe Tool Center</title>
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
    <h1>Réinitialisation de votre mot de passe</h1>
    <p>Bonjour <span class="highlight">%s</span>,</p>
    <p>Vous avez demandé à réinitialiser votre mot de passe Tool Center.<br>
    Cliquez sur le bouton ci-dessous pour choisir un nouveau mot de passe :</p>
    <div style="text-align:center;">
        <a href="%s" class="button">Réinitialiser mon mot de passe</a>
        <p class="expire-note">Ce lien expirera dans %d minutes</p>
    </div>
    <div class="divider"></div>
    <p>Si vous n'êtes pas à l'origine de cette demande, vous pouvez ignorer cet email en toute sécurité.</p>
    <div class="security-note">
        <strong>Conseil de sécurité&nbsp;:</strong> Ne partagez jamais ce lien avec qui que ce soit. Tool Center ne vous demandera jamais vos informations de connexion par email.
    </div>
    </div>
    <div class="footer">
    <p>© Tool Center %d. Tous droits réservés.</p>
    </div>
</div>
</body>
</html>`, username, link, cfg.PasswordReset.TokenExpiryMinutes, time.Now().Year())
	msg := gomail.NewMessage()
	msg.SetHeader("From", cfg.Email.From)
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", "Réinitialisation de votre mot de passe")
	msg.SetBody("text/html", body)
	d := gomail.NewDialer(cfg.Email.Host, cfg.Email.Port, cfg.Email.Username, cfg.Email.Password)
	d.SSL = true
	_ = d.DialAndSend(msg)
}

func ResetPasswordHandler(c *gin.Context) {
	var req resetConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Token == "" || len(req.NewPassword) < 7 {
		utils.LogActivity(c, "", "pwd_reset", false, "bad request")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Requête invalide."})
		return
	}
	tokenHash := hashTok(req.Token)
	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, "", "pwd_reset", false, "db error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	var uid string
	var expires time.Time
	err = db.QueryRow(`SELECT user_id, expires_at FROM password_resets WHERE token=?`, tokenHash).Scan(&uid, &expires)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Token invalide"})
		return
	}
	if err != nil {
		utils.LogActivity(c, "", "pwd_reset", false, "query error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	if time.Now().After(expires) {
		db.Exec(`DELETE FROM password_resets WHERE token=?`, tokenHash)
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Token expiré"})
		return
	}
	pwHash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	_, err = db.Exec(`UPDATE users SET password_hash=? WHERE user_id=?`, pwHash, uid)
	if err != nil {
		utils.LogActivity(c, uid, "pwd_reset", false, "update error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	db.Exec(`DELETE FROM password_resets WHERE user_id=?`, uid)
	utils.LogActivity(c, uid, "pwd_reset", true, "")
	c.JSON(http.StatusOK, gin.H{"success": true})
}
