package user

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"time"

	"toolcenter/config"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func newToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func VerifyEmailHandler(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Token manquant."})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		log.Println("DB open:", err)
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
	err = db.QueryRow(query, token).Scan(&userID, &expiresAt, &verifiedAt)
	switch {
	case err == sql.ErrNoRows:
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Token invalide."})
		return
	case err != nil:
		log.Println("Scan:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erreur interne."})
		return
	}

	if verifiedAt.Valid {
		redirectWithMsg(c, "Compte déjà vérifié.", "")
		return
	}

	if time.Now().After(expiresAt) {
		db.Exec("DELETE FROM email_verification_tokens WHERE token = ?", token)
		redirectWithMsg(c, "Token expiré.", "")
		return
	}

	sessionToken := newToken(64)
	sessionExpire := time.Now().Add(30 * 24 * time.Hour)
	deviceInfo := c.Request.UserAgent()
	ip := c.ClientIP()

	tx, err := db.Begin()
	if err != nil {
		log.Println("Begin:", err)
		redirectWithMsg(c, "Erreur interne.", "")
		return
	}

	_, err = tx.Exec(`UPDATE users SET email_verified_at = NOW(), last_login = NOW() WHERE user_id = ?`, userID)
	if err != nil {
		tx.Rollback()
		log.Println("Update user:", err)
		redirectWithMsg(c, "Erreur interne.", "")
		return
	}

	_, err = tx.Exec(`DELETE FROM email_verification_tokens WHERE token = ?`, token)
	if err != nil {
		tx.Rollback()
		log.Println("Delete token:", err)
		redirectWithMsg(c, "Erreur interne.", "")
		return
	}

	_, err = tx.Exec(`
		INSERT INTO user_tokens (user_id, token, device_info, expires_at)
		VALUES (?, ?, ?, ?)`,
		userID, sessionToken, deviceInfo+" | "+ip, sessionExpire)
	if err != nil {
		tx.Rollback()
		log.Println("Insert session token:", err)
		redirectWithMsg(c, "Erreur interne.", "")
		return
	}

	tx.Commit()

	redirectURL := fmt.Sprintf(
		"https://tool-center.fr/account/?event=email_verified&token=%s",
		sessionToken,
	)
	c.Redirect(http.StatusFound, redirectURL)
}

func redirectWithMsg(c *gin.Context, msg, token string) {
	html := fmt.Sprintf(`<html><head><meta charset="utf-8"><title>Tool Center</title></head>
<body><p>%s</p></body></html>`, msg)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
