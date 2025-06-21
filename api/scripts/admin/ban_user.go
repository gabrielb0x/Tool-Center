package admin

import (
        "database/sql"
        "net/http"
        "time"

        "toolcenter/config"
        "toolcenter/utils"

        "github.com/gin-gonic/gin"
        _ "github.com/go-sql-driver/mysql"
)

type banRequest struct {
        Reason   string `json:"reason"`
        Duration int    `json:"duration_hours"`
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
       moderatorRole := c.GetString("role")
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

       var targetRole string
       err = db.QueryRow(`SELECT role FROM users WHERE user_id = ?`, targetID).Scan(&targetRole)
       if err == sql.ErrNoRows {
               utils.LogActivity(c, moderatorID, "ban_user", false, "target not found")
               c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "utilisateur introuvable"})
               return
       }
       if err != nil {
               utils.LogActivity(c, moderatorID, "ban_user", false, "query error")
               c.JSON(http.StatusInternalServerError, gin.H{"success": false})
               return
       }

       if moderatorRole == "Moderator" && targetRole != "User" {
               utils.LogActivity(c, moderatorID, "ban_user", false, "forbidden")
               c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "action interdite"})
               return
       }
       if moderatorRole == "Admin" && targetRole == "Admin" {
               utils.LogActivity(c, moderatorID, "ban_user", false, "forbidden")
               c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "action interdite"})
               return
       }

       maxHours := config.Get().Moderation.MaxBanDays * 24
       if req.Duration < 0 || req.Duration > maxHours {
               req.Duration = maxHours
       }

       var end sql.NullTime
       if req.Duration > 0 {
               end.Valid = true
               end.Time = time.Now().Add(time.Duration(req.Duration) * time.Hour)
       }

       _, err = db.Exec(`UPDATE users SET account_status = 'Banned' WHERE user_id = ?`, targetID)
       if err != nil {
               utils.LogActivity(c, moderatorID, "ban_user", false, "update error")
               c.JSON(http.StatusInternalServerError, gin.H{"success": false})
               return
       }
       _, _ = db.Exec(`INSERT INTO moderation_actions (moderator_id, user_id, action_type, reason, start_date, end_date) VALUES (?, ?, 'Ban', ?, NOW(), ?)`, moderatorID, targetID, req.Reason, end)
       utils.LogActivity(c, moderatorID, "ban_user", true, "")
       c.JSON(http.StatusOK, gin.H{"success": true})
}
