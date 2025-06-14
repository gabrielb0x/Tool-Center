package user

import (
	"net/http"
	"strings"
	"time"

	"toolcenter/config"
	"toolcenter/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type updateUsernameRequest struct {
	Username string `json:"username"`
}

func UpdateUsernameHandler(c *gin.Context) {
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
		utils.LogActivity(c, uid, "update_username", false, err.Error())
		c.JSON(code, gin.H{"success": false, "message": err.Error()})
		return
	}

	var req updateUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogActivity(c, uid, "update_username", false, "bad request")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides"})
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if len(req.Username) < 3 || len(req.Username) > 50 {
		utils.LogActivity(c, uid, "update_username", false, "invalid username")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Pseudo invalide"})
		return
	}

	cfg := config.Get()
	cooldown := time.Duration(cfg.Limits.UsernameChangeHours) * time.Hour
	if cooldown == 0 {
		cooldown = 24 * time.Hour
	}

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, uid, "update_username", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	var lastChangedAt time.Time
	_ = db.QueryRow(`SELECT username_changed_at FROM users WHERE user_id = ?`, uid).Scan(&lastChangedAt)
	if !lastChangedAt.IsZero() && time.Since(lastChangedAt) < cooldown {
		retryAt := lastChangedAt.Add(cooldown)
		utils.LogActivity(c, uid, "update_username", false, "cooldown")
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success":  false,
			"message":  "Vous ne pouvez changer votre pseudo qu'une fois par jour.",
			"retry_at": retryAt.Format(time.RFC3339),
		})
		return
	}

	var exists int
	_ = db.QueryRow(`SELECT COUNT(*) FROM users WHERE username = ?`, req.Username).Scan(&exists)
	if exists > 0 {
		utils.LogActivity(c, uid, "update_username", false, "taken")
		c.JSON(http.StatusConflict, gin.H{"success": false, "message": "Pseudo déjà pris"})
		return
	}

	_, err = db.Exec(`UPDATE users SET username = ?, username_changed_at = NOW() WHERE user_id = ?`, req.Username, uid)
	if err != nil {
		utils.LogActivity(c, uid, "update_username", false, "update error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	utils.LogActivity(c, uid, "update_username", true, "")
	c.JSON(http.StatusOK, gin.H{"success": true})
}
