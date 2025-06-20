package admin

import (
        "database/sql"
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

        moderatorID := c.GetString("user_id")
        moderatorRole := c.GetString("role")

        db, err := config.OpenDB()
        if err != nil {
                utils.LogActivity(c, moderatorID, "unban_user", false, "db open error")
                c.JSON(http.StatusInternalServerError, gin.H{"success": false})
                return
        }
        defer db.Close()

        var targetRole string
        err = db.QueryRow(`SELECT role FROM users WHERE user_id = ?`, uid).Scan(&targetRole)
        if err == sql.ErrNoRows {
                utils.LogActivity(c, moderatorID, "unban_user", false, "not found")
                c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "utilisateur introuvable"})
                return
        }
        if err != nil {
                utils.LogActivity(c, moderatorID, "unban_user", false, "query error")
                c.JSON(http.StatusInternalServerError, gin.H{"success": false})
                return
        }

        if moderatorRole == "Moderator" && targetRole != "User" {
                utils.LogActivity(c, moderatorID, "unban_user", false, "forbidden role")
                c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "action interdite"})
                return
        }

        _, err = db.Exec(`UPDATE users SET account_status = 'Good', ban_expires_at=NULL WHERE user_id = ?`, uid)
        if err != nil {
                utils.LogActivity(c, moderatorID, "unban_user", false, "update error")
                c.JSON(http.StatusInternalServerError, gin.H{"success": false})
                return
        }
        _, _ = db.Exec(`INSERT INTO moderation_actions (moderator_id, user_id, action_type) VALUES (?, ?, 'Unban')`, moderatorID, uid)
        utils.LogActivity(c, moderatorID, "unban_user", true, "")
        c.JSON(http.StatusOK, gin.H{"success": true})
}
