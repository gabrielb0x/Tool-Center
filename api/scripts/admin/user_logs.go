package admin

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"toolcenter/config"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// UserLogsHandler returns activity logs for a specific user
func UserLogsHandler(c *gin.Context) {
	uid := c.Param("id")
	if uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "id manquant"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit := 50
	offset := (page - 1) * limit

	db, err := config.OpenDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	rows, err := db.Query(`
        SELECT al.created_at, al.action, al.message, al.ip_address, al.success,
               u.user_id, u.username, u.avatar_url
        FROM activity_logs al
        LEFT JOIN users u ON u.user_id = al.user_id
        WHERE al.user_id = ?
        ORDER BY al.created_at DESC
        LIMIT ? OFFSET ?`, uid, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer rows.Close()

	logs := make([]gin.H, 0)
	for rows.Next() {
		var ts time.Time
		var action, message, ip string
		var success bool
		var userID, username, avatar sql.NullString
		if err := rows.Scan(&ts, &action, &message, &ip, &success, &userID, &username, &avatar); err != nil {
			continue
		}
		entry := gin.H{
			"timestamp": ts,
			"action":    action,
			"details":   message,
			"ipAddress": ip,
			"success":   success,
		}
		if userID.Valid {
			user := gin.H{"user_id": userID.String, "username": username.String}
			if avatar.Valid {
				user["avatar_url"] = avatar.String
			}
			entry["user"] = user
		}
		logs = append(logs, entry)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "logs": logs})
}
