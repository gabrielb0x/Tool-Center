package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"time"

	"toolcenter/config"
	"toolcenter/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil ||
		req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Requête invalide.",
		})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
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
	)
	err = db.QueryRow(`SELECT user_id, password_hash FROM users WHERE email = ?`, req.Email).
		Scan(&uid, &storedHash)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Utilisateur ou mot de passe incorrect.",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erreur DB."})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Utilisateur ou mot de passe incorrect.",
		})
		return
	}

	_ = c.Request.Body.Close()
	c.Request.Body = http.NoBody
	c.Set("user_id_override", uid)

	if _, _, _, err = utils.Check(c, utils.CheckOpts{
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
		return
	}

	b := make([]byte, 64)
	_, _ = rand.Read(b)
	token := hex.EncodeToString(b)

	_, err = db.Exec(`
INSERT INTO user_tokens (user_id, token, device_info, expires_at)
VALUES (?, ?, ?, ?)`,
		uid, token,
		c.Request.UserAgent()+" | "+c.ClientIP(),
		time.Now().Add(30*24*time.Hour),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Erreur lors de l'enregistrement de la session.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   token,
	})
}
