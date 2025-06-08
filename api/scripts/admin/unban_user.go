package admin

import (
	"net/http"

	"toolcenter/config"

	"github.com/gin-gonic/gin"
)

func UnbanUserHandler(c *gin.Context) {
	uid := c.Param("id")
	if uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "id manquant"})
		return
	}

	moderatorID, _ := c.Get("user_id")

	db, err := config.OpenDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	_, err = db.Exec(`UPDATE users SET account_status = 'Good' WHERE user_id = ?`, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	_, _ = db.Exec(`INSERT INTO moderation_actions (moderator_id, user_id, action_type) VALUES (?, ?, 'Unban')`, moderatorID, uid)

	c.JSON(http.StatusOK, gin.H{"success": true})
}
