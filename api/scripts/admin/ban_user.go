package admin

import (
	"net/http"

	"toolcenter/config"
	"toolcenter/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type banRequest struct {
	Reason string `json:"reason"`
}

func BanUserHandler(c *gin.Context) {
	targetID := c.Param("id")
	if targetID == "" {
		utils.LogActivity(c, "", "ban_user", false, "id missing")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "id manquant"})
		return
	}

	var req banRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Reason == "" {
		utils.LogActivity(c, "", "ban_user", false, "reason missing")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "raison manquante"})
		return
	}

	moderatorID := c.GetString("user_id")
	if moderatorID == targetID {
		utils.LogActivity(c, moderatorID, "ban_user", false, "self ban")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Impossible de se bannir soi-même"})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, moderatorID, "ban_user", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	_, err = db.Exec(`UPDATE users SET account_status = 'Banned' WHERE user_id = ?`, targetID)
	if err != nil {
		utils.LogActivity(c, moderatorID, "ban_user", false, "update error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	_, _ = db.Exec(`INSERT INTO moderation_actions (moderator_id, user_id, action_type, reason) VALUES (?, ?, 'Ban', ?)`, moderatorID, targetID, req.Reason)
	utils.LogActivity(c, moderatorID, "ban_user", true, "")
	c.JSON(http.StatusOK, gin.H{"success": true})
}
