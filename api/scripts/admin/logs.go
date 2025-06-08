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

func LogsHandler(c *gin.Context) {
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
        SELECT al.created_at, al.event_type, al.target_resource, al.payload,
               al.actor_user_id, u.username, u.avatar_url
        FROM audit_logs al
        LEFT JOIN users u ON u.user_id = al.actor_user_id
        ORDER BY al.created_at DESC
        LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer rows.Close()

	logs := make([]gin.H, 0)
	for rows.Next() {
		var ts time.Time
		var eventType, targetRes string
		var payload sql.NullString
		var actorID, username, avatar sql.NullString
		if err := rows.Scan(&ts, &eventType, &targetRes, &payload, &actorID, &username, &avatar); err != nil {
			continue
		}
		entry := gin.H{
			"timestamp": ts,
			"action":    eventType,
			"details":   targetRes,
		}
		if payload.Valid && payload.String != "" {
			entry["details"] = payload.String
		}
		if actorID.Valid {
			user := gin.H{"user_id": actorID.String, "username": username.String}
			if avatar.Valid {
				user["avatar"] = avatar.String
			}
			entry["user"] = user
		}
		logs = append(logs, entry)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "logs": logs})
}
