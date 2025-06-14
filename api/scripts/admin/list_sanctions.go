package admin

import (
    "database/sql"
    "net/http"
    "time"

    "toolcenter/config"
    "toolcenter/utils"

    "github.com/gin-gonic/gin"
)

// ListSanctionsHandler returns recent sanctions
func ListSanctionsHandler(c *gin.Context) {
    modID := c.GetString("user_id")

    db, err := config.OpenDB()
    if err != nil {
        utils.LogActivity(c, modID, "list_sanctions", false, "db open error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer db.Close()

    rows, err := db.Query(`SELECT sanction_id, user_id, sanction_type, reason, start_date, end_date, contestable, appeal_status FROM user_sanctions ORDER BY start_date DESC LIMIT 100`)
    if err != nil {
        utils.LogActivity(c, modID, "list_sanctions", false, "query error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer rows.Close()

    sanctions := make([]gin.H, 0)
    for rows.Next() {
        var (
            id int
            uid, typ, reason string
            start time.Time
            end sql.NullTime
            contestable bool
            status sql.NullString
        )
        if err := rows.Scan(&id, &uid, &typ, &reason, &start, &end, &contestable, &status); err != nil {
            continue
        }
        entry := gin.H{
            "sanction_id": id,
            "user_id":     uid,
            "type":        typ,
            "reason":      reason,
            "start":       start,
            "contestable": contestable,
        }
        if end.Valid {
            entry["end"] = end.Time
        }
        if status.Valid {
            entry["appeal_status"] = status.String
        }
        sanctions = append(sanctions, entry)
    }

    utils.LogActivity(c, modID, "list_sanctions", true, "")
    c.JSON(http.StatusOK, gin.H{"success": true, "sanctions": sanctions})
}
