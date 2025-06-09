package utils

import (
	"database/sql"

	"toolcenter/config"

	"github.com/gin-gonic/gin"
)

// LogActivity stores an action made by a user or guest in the database.
// If userID is empty, the value is stored as NULL.
func LogActivity(c *gin.Context, userID, action string, success bool, message string) {
	db, err := config.OpenDB()
	if err != nil {
		return
	}
	defer db.Close()

	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	var uid interface{}
	if userID == "" {
		uid = sql.NullString{}
	} else {
		uid = userID
	}

	_, _ = db.Exec(`INSERT INTO activity_logs (user_id, ip_address, user_agent, action, success, message) VALUES (?, ?, ?, ?, ?, ?)`,
		uid, ip, ua, action, success, message)
}
