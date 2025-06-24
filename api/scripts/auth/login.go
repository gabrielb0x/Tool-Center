package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
	"time"

	"toolcenter/config"
	"toolcenter/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	TwoFactorCode  string `json:"two_factor_code"`
	TurnstileToken string `json:"turnstile_token"`
}

func LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogActivity(c, "", "login_attempt", false, "invalid request")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Requête invalide.",
		})
		return
	}
	if req.TurnstileToken == "" {
		utils.LogActivity(c, "", "login_attempt", false, "missing captcha")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Captcha manquant.",
		})
		return
	}
	if req.Email == "" || req.Password == "" {
		utils.LogActivity(c, "", "login_attempt", false, "missing fields")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Requête invalide.",
		})
		return
	}

	ok, err := utils.VerifyTurnstile(req.TurnstileToken, config.Get().Turnstile.SignInSecret, c.ClientIP())
	if err != nil || !ok {
		utils.LogActivity(c, "", "login_attempt", false, "captcha invalid")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Captcha invalide"})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, "", "login_attempt", false, "db error")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Erreur DB.",
		})
		return
	}
	defer db.Close()

	var (
		uid        string
		storedHash string
		authSecret sql.NullString
	)
	err = db.QueryRow(`SELECT user_id, password_hash, authenticator_secret FROM users WHERE email = ?`, req.Email).
		Scan(&uid, &storedHash, &authSecret)
	if err == sql.ErrNoRows {
		utils.LogActivity(c, "", "login_attempt", false, "unknown user")
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Utilisateur ou mot de passe incorrect.",
		})
		return
	}
	if err != nil {
		utils.LogActivity(c, "", "login_attempt", false, "db error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erreur DB."})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password)) != nil {
		utils.LogActivity(c, uid, "login_attempt", false, "wrong password")
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Utilisateur ou mot de passe incorrect.",
		})
		return
	}

	if authSecret.Valid && authSecret.String != "" {
		if req.TwoFactorCode == "" {
			utils.LogActivity(c, uid, "login_attempt", false, "2fa required")
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "two_factor_required": true})
			return
		}
		if !totp.Validate(req.TwoFactorCode, authSecret.String) {
			utils.LogActivity(c, uid, "login_attempt", false, "invalid 2fa")
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Code 2FA invalide"})
			return
		}
	}

	_ = c.Request.Body.Close()
	c.Request.Body = http.NoBody
	c.Set("user_id_override", uid)

	if _, _, _, _, err = utils.Check(c, utils.CheckOpts{
		RequireToken:     false,
		RequireVerified:  true,
		RequireNotBanned: true,
		UpdateLastLogin:  true,
	}); err != nil {
		code := map[error]int{
			utils.ErrEmailNotVerified: http.StatusForbidden,
			utils.ErrAccountBanned:    http.StatusForbidden,
		}[err]
		if code == 0 {
			code = http.StatusInternalServerError
		}
		c.JSON(code, gin.H{"success": false, "message": err.Error()})
		utils.LogActivity(c, uid, "login_attempt", false, err.Error())
		return
	}

	b := make([]byte, 64)
	_, _ = rand.Read(b)
	token := hex.EncodeToString(b)
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	_, err = db.Exec(`
INSERT INTO user_tokens (user_id, token, device_info, expires_at)
VALUES (?, ?, ?, ?)`,
		uid, tokenHash,
		c.Request.UserAgent()+" | "+c.ClientIP(),
		time.Now().Add(30*24*time.Hour),
	)
	if err != nil {
		utils.LogActivity(c, uid, "login", false, "session insert error")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Erreur lors de l'enregistrement de la session.",
		})
		return
	}

	utils.LogActivity(c, uid, "login", true, "")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   token,
	})
}
