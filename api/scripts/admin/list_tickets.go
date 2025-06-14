package admin

import (
    "net/http"
    "time"

    "toolcenter/config"
    "toolcenter/utils"

    "github.com/gin-gonic/gin"
)

// ListTicketsHandler returns all support tickets
func ListTicketsHandler(c *gin.Context) {
    modID := c.GetString("user_id")

    db, err := config.OpenDB()
    if err != nil {
        utils.LogActivity(c, modID, "list_tickets", false, "db open error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer db.Close()

    rows, err := db.Query(`SELECT ticket_id, user_id, subject, message, status, created_at FROM support_tickets ORDER BY created_at DESC LIMIT 100`)
    if err != nil {
        utils.LogActivity(c, modID, "list_tickets", false, "query error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer rows.Close()

    tickets := make([]gin.H, 0)
    for rows.Next() {
        var (
            id int
            uid, subject, message, status string
            created time.Time
        )
        if err := rows.Scan(&id, &uid, &subject, &message, &status, &created); err != nil {
            continue
        }
        tickets = append(tickets, gin.H{
            "ticket_id": id,
            "user_id":   uid,
            "subject":   subject,
            "message":   message,
            "status":    status,
            "created":   created,
        })
    }

    utils.LogActivity(c, modID, "list_tickets", true, "")
    c.JSON(http.StatusOK, gin.H{"success": true, "tickets": tickets})
}
