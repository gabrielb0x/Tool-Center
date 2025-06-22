package admin

import (
    "net/http"

    "toolcenter/config"
    "toolcenter/utils"

    "github.com/gin-gonic/gin"
)

func ClearLogsHandler(c *gin.Context) {
    adminID := c.GetString("user_id")

    db, err := config.OpenDB()
    if err != nil {
        utils.LogActivity(c, adminID, "clear_logs", false, "db open error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer db.Close()

    if _, err := db.Exec("DELETE FROM activity_logs"); err != nil {
        utils.LogActivity(c, adminID, "clear_logs", false, "delete error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }

    utils.LogActivity(c, adminID, "clear_logs", true, "")
    c.JSON(http.StatusOK, gin.H{"success": true})
}

