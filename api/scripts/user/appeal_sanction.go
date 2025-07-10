package user

import (
    "net/http"
    "strconv"

    "toolcenter/config"
    "toolcenter/utils"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

type appealRequest struct {
    Message string `json:"message"`
}

func AppealSanctionHandler(c *gin.Context) {
    uid, _, _, _, err := utils.Check(c, utils.CheckOpts{RequireToken: true})
    if err != nil {
        code := http.StatusInternalServerError
        if err == utils.ErrMissingToken || err == utils.ErrInvalidToken || err == utils.ErrExpiredToken {
            code = http.StatusUnauthorized
        }
        c.JSON(code, gin.H{"success": false, "message": err.Error()})
        return
    }

    sid, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "id invalide"})
        return
    }

    var req appealRequest
    if err := c.ShouldBindJSON(&req); err != nil || req.Message == "" {
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "message manquant"})
        return
    }

    db, err := config.OpenDB()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer db.Close()

    var exists int
    err = db.QueryRow(`SELECT COUNT(*) FROM moderation_actions WHERE action_id=? AND user_id=?`, sid, uid).Scan(&exists)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    if exists == 0 {
        c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "sanction introuvable"})
        return
    }

    _ = db.QueryRow(`SELECT COUNT(*) FROM sanction_appeals WHERE action_id=? AND status='Pending'`, sid).Scan(&exists)
    if exists > 0 {
        c.JSON(http.StatusConflict, gin.H{"success": false, "message": "déjà contestée"})
        return
    }

    uuidV7, _ := uuid.NewV7()
    _, err = db.Exec(`INSERT INTO sanction_appeals (appeal_id, action_id, user_id, message) VALUES (?,?,?,?)`, uuidV7.String(), sid, uid, req.Message)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }

    var email string
    _ = db.QueryRow(`SELECT email FROM users WHERE user_id=?`, uid).Scan(&email)
    if email != "" {
        body := utils.BuildStyledEmail("Contestation enregistrée", "Votre contestation a bien été reçue et sera examinée par l'équipe.", "", "")
        _ = utils.QueueEmail(db, email, "Contestation reçue", body)
    }

    c.JSON(http.StatusOK, gin.H{"success": true})
}

