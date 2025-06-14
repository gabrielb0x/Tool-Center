package admin

import (
    "database/sql"
    "net/http"
    "time"

    "toolcenter/config"
    "toolcenter/utils"

    "github.com/gin-gonic/gin"
)

type sanctionRequest struct {
    UserID      string     `json:"user_id"`
    Type        string     `json:"type"`
    Reason      string     `json:"reason"`
    EndDate     *time.Time `json:"end_date"`
    Contestable bool       `json:"contestable"`
}

// CreateSanctionHandler creates a sanction for a user
func CreateSanctionHandler(c *gin.Context) {
    modID := c.GetString("user_id")

    var req sanctionRequest
    if err := c.ShouldBindJSON(&req); err != nil || req.UserID == "" || req.Reason == "" {
        utils.LogActivity(c, modID, "create_sanction", false, "bad request")
        c.JSON(http.StatusBadRequest, gin.H{"success": false})
        return
    }
    if req.Type == "" {
        req.Type = "Warn"
    }

    db, err := config.OpenDB()
    if err != nil {
        utils.LogActivity(c, modID, "create_sanction", false, "db open error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer db.Close()

    var end sql.NullTime
    if req.EndDate != nil {
        end = sql.NullTime{Valid: true, Time: *req.EndDate}
    }

    _, err = db.Exec(`INSERT INTO user_sanctions (user_id, moderator_id, sanction_type, reason, end_date, contestable, appeal_status)
        VALUES (?, ?, ?, ?, ?, ?, 'Pending')`, req.UserID, modID, req.Type, req.Reason, end, req.Contestable)
    if err != nil {
        utils.LogActivity(c, modID, "create_sanction", false, "insert error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }

    utils.LogActivity(c, modID, "create_sanction", true, "")
    c.JSON(http.StatusOK, gin.H{"success": true})
}
