package user

import (
    "net/http"
    "time"

    "toolcenter/config"
    "toolcenter/utils"

    "github.com/gin-gonic/gin"
)

// ListTicketsHandler returns all support tickets for the authenticated user
func ListTicketsHandler(c *gin.Context) {
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
        utils.LogActivity(c, "", "list_tickets", false, err.Error())
        c.JSON(code, gin.H{"success": false, "message": err.Error()})
        return
    }

    db, err := config.OpenDB()
    if err != nil {
        utils.LogActivity(c, uid, "list_tickets", false, "db open error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer db.Close()

    rows, err := db.Query(`SELECT ticket_id, subject, message, status, created_at FROM support_tickets WHERE user_id = ? ORDER BY created_at DESC`, uid)
    if err != nil {
        utils.LogActivity(c, uid, "list_tickets", false, "query error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer rows.Close()

    tickets := make([]gin.H, 0)
    for rows.Next() {
        var (
            id int
            subject, message, status string
            created time.Time
        )
        if err := rows.Scan(&id, &subject, &message, &status, &created); err != nil {
            continue
        }
        tickets = append(tickets, gin.H{
            "ticket_id": id,
            "subject":   subject,
            "message":   message,
            "status":    status,
            "created":   created,
        })
    }

    utils.LogActivity(c, uid, "list_tickets", true, "")
    c.JSON(http.StatusOK, gin.H{"success": true, "tickets": tickets})
}
