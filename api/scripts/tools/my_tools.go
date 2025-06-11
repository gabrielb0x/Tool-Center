package tools

import (
	"net/http"
	"time"

	"toolcenter/config"
	"toolcenter/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type Tool struct {
	ID           string    `json:"tool_id"`
	UserID       string    `json:"user_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	ContentURL   string    `json:"content_url"`
	ThumbnailURL string    `json:"thumbnail_url"`
	Status       string    `json:"status"`
	Views        int       `json:"views"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func MyToolsHandler(c *gin.Context) {
	uid, _, _, _, err := utils.Check(c, utils.CheckOpts{
		RequireToken:     true,
		RequireVerified:  true,
		RequireNotBanned: true,
		UpdateLastLogin:  true,
	})
	if err != nil {
		utils.LogActivity(c, uid, "my_tools", false, err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": err.Error()})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, uid, "my_tools", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erreur DB."})
		return
	}
	defer db.Close()

	rows, err := db.Query(`SELECT tool_id, user_id, title, description, content_url, thumbnail_url, status, views, created_at, updated_at FROM tools WHERE user_id = ? ORDER BY created_at DESC`, uid)
	if err != nil {
		utils.LogActivity(c, uid, "my_tools", false, "query error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erreur DB."})
		return
	}
	defer rows.Close()

	tools := make([]Tool, 0)
	for rows.Next() {
		var t Tool
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.ContentURL, &t.ThumbnailURL, &t.Status, &t.Views, &t.CreatedAt, &t.UpdatedAt); err != nil {
			utils.LogActivity(c, uid, "my_tools", false, "scan error")
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erreur DB."})
			return
		}
		tools = append(tools, t)
	}

	utils.LogActivity(c, uid, "my_tools", true, "")
	c.JSON(http.StatusOK, gin.H{"success": true, "tools": tools})
}
