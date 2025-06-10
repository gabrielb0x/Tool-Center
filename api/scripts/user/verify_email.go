package user

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"time"

	"toolcenter/config"
	"toolcenter/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func newToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func hashToken(t string) string {
	sum := sha256.Sum256([]byte(t))
	return hex.EncodeToString(sum[:])
}

func VerifyEmailHandler(c *gin.Context) {
	token := c.Query("token")
	tokenHash := hashToken(token)
	if token == "" {
		utils.LogActivity(c, "", "verify_email", false, "token missing")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Token manquant."})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		log.Println("DB open:", err)
		utils.LogActivity(c, "", "verify_email", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erreur DB."})
		return
	}
	defer db.Close()

	var (
		userID     string
		expiresAt  time.Time
		verifiedAt sql.NullTime
	)
	query := `
SELECT u.user_id, evt.expires_at, u.email_verified_at
FROM email_verification_tokens evt
JOIN users u ON u.user_id = evt.user_id
WHERE evt.token = ? LIMIT 1`
	err = db.QueryRow(query, tokenHash).Scan(&userID, &expiresAt, &verifiedAt)
	switch {
	case err == sql.ErrNoRows:
		utils.LogActivity(c, "", "verify_email", false, "invalid token")
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Token invalide."})
		return
	case err != nil:
		log.Println("Scan:", err)
		utils.LogActivity(c, "", "verify_email", false, "scan error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erreur interne."})
		return
	}

	if verifiedAt.Valid {
		utils.LogActivity(c, userID, "verify_email", false, "already verified")
		redirectWithMsg(c, "Compte déjà vérifié.", "")
		return
	}

	if time.Now().After(expiresAt) {
		db.Exec("DELETE FROM email_verification_tokens WHERE token = ?", tokenHash)
		utils.LogActivity(c, userID, "verify_email", false, "token expired")
		redirectWithMsg(c, "Token expiré.", "")
		return
	}

	sessionToken := newToken(64)
	sessionTokenHash := hashToken(sessionToken)
	sessionExpire := time.Now().Add(30 * 24 * time.Hour)
	deviceInfo := c.Request.UserAgent()
	ip := c.ClientIP()

	tx, err := db.Begin()
	if err != nil {
		log.Println("Begin:", err)
		utils.LogActivity(c, userID, "verify_email", false, "tx begin error")
		redirectWithMsg(c, "Erreur interne.", "")
		return
	}

	_, err = tx.Exec(`UPDATE users SET email_verified_at = NOW(), last_login = NOW() WHERE user_id = ?`, userID)
	if err != nil {
		tx.Rollback()
		log.Println("Update user:", err)
		utils.LogActivity(c, userID, "verify_email", false, "update user error")
		redirectWithMsg(c, "Erreur interne.", "")
		return
	}

	_, err = tx.Exec(`DELETE FROM email_verification_tokens WHERE token = ?`, tokenHash)
	if err != nil {
		tx.Rollback()
		log.Println("Delete token:", err)
		utils.LogActivity(c, userID, "verify_email", false, "delete token error")
		redirectWithMsg(c, "Erreur interne.", "")
		return
	}

	_, err = tx.Exec(`
               INSERT INTO user_tokens (user_id, token, device_info, expires_at)
               VALUES (?, ?, ?, ?)`,
		userID, sessionTokenHash, deviceInfo+" | "+ip, sessionExpire)
	if err != nil {
		tx.Rollback()
		log.Println("Insert session token:", err)
		utils.LogActivity(c, userID, "verify_email", false, "insert session error")
		redirectWithMsg(c, "Erreur interne.", "")
		return
	}

	tx.Commit()

	utils.LogActivity(c, userID, "verify_email", true, "")

	redirectURL := fmt.Sprintf(
		"https://tool-center.fr/account/?event=email_verified&token=%s",
		sessionToken,
	)
	c.Redirect(http.StatusFound, redirectURL)
}

func redirectWithMsg(c *gin.Context, msg, _ string) {
	html := fmt.Sprintf(`<html><head><meta charset="utf-8"><title>Tool Center</title></head>
<body><p>%s</p></body></html>`, msg)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
