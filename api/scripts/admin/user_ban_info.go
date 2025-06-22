package admin

import (
    "database/sql"
    "net/http"

    "toolcenter/config"
    "toolcenter/utils"

    "github.com/gin-gonic/gin"
)

func UserBanInfoHandler(c *gin.Context) {
    adminID := c.GetString("user_id")
    uid := c.Param("id")
    if uid == "" {
        utils.LogActivity(c, adminID, "ban_info", false, "id manquant")
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "id manquant"})
        return
    }

    db, err := config.OpenDB()
    if err != nil {
        utils.LogActivity(c, adminID, "ban_info", false, "db open error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer db.Close()

    var reason sql.NullString
    var start, end sql.NullTime
    err = db.QueryRow(`SELECT reason, start_date, end_date FROM moderation_actions WHERE user_id = ? AND action_type='Ban' ORDER BY action_date DESC LIMIT 1`, uid).Scan(&reason, &start, &end)
    if err == sql.ErrNoRows {
        utils.LogActivity(c, adminID, "ban_info", false, "not found")
        c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "not found"})
        return
    }
    if err != nil {
        utils.LogActivity(c, adminID, "ban_info", false, "query error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }

    info := gin.H{}
    if reason.Valid {
        info["reason"] = reason.String
    }
    if start.Valid {
        info["start"] = start.Time
    }
    if end.Valid {
        info["end"] = end.Time
    }

    utils.LogActivity(c, adminID, "ban_info", true, "")
    c.JSON(http.StatusOK, gin.H{"success": true, "ban": info})
}

