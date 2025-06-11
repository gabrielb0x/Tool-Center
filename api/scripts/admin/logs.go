package admin

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"toolcenter/config"
	"toolcenter/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

const logsPerPage = 50

func LogsHandler(c *gin.Context) {
	adminID, _ := c.Get("user_id")
	uid, _ := adminID.(string)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit := logsPerPage
	offset := (page - 1) * limit

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, uid, "view_logs", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	rows, err := db.Query(`
        SELECT al.created_at, al.action, al.message, al.success, al.ip_address,
               al.user_id, u.username, u.avatar_url
        FROM activity_logs al
        LEFT JOIN users u ON u.user_id = al.user_id
        ORDER BY al.created_at DESC
        LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		utils.LogActivity(c, uid, "view_logs", false, "query error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer rows.Close()

	logs := make([]gin.H, 0)
	for rows.Next() {
		var ts time.Time
		var action, message string
		var success bool
		var ip sql.NullString
		var userID, username, avatar sql.NullString
		if err := rows.Scan(&ts, &action, &message, &success, &ip, &userID, &username, &avatar); err != nil {
			continue
		}
		entry := gin.H{
			"timestamp": ts,
			"action":    action,
			"success":   success,
		}
		if message != "" {
			entry["details"] = message
		}
		if ip.Valid {
			entry["ip"] = ip.String
		}
		if userID.Valid {
			user := gin.H{"user_id": userID.String, "username": username.String}
			if avatar.Valid {
				user["avatar"] = avatar.String
			}
			entry["user"] = user
		}
		logs = append(logs, entry)
	}

	utils.LogActivity(c, uid, "view_logs", true, "")

	c.JSON(http.StatusOK, gin.H{"success": true, "logs": logs})
}
