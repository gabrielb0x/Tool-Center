package admin

import (
	"net/http"

	"toolcenter/config"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type banRequest struct {
	Reason string `json:"reason"`
}

func BanUserHandler(c *gin.Context) {
	targetID := c.Param("id")
	if targetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "id manquant"})
		return
	}

	var req banRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "raison manquante"})
		return
	}

	moderatorID, _ := c.Get("user_id")

	db, err := config.OpenDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	_, err = db.Exec(`UPDATE users SET account_status = 'Banned' WHERE user_id = ?`, targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	_, _ = db.Exec(`INSERT INTO moderation_actions (moderator_id, user_id, action_type, reason) VALUES (?, ?, 'Ban', ?)`, moderatorID, targetID, req.Reason)

	c.JSON(http.StatusOK, gin.H{"success": true})
}
