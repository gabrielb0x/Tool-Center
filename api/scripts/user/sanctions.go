package user

import (
    "database/sql"
    "net/http"

    "toolcenter/config"
    "toolcenter/utils"

    "github.com/gin-gonic/gin"
)

// SanctionsHandler returns all sanctions for the authenticated user.
func SanctionsHandler(c *gin.Context) {
    uid, _, _, _, err := utils.Check(c, utils.CheckOpts{RequireToken: true})
    if err != nil {
        code := http.StatusInternalServerError
        if err == utils.ErrMissingToken || err == utils.ErrInvalidToken || err == utils.ErrExpiredToken {
            code = http.StatusUnauthorized
        }
        c.JSON(code, gin.H{"success": false, "message": err.Error()})
        return
    }

    db, err := config.OpenDB()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer db.Close()

    rows, err := db.Query(`SELECT sanction_id,type,reason,start_date,end_date FROM user_sanctions WHERE user_id=? ORDER BY start_date DESC`, uid)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer rows.Close()

    sanctions := []gin.H{}
    for rows.Next() {
        var id int
        var t, reason string
        var start, end sql.NullTime
        if err := rows.Scan(&id, &t, &reason, &start, &end); err == nil {
            s := gin.H{
                "sanction_id": id,
                "type":        t,
                "reason":      reason,
                "start_date":  start.Time,
            }
            if end.Valid {
                s["end_date"] = end.Time
            }
            sanctions = append(sanctions, s)
        }
    }

    c.JSON(http.StatusOK, gin.H{"success": true, "sanctions": sanctions})
}
