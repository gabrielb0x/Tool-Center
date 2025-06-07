package tools

import (
	"net/http"

	"toolcenter/config"
	"toolcenter/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func DeleteToolHandler(c *gin.Context) {
	uid, _, _, err := utils.Check(c, utils.CheckOpts{
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
		c.JSON(code, gin.H{"success": false, "message": err.Error()})
		return
	}

	toolID := c.Param("id")
	if toolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ID manquant"})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	res, err := db.Exec(`DELETE FROM tools WHERE tool_id = ? AND user_id = ?`, toolID, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Outil introuvable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
