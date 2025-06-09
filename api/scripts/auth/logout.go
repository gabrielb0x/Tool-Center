package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"toolcenter/config"
	"toolcenter/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func LogoutHandler(c *gin.Context) {
	header := c.GetHeader("Authorization")
	if len(header) < 8 || header[:7] != "Bearer " {
		utils.LogActivity(c, "", "logout_attempt", false, "missing token")
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Token manquant"})
		return
	}
	raw := header[7:]
	sum := sha256.Sum256([]byte(raw))
	tokenHash := hex.EncodeToString(sum[:])

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, "", "logout_attempt", false, "db error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	_, err = db.Exec(`DELETE FROM user_tokens WHERE token = ?`, tokenHash)
	if err != nil {
		utils.LogActivity(c, "", "logout", false, "delete failed")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	utils.LogActivity(c, "", "logout", true, "")
	c.JSON(http.StatusOK, gin.H{"success": true})
}
