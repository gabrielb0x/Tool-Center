package admin

import (
    "database/sql"
    "net/http"

    "toolcenter/config"

    "github.com/gin-gonic/gin"
)

func UserSanctionsHandler(c *gin.Context) {
    uid := c.Param("id")
    if uid == "" {
        c.JSON(http.StatusBadRequest, gin.H{"success": false})
        return
    }
    db, err := config.OpenDB()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer db.Close()

    rows, err := db.Query(`SELECT action_id, action_type, reason, start_date, end_date, COALESCE(moderator_id,'systauto') FROM moderation_actions WHERE user_id=? ORDER BY action_date DESC`, uid)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer rows.Close()

    list := make([]gin.H, 0)
    for rows.Next() {
        var id int
        var t string
        var reason sql.NullString
        var start, end sql.NullTime
        var by sql.NullString
        if err := rows.Scan(&id, &t, &reason, &start, &end, &by); err == nil {
            item := gin.H{"id": id, "type": t, "by": by.String}
            if reason.Valid {
                item["reason"] = reason.String
            }
            if start.Valid {
                item["start"] = start.Time
            }
            if end.Valid {
                item["end"] = end.Time
            }
            var status sql.NullString
            _ = db.QueryRow(`SELECT status FROM sanction_appeals WHERE action_id=? ORDER BY created_at DESC LIMIT 1`, id).Scan(&status)
            if status.Valid {
                item["appeal_status"] = status.String
            }
            list = append(list, item)
        }
    }

    c.JSON(http.StatusOK, gin.H{"success": true, "sanctions": list})
}

