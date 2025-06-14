package user

import (
    "database/sql"
    "net/http"
    "time"

    "toolcenter/config"
    "toolcenter/utils"

    "github.com/gin-gonic/gin"
)

// SanctionsHandler returns all sanctions of the authenticated user
func SanctionsHandler(c *gin.Context) {
    uid, _, _, _, err := utils.Check(c, utils.CheckOpts{
        RequireToken:    true,
        RequireVerified: true,
    })
    if err != nil {
        code := http.StatusInternalServerError
        switch err {
        case utils.ErrMissingToken, utils.ErrInvalidToken, utils.ErrExpiredToken:
            code = http.StatusUnauthorized
        case utils.ErrEmailNotVerified:
            code = http.StatusForbidden
        }
        utils.LogActivity(c, "", "get_sanctions", false, err.Error())
        c.JSON(code, gin.H{"success": false, "message": err.Error()})
        return
    }

    db, err := config.OpenDB()
    if err != nil {
        utils.LogActivity(c, uid, "get_sanctions", false, "db open error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer db.Close()

    rows, err := db.Query(`SELECT sanction_id, sanction_type, reason, start_date, end_date, contestable, appeal_status, appeal_text
        FROM user_sanctions WHERE user_id = ? ORDER BY start_date DESC`, uid)
    if err != nil {
        utils.LogActivity(c, uid, "get_sanctions", false, "query error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer rows.Close()

    sanctions := make([]gin.H, 0)
    for rows.Next() {
        var (
            id int
            typ, reason string
            start time.Time
            end sql.NullTime
            contestable bool
            status sql.NullString
            appeal sql.NullString
        )
        if err := rows.Scan(&id, &typ, &reason, &start, &end, &contestable, &status, &appeal); err != nil {
            continue
        }
        entry := gin.H{
            "sanction_id": id,
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
        if appeal.Valid {
            entry["appeal_text"] = appeal.String
        }
        sanctions = append(sanctions, entry)
    }

    utils.LogActivity(c, uid, "get_sanctions", true, "")
    c.JSON(http.StatusOK, gin.H{"success": true, "sanctions": sanctions})
}
