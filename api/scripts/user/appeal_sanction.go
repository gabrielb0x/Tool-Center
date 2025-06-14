package user

import (
    "net/http"

    "toolcenter/config"
    "toolcenter/utils"

    "github.com/gin-gonic/gin"
)

type appealRequest struct {
    Appeal string `json:"appeal"`
}

// AppealSanctionHandler allows a user to appeal one of their sanctions
func AppealSanctionHandler(c *gin.Context) {
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
        utils.LogActivity(c, "", "appeal_sanction", false, err.Error())
        c.JSON(code, gin.H{"success": false, "message": err.Error()})
        return
    }

    sid := c.Param("id")
    if sid == "" {
        utils.LogActivity(c, uid, "appeal_sanction", false, "id missing")
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "id manquant"})
        return
    }

    var req appealRequest
    if err := c.ShouldBindJSON(&req); err != nil || req.Appeal == "" {
        utils.LogActivity(c, uid, "appeal_sanction", false, "bad request")
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides"})
        return
    }

    db, err := config.OpenDB()
    if err != nil {
        utils.LogActivity(c, uid, "appeal_sanction", false, "db open error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer db.Close()

    var contestable bool
    var sanctionUser string
    var appealStatus string
    err = db.QueryRow(`SELECT user_id, contestable, appeal_status FROM user_sanctions WHERE sanction_id = ?`, sid).
        Scan(&sanctionUser, &contestable, &appealStatus)
    if err != nil {
        utils.LogActivity(c, uid, "appeal_sanction", false, "select error")
        c.JSON(http.StatusNotFound, gin.H{"success": false})
        return
    }
    if sanctionUser != uid {
        utils.LogActivity(c, uid, "appeal_sanction", false, "access denied")
        c.JSON(http.StatusForbidden, gin.H{"success": false})
        return
    }
    if !contestable || appealStatus != "Pending" {
        utils.LogActivity(c, uid, "appeal_sanction", false, "not contestable")
        c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "Sanction non contestable"})
        return
    }

    _, err = db.Exec(`UPDATE user_sanctions SET appeal_text = ?, appeal_status = 'Pending' WHERE sanction_id = ?`, req.Appeal, sid)
    if err != nil {
        utils.LogActivity(c, uid, "appeal_sanction", false, "update error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }

    utils.LogActivity(c, uid, "appeal_sanction", true, "")
    c.JSON(http.StatusOK, gin.H{"success": true})
}
