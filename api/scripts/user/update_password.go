package user

import (
        "net/http"

        "toolcenter/config"
        "toolcenter/utils"

        "github.com/gin-gonic/gin"
        _ "github.com/go-sql-driver/mysql"
        "golang.org/x/crypto/bcrypt"
)

type updatePasswordRequest struct {
        CurrentPassword string `json:"current_password"`
        NewPassword     string `json:"new_password"`
}

func UpdatePasswordHandler(c *gin.Context) {
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
                utils.LogActivity(c, uid, "update_password", false, err.Error())
                c.JSON(code, gin.H{"success": false, "message": err.Error()})
                return
        }

        var req updatePasswordRequest
        if err := c.ShouldBindJSON(&req); err != nil || req.CurrentPassword == "" || len(req.NewPassword) < 7 || len(req.NewPassword) > 30 {
                utils.LogActivity(c, uid, "update_password", false, "bad request")
                c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Requête invalide"})
                return
        }

        db, err := config.OpenDB()
        if err != nil {
                utils.LogActivity(c, uid, "update_password", false, "db open error")
                c.JSON(http.StatusInternalServerError, gin.H{"success": false})
                return
        }
        defer db.Close()

        var hash string
        err = db.QueryRow(`SELECT password_hash FROM users WHERE user_id = ?`, uid).Scan(&hash)
        if err != nil {
                utils.LogActivity(c, uid, "update_password", false, "select error")
                c.JSON(http.StatusInternalServerError, gin.H{"success": false})
                return
        }
        if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.CurrentPassword)) != nil {
                utils.LogActivity(c, uid, "update_password", false, "wrong password")
                c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Mot de passe incorrect"})
                return
        }

        newHash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
        _, err = db.Exec(`UPDATE users SET password_hash = ?, password_changed_at = NOW() WHERE user_id = ?`, newHash, uid)
        if err != nil {
                utils.LogActivity(c, uid, "update_password", false, "update error")
                c.JSON(http.StatusInternalServerError, gin.H{"success": false})
                return
        }

        utils.LogActivity(c, uid, "update_password", true, "")
        c.JSON(http.StatusOK, gin.H{"success": true})
}
