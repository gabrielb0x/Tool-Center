package admin

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"toolcenter/config"
	"toolcenter/utils"

	"github.com/gin-gonic/gin"
)

const userActivityPerPage = 20

func UserActivityHandler(c *gin.Context) {
	adminID, _ := c.Get("user_id")
	modID, _ := adminID.(string)

	uid := c.Param("id")
	if uid == "" {
		utils.LogActivity(c, modID, "user_activity", false, "id manquant")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "id manquant"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit := userActivityPerPage
	offset := (page - 1) * limit

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, modID, "user_activity", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	rows, err := db.Query(`SELECT log_id, created_at, action, success, message, ip_address
        FROM activity_logs WHERE user_id = ?
        ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		uid, limit, offset)
	if err != nil {
		utils.LogActivity(c, modID, "user_activity", false, "query error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer rows.Close()

	logs := make([]gin.H, 0)
	for rows.Next() {
		var (
			id      int
			ts      time.Time
			action  string
			success bool
			msg     sql.NullString
			ip      sql.NullString
		)
		if err := rows.Scan(&id, &ts, &action, &success, &msg, &ip); err != nil {
			continue
		}
		entry := gin.H{
			"id":        id,
			"timestamp": ts,
			"action":    action,
			"success":   success,
		}
		if msg.Valid {
			entry["message"] = msg.String
		}
		if ip.Valid {
			entry["ip"] = ip.String
		}
		logs = append(logs, entry)
	}

	utils.LogActivity(c, modID, "user_activity", true, "")

	c.JSON(http.StatusOK, gin.H{"success": true, "logs": logs})
}
