package admin

import (
	"net/http"

	"toolcenter/config"
	"toolcenter/utils"

	"github.com/gin-gonic/gin"
)

func UnbanUserHandler(c *gin.Context) {
	uid := c.Param("id")
	if uid == "" {
		utils.LogActivity(c, "", "unban_user", false, "id manquant")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "id manquant"})
		return
	}

	moderatorID, _ := c.Get("user_id")

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, moderatorID.(string), "unban_user", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	_, err = db.Exec(`UPDATE users SET account_status = 'Good' WHERE user_id = ?`, uid)
	if err != nil {
		utils.LogActivity(c, moderatorID.(string), "unban_user", false, "update error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	_, _ = db.Exec(`INSERT INTO moderation_actions (moderator_id, user_id, action_type) VALUES (?, ?, 'Unban')`, moderatorID, uid)
	utils.LogActivity(c, moderatorID.(string), "unban_user", true, "")
	c.JSON(http.StatusOK, gin.H{"success": true})
}
