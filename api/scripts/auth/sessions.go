package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"toolcenter/config"
	"toolcenter/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mssola/user_agent"
)

type sessionInfo struct {
	ID         int64     `json:"id"`
	DeviceType string    `json:"device_type"`
	OS         string    `json:"os"`
	Browser    string    `json:"browser"`
	Location   string    `json:"location"`
	CreatedAt  time.Time `json:"created_at"`
}

func currentTokenHash(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	raw := strings.TrimPrefix(header, "Bearer ")
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func GetSessionsHandler(c *gin.Context) {
	uid, _, _, _, err := utils.Check(c, utils.CheckOpts{RequireToken: true, RequireVerified: true, RequireNotBanned: true})
	if err != nil {
		code := http.StatusInternalServerError
		switch err {
		case utils.ErrMissingToken, utils.ErrInvalidToken, utils.ErrExpiredToken:
			code = http.StatusUnauthorized
		case utils.ErrEmailNotVerified, utils.ErrAccountBanned:
			code = http.StatusForbidden
		}
		c.JSON(code, gin.H{"success": false, "message": err.Error()})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	rows, err := db.Query(`SELECT token_id, token, device_info, created_at FROM user_tokens WHERE user_id=? ORDER BY created_at DESC`, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer rows.Close()

	sessions := make([]sessionInfo, 0)
	for rows.Next() {
		var id int64
		var tokenHash, deviceInfo string
		var createdAt time.Time
		if err := rows.Scan(&id, &tokenHash, &deviceInfo, &createdAt); err != nil {
			continue
		}
		parts := strings.SplitN(deviceInfo, " | ", 2)
		uaStr := parts[0]
		ua := user_agent.New(uaStr)
		browser, _ := ua.Browser()
		deviceType := "desktop"
		if ua.Mobile() {
			deviceType = "mobile"
		}
		sessions = append(sessions, sessionInfo{
			ID:         id,
			DeviceType: deviceType,
			OS:         ua.OS(),
			Browser:    browser,
			Location:   "",
			CreatedAt:  createdAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "sessions": sessions})
}

func DeleteSessionHandler(c *gin.Context) {
	uid, _, _, _, err := utils.Check(c, utils.CheckOpts{RequireToken: true, RequireVerified: true, RequireNotBanned: true})
	if err != nil {
		code := http.StatusInternalServerError
		switch err {
		case utils.ErrMissingToken, utils.ErrInvalidToken, utils.ErrExpiredToken:
			code = http.StatusUnauthorized
		case utils.ErrEmailNotVerified, utils.ErrAccountBanned:
			code = http.StatusForbidden
		}
		c.JSON(code, gin.H{"success": false, "message": err.Error()})
		return
	}

	id := c.Param("id")
	db, err := config.OpenDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	res, err := db.Exec(`DELETE FROM user_tokens WHERE token_id=? AND user_id=?`, id, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Session inconnue"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func DeleteAllSessionsHandler(c *gin.Context) {
	uid, _, _, _, err := utils.Check(c, utils.CheckOpts{RequireToken: true, RequireVerified: true, RequireNotBanned: true})
	if err != nil {
		code := http.StatusInternalServerError
		switch err {
		case utils.ErrMissingToken, utils.ErrInvalidToken, utils.ErrExpiredToken:
			code = http.StatusUnauthorized
		case utils.ErrEmailNotVerified, utils.ErrAccountBanned:
			code = http.StatusForbidden
		}
		c.JSON(code, gin.H{"success": false, "message": err.Error()})
		return
	}

	currentHash := currentTokenHash(c)

	db, err := config.OpenDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	if currentHash != "" {
		_, _ = db.Exec(`DELETE FROM user_tokens WHERE user_id=? AND token<>?`, uid, currentHash)
	} else {
		_, _ = db.Exec(`DELETE FROM user_tokens WHERE user_id=?`, uid)
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
