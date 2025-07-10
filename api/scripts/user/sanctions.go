package user

import (
    "database/sql"
    "net/http"

    "toolcenter/config"
    "toolcenter/utils"

    "github.com/gin-gonic/gin"
)

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

    type sanction struct {
        ID     int            `json:"id"`
        Type   string         `json:"type"`
        Reason sql.NullString `json:"reason"`
        Start  sql.NullTime   `json:"start"`
        End    sql.NullTime   `json:"end"`
    }

    rows, err := db.Query(`SELECT action_id, action_type, reason, start_date, end_date FROM moderation_actions WHERE user_id=? ORDER BY action_date DESC`, uid)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer rows.Close()

    active := make([]gin.H, 0)
    expired := make([]gin.H, 0)
    for rows.Next() {
        var s sanction
        if err := rows.Scan(&s.ID, &s.Type, &s.Reason, &s.Start, &s.End); err == nil {
            item := gin.H{"id": s.ID, "type": s.Type}
            if s.Reason.Valid {
                item["reason"] = s.Reason.String
            }
            if s.Start.Valid {
                item["start"] = s.Start.Time
            }
            if s.End.Valid {
                item["end"] = s.End.Time
            }
            if utils.SanctionActive(s.End) {
                active = append(active, item)
            } else {
                expired = append(expired, item)
            }
        }
    }

    rows2, err := db.Query(`SELECT warn_id, reason, warn_date, expires_at FROM user_warns WHERE user_id=? ORDER BY warn_date DESC`, uid)
    if err == nil {
        defer rows2.Close()
        for rows2.Next() {
            var id int
            var reason sql.NullString
            var start, end sql.NullTime
            if err := rows2.Scan(&id, &reason, &start, &end); err == nil {
                item := gin.H{"id": id, "type": "Warn"}
                if reason.Valid {
                    item["reason"] = reason.String
                }
                if start.Valid {
                    item["start"] = start.Time
                }
                if end.Valid {
                    item["end"] = end.Time
                }
                if utils.SanctionActive(end) {
                    active = append(active, item)
                } else {
                    expired = append(expired, item)
                }
            }
        }
    }

    c.JSON(http.StatusOK, gin.H{"success": true, "active": active, "expired": expired})
}
