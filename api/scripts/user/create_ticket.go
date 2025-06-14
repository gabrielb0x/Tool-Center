package user

import (
    "net/http"

    "toolcenter/config"
    "toolcenter/utils"

    "github.com/gin-gonic/gin"
)

type ticketRequest struct {
    Subject string `json:"subject"`
    Message string `json:"message"`
}

// CreateTicketHandler creates a new support ticket
func CreateTicketHandler(c *gin.Context) {
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
        utils.LogActivity(c, "", "create_ticket", false, err.Error())
        c.JSON(code, gin.H{"success": false, "message": err.Error()})
        return
    }

    var req ticketRequest
    if err := c.ShouldBindJSON(&req); err != nil || req.Subject == "" || req.Message == "" {
        utils.LogActivity(c, uid, "create_ticket", false, "bad request")
        c.JSON(http.StatusBadRequest, gin.H{"success": false})
        return
    }

    db, err := config.OpenDB()
    if err != nil {
        utils.LogActivity(c, uid, "create_ticket", false, "db open error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer db.Close()

    _, err = db.Exec(`INSERT INTO support_tickets (user_id, subject, message) VALUES (?, ?, ?)`, uid, req.Subject, req.Message)
    if err != nil {
        utils.LogActivity(c, uid, "create_ticket", false, "insert error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }

    utils.LogActivity(c, uid, "create_ticket", true, "")
    c.JSON(http.StatusOK, gin.H{"success": true})
}
