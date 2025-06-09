package user

import (
	"net/http"

	"toolcenter/config"
	"toolcenter/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type deleteAccountRequest struct {
	Password string `json:"password"`
}

func DeleteAccountHandler(c *gin.Context) {
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
		utils.LogActivity(c, uid, "delete_account", false, err.Error())
		c.JSON(code, gin.H{"success": false, "message": err.Error()})
		return
	}

	var req deleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Password == "" {
		utils.LogActivity(c, uid, "delete_account", false, "password missing")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Mot de passe requis"})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, uid, "delete_account", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	var hash string
	err = db.QueryRow(`SELECT password_hash FROM users WHERE user_id = ?`, uid).Scan(&hash)
	if err != nil {
		utils.LogActivity(c, uid, "delete_account", false, "select error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		utils.LogActivity(c, uid, "delete_account", false, "wrong password")
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Mot de passe incorrect"})
		return
	}

	_, err = db.Exec(`DELETE FROM users WHERE user_id = ?`, uid)
	if err != nil {
		utils.LogActivity(c, uid, "delete_account", false, "delete error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	utils.LogActivity(c, uid, "delete_account", true, "")
	c.JSON(http.StatusOK, gin.H{"success": true})
}
