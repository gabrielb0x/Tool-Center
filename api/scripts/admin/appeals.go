package admin

import (
    "net/http"
    "toolcenter/config"
    "toolcenter/utils"

    "github.com/gin-gonic/gin"
)

func ListAppealsHandler(c *gin.Context) {
    db, err := config.OpenDB()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer db.Close()

    rows, err := db.Query(`SELECT appeal_id, action_id, user_id, message, status, created_at FROM sanction_appeals ORDER BY created_at DESC`)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer rows.Close()

    appeals := make([]gin.H, 0)
    for rows.Next() {
        var id, action, uid, msg, status string
        var created string
        if err := rows.Scan(&id, &action, &uid, &msg, &status, &created); err == nil {
            appeals = append(appeals, gin.H{
                "appeal_id": id,
                "action_id": action,
                "user_id":   uid,
                "message":   msg,
                "status":    status,
                "created_at": created,
            })
        }
    }
    c.JSON(http.StatusOK, gin.H{"success": true, "appeals": appeals})
}

func ReviewAppealHandler(c *gin.Context) {
    adminID := c.GetString("user_id")
    appealID := c.Param("id")

    var req struct {
        Approve bool   `json:"approve"`
        Message string `json:"message"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"success": false})
        return
    }

    db, err := config.OpenDB()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer db.Close()

    var actionID, userID string
    err = db.QueryRow(`SELECT action_id, user_id FROM sanction_appeals WHERE appeal_id=? AND status='Pending'`, appealID).Scan(&actionID, &userID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"success": false})
        return
    }

    status := "Rejected"
    if req.Approve {
        status = "Approved"
        _, _ = db.Exec(`UPDATE moderation_actions SET end_date=NOW() WHERE action_id=?`, actionID)
    }
    _, _ = db.Exec(`UPDATE sanction_appeals SET status=?, reviewed_by=?, reviewed_at=NOW() WHERE appeal_id=?`, status, adminID, appealID)

    var email string
    _ = db.QueryRow(`SELECT email FROM users WHERE user_id=?`, userID).Scan(&email)
    if email != "" {
        body := utils.BuildStyledEmail("Décision de contestation", req.Message, "", "")
        _ = utils.QueueEmail(db, email, "Contestation", body)
    }

    c.JSON(http.StatusOK, gin.H{"success": true})
}

