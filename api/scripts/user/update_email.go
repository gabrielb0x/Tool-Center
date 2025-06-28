package user

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"toolcenter/config"
	"toolcenter/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

type updateEmailRequest struct {
	NewEmail        string `json:"new_email"`
	CurrentPassword string `json:"current_password"`
	TwoFactorCode   string `json:"two_factor_code"`
}

func UpdateEmailHandler(c *gin.Context) {
	uid, _, _, _, err := utils.Check(c, utils.CheckOpts{
		RequireToken:     true,
		RequireVerified:  true,
		RequireNotBanned: true,
	})
	if err != nil {
		code := http.StatusInternalServerError
		switch err {
		case utils.ErrMissingToken, utils.ErrInvalidToken, utils.ErrExpiredToken:
			code = http.StatusUnauthorized
		case utils.ErrEmailNotVerified, utils.ErrAccountBanned:
			code = http.StatusForbidden
		}
		utils.LogActivity(c, uid, "update_email", false, err.Error())
		c.JSON(code, gin.H{"success": false, "message": err.Error()})
		return
	}

	var req updateEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogActivity(c, uid, "update_email", false, "bad request")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides"})
		return
	}
	req.NewEmail = strings.TrimSpace(req.NewEmail)
	if !strings.Contains(req.NewEmail, "@") || len(req.NewEmail) < 6 {
		utils.LogActivity(c, uid, "update_email", false, "invalid email")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Email invalide"})
		return
	}
	if req.CurrentPassword == "" {
		utils.LogActivity(c, uid, "update_email", false, "password missing")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Mot de passe requis"})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, uid, "update_email", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	var lastChangedAt time.Time
	_ = db.QueryRow(`SELECT email_changed_at FROM users WHERE user_id = ?`, uid).Scan(&lastChangedAt)
	cooldown := time.Duration(config.Get().Cooldowns.EmailChangeDays) * 24 * time.Hour
	if !lastChangedAt.IsZero() && time.Since(lastChangedAt) < cooldown {
		retryAt := lastChangedAt.Add(cooldown)
		utils.LogActivity(c, uid, "update_email", false, "cooldown")
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success":  false,
			"message":  "Vous ne pouvez changer votre email qu'une fois toutes les 30 jours.",
			"retry_at": retryAt.Format(time.RFC3339),
		})
		return
	}

	var (
		hash   string
		secret sql.NullString
	)
	err = db.QueryRow(`SELECT password_hash, authenticator_secret FROM users WHERE user_id = ?`, uid).Scan(&hash, &secret)
	if err != nil {
		utils.LogActivity(c, uid, "update_email", false, "select error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.CurrentPassword)) != nil {
		utils.LogActivity(c, uid, "update_email", false, "wrong password")
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Mot de passe incorrect"})
		return
	}

	if secret.Valid && secret.String != "" {
		if len(req.TwoFactorCode) != 6 || !totp.Validate(req.TwoFactorCode, secret.String) {
			utils.LogActivity(c, uid, "update_email", false, "invalid 2fa code")
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Code 2FA invalide"})
			return
		}
	}

	var exists int
	_ = db.QueryRow(`SELECT COUNT(*) FROM users WHERE email = ?`, req.NewEmail).Scan(&exists)
	if exists > 0 {
		utils.LogActivity(c, uid, "update_email", false, "email exists")
		c.JSON(http.StatusConflict, gin.H{"success": false, "message": "Email déjà utilisé"})
		return
	}

	_, err = db.Exec(`UPDATE users SET email = ?, email_changed_at = NOW(), email_verified_at = NULL WHERE user_id = ?`, req.NewEmail, uid)
	if err != nil {
		utils.LogActivity(c, uid, "update_email", false, "update error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	utils.LogActivity(c, uid, "update_email", true, "")
	c.JSON(http.StatusOK, gin.H{"success": true})
}
