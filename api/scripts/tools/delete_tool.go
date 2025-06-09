package tools

import (
	"net/http"

	"toolcenter/config"
	"toolcenter/utils"

	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func DeleteToolHandler(c *gin.Context) {
	uid, _, _, _, err := utils.Check(c, utils.CheckOpts{
		RequireToken:     true,
		RequireVerified:  true,
		RequireNotBanned: true,
	})
	if err != nil {
		code := http.StatusInternalServerError
		switch err {
		case utils.ErrMissingToken, utils.ErrInvalidToken, utils.ErrExpiredToken:
			code = http.StatusUnauthorized
		case utils.ErrEmailNotVerified, utils.ErrAccountBanned:
			code = http.StatusForbidden
		}
		utils.LogActivity(c, uid, "delete_tool", false, err.Error())
		c.JSON(code, gin.H{"success": false, "message": err.Error()})
		return
	}

	toolID := c.Param("id")
	if toolID == "" {
		utils.LogActivity(c, uid, "delete_tool", false, "id missing")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ID manquant"})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, uid, "delete_tool", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	var imagePath string
	err = db.QueryRow(`SELECT thumbnail_url FROM tools WHERE tool_id = ? AND user_id = ?`, toolID, uid).Scan(&imagePath)
	if err != nil {
		utils.LogActivity(c, uid, "delete_tool", false, "not found")
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Outil introuvable"})
		return
	}

	res, err := db.Exec(`DELETE FROM tools WHERE tool_id = ? AND user_id = ?`, toolID, uid)
	if err != nil {
		utils.LogActivity(c, uid, "delete_tool", false, "delete error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		utils.LogActivity(c, uid, "delete_tool", false, "not found")
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Outil introuvable"})
		return
	}

	if imagePath != "" {
		absPath := filepath.Join("/var/www/toolcenter/storage/", imagePath)
		_ = os.Remove(absPath)
	}
	utils.LogActivity(c, uid, "delete_tool", true, "")
	c.JSON(http.StatusOK, gin.H{"success": true})
}
