package user

import (
	"database/sql"
	"encoding/base64"
	"net/http"

	"toolcenter/config"
	"toolcenter/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

type enable2FARequest struct {
	Secret string `json:"secret"`
	Code   string `json:"code"`
}

type disable2FARequest struct {
	Code string `json:"code"`
}

func Generate2FAHandler(c *gin.Context) {
	uid, _, _, _, err := utils.Check(c, utils.CheckOpts{RequireToken: true, RequireVerified: true, RequireNotBanned: true})
	if err != nil {
		code := http.StatusInternalServerError
		switch err {
		case utils.ErrMissingToken, utils.ErrInvalidToken, utils.ErrExpiredToken:
			code = http.StatusUnauthorized
		case utils.ErrEmailNotVerified, utils.ErrAccountBanned:
			code = http.StatusForbidden
		}
		utils.LogActivity(c, uid, "generate_2fa", false, err.Error())
		c.JSON(code, gin.H{"success": false, "message": err.Error()})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, uid, "generate_2fa", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	var existing sql.NullString
	var username string
	if err := db.QueryRow(`SELECT authenticator_secret, username FROM users WHERE user_id=?`, uid).Scan(&existing, &username); err != nil {
		utils.LogActivity(c, uid, "generate_2fa", false, "query error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	if existing.Valid && existing.String != "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "2FA déjà activée"})
		return
	}
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      config.Get().TwoFactor.Issuer,
		AccountName: username,
	})
	if err != nil {
		utils.LogActivity(c, uid, "generate_2fa", false, "generate error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	png, err := qrcode.Encode(key.URL(), qrcode.Medium, 256)
	if err != nil {
		utils.LogActivity(c, uid, "generate_2fa", false, "qrcode error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	encoded := base64.StdEncoding.EncodeToString(png)
	utils.LogActivity(c, uid, "generate_2fa", true, "")
	c.JSON(http.StatusOK, gin.H{"success": true, "secret": key.Secret(), "qr_code": encoded})
}

func Enable2FAHandler(c *gin.Context) {
	uid, _, _, _, err := utils.Check(c, utils.CheckOpts{RequireToken: true, RequireVerified: true, RequireNotBanned: true})
	if err != nil {
		code := http.StatusInternalServerError
		switch err {
		case utils.ErrMissingToken, utils.ErrInvalidToken, utils.ErrExpiredToken:
			code = http.StatusUnauthorized
		case utils.ErrEmailNotVerified, utils.ErrAccountBanned:
			code = http.StatusForbidden
		}
		utils.LogActivity(c, uid, "enable_2fa", false, err.Error())
		c.JSON(code, gin.H{"success": false, "message": err.Error()})
		return
	}

	var req enable2FARequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Secret == "" || req.Code == "" {
		utils.LogActivity(c, uid, "enable_2fa", false, "bad request")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides"})
		return
	}

	if !totp.Validate(req.Code, req.Secret) {
		utils.LogActivity(c, uid, "enable_2fa", false, "invalid code")
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Code invalide"})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, uid, "enable_2fa", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	_, err = db.Exec(`UPDATE users SET authenticator_secret=? WHERE user_id=?`, req.Secret, uid)
	if err != nil {
		utils.LogActivity(c, uid, "enable_2fa", false, "update error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	utils.LogActivity(c, uid, "enable_2fa", true, "")
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func Disable2FAHandler(c *gin.Context) {
	uid, _, _, _, err := utils.Check(c, utils.CheckOpts{RequireToken: true, RequireVerified: true, RequireNotBanned: true})
	if err != nil {
		code := http.StatusInternalServerError
		switch err {
		case utils.ErrMissingToken, utils.ErrInvalidToken, utils.ErrExpiredToken:
			code = http.StatusUnauthorized
		case utils.ErrEmailNotVerified, utils.ErrAccountBanned:
			code = http.StatusForbidden
		}
		utils.LogActivity(c, uid, "disable_2fa", false, err.Error())
		c.JSON(code, gin.H{"success": false, "message": err.Error()})
		return
	}

	var req disable2FARequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Code == "" {
		utils.LogActivity(c, uid, "disable_2fa", false, "bad request")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides"})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, uid, "disable_2fa", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	var secret sql.NullString
	err = db.QueryRow(`SELECT authenticator_secret FROM users WHERE user_id=?`, uid).Scan(&secret)
	if err != nil {
		utils.LogActivity(c, uid, "disable_2fa", false, "select error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	if !secret.Valid || secret.String == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "2FA non activée"})
		return
	}

	if !totp.Validate(req.Code, secret.String) {
		utils.LogActivity(c, uid, "disable_2fa", false, "invalid code")
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Code invalide"})
		return
	}

	_, err = db.Exec(`UPDATE users SET authenticator_secret=NULL WHERE user_id=?`, uid)
	if err != nil {
		utils.LogActivity(c, uid, "disable_2fa", false, "update error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	utils.LogActivity(c, uid, "disable_2fa", true, "")
	c.JSON(http.StatusOK, gin.H{"success": true})
}
