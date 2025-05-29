package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"crypto/sha256"
	"toolcenter/config"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func validUsername(u string) bool {
	return len(u) >= 3 && len(u) <= 50 && !strings.ContainsAny(u, " ")
}

func validEmail(e string) bool {
	return strings.Contains(e, "@") && strings.Contains(e, ".")
}

func randomToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func hashToken(t string) string {
	sum := sha256.Sum256([]byte(t))
	return hex.EncodeToString(sum[:])
}

func RegisterHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Requête invalide."})
		return
	}

	ip := c.ClientIP()

	if !validUsername(req.Username) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Nom d'utilisateur invalide."})
		return
	}
	if !validEmail(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Adresse email invalide."})
		return
	}
	if len(req.Password) < 7 || len(req.Password) > 30 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Le mot de passe doit faire entre 7 et 30 caractères."})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erreur DB."})
		return
	}
	defer db.Close()

	var ipCount int
	db.QueryRow("SELECT COUNT(*) FROM users WHERE ip_address = ?", ip).Scan(&ipCount)
	if ipCount > 0 {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "Compte déjà créé."})
		return
	}

	var dup int
	db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? OR email = ?", req.Username, req.Email).Scan(&dup)
	if dup > 0 {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "Nom d'utilisateur ou email déjà pris."})
		return
	}

	pwHash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	verifyToken := randomToken(32)
	verifyTokenHash := hashToken(verifyToken)
	expires := time.Now().Add(10 * time.Minute)

	// ===== génération de l’UUID v7 =====
	uuidV7, _ := uuid.NewV7() // (retourne uuid.UUID, error)
	userID := uuidV7.String() // forme texte 36 caractères

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erreur interne."})
		return
	}

	_, err = tx.Exec(`
		INSERT INTO users (user_id, username, email, password_hash, ip_address)
		VALUES (?,?,?,?,?)`,
		userID, req.Username, req.Email, pwHash, ip)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erreur création utilisateur."})
		return
	}

	_, err = tx.Exec(`
        INSERT INTO email_verification_tokens (user_id, token, expires_at)
        VALUES (?,?,?)`,
		userID, verifyTokenHash, expires)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erreur création token."})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erreur commit."})
		return
	}

	go sendVerificationEmail(req.Email, req.Username, verifyToken)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Inscription réussie. Vérifie ton email."})
}

func sendVerificationEmail(email, username, token string) {
	cfg := config.Get()
	verifyURL := fmt.Sprintf("https://api.tool-center.fr/%s/user/verify_email?token=%s", cfg.URLVersion, token)

	body := fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="fr">
	<head>
	<meta charset="UTF-8">
	<title>Vérification de votre compte Tool Center</title>
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
		<h1>Vérification de votre compte</h1>
		<p>Bonjour <span class="highlight">%s</span>,</p>
		<p>Nous sommes ravis de vous accueillir sur Tool Center ! Pour finaliser votre inscription et accéder à toutes les fonctionnalités, veuillez vérifier votre adresse email en cliquant sur le bouton ci-dessous :</p>
		<div style="text-align:center;">
			<a href="%s" class="button">Vérifier mon compte</a>
			<p class="expire-note">Ce lien expirera dans 10 minutes</p>
		</div>
		<div class="divider"></div>
		<p>Si vous n'avez pas créé de compte sur Tool Center, vous pouvez ignorer cet email en toute sécurité.</p>
		<div class="security-note">
			<strong>Conseil de sécurité&nbsp;:</strong> Ne partagez jamais ce lien avec qui que ce soit. Tool Center ne vous demandera jamais vos informations de connexion par email.
		</div>
		</div>
		<div class="footer">
		<p>© Tool Center %d. Tous droits réservés.</p>
		</div>
	</div>
	</body>
	</html>
	`, username, verifyURL, time.Now().Year())

	msg := gomail.NewMessage()
	msg.SetHeader("From", cfg.Email.From)
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", "Vérification de votre compte")
	msg.SetBody("text/html", body)

	d := gomail.NewDialer(cfg.Email.Host, cfg.Email.Port, cfg.Email.Username, cfg.Email.Password)
	d.SSL = true
	_ = d.DialAndSend(msg)
}
