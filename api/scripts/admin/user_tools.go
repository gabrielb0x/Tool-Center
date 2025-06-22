package admin

import (
    "net/http"
    "time"

    "toolcenter/config"
    "toolcenter/utils"

    "github.com/gin-gonic/gin"
)

type AdminTool struct {
    ID        string    `json:"tool_id"`
    Title     string    `json:"title"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
}

func UserToolsHandler(c *gin.Context) {
    adminID := c.GetString("user_id")
    uid := c.Param("id")
    if uid == "" {
        utils.LogActivity(c, adminID, "user_tools", false, "id manquant")
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "id manquant"})
        return
    }

    db, err := config.OpenDB()
    if err != nil {
        utils.LogActivity(c, adminID, "user_tools", false, "db open error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer db.Close()

    rows, err := db.Query(`SELECT tool_id, title, status, created_at FROM tools WHERE user_id = ? ORDER BY created_at DESC`, uid)
    if err != nil {
        utils.LogActivity(c, adminID, "user_tools", false, "query error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer rows.Close()

    tools := make([]AdminTool, 0)
    for rows.Next() {
        var t AdminTool
        if err := rows.Scan(&t.ID, &t.Title, &t.Status, &t.CreatedAt); err != nil {
            continue
        }
        tools = append(tools, t)
    }

    utils.LogActivity(c, adminID, "user_tools", true, "")
    c.JSON(http.StatusOK, gin.H{"success": true, "tools": tools})
}

