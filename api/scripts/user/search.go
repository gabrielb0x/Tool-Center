package user

import (
    "database/sql"
    "net/http"
    "strconv"
    "strings"
    "time"

    "toolcenter/config"
    "toolcenter/utils"

    "github.com/gin-gonic/gin"
    _ "github.com/go-sql-driver/mysql"
)

// SearchUser represents a public user entry for search results.
type SearchUser struct {
    UserID    string    `json:"user_id"`
    Username  string    `json:"username"`
    AvatarURL *string   `json:"avatar_url,omitempty"`
    CreatedAt time.Time `json:"created_at"`
}

// SearchHandler searches users by username with pagination.
func SearchHandler(c *gin.Context) {
    query := strings.TrimSpace(c.Query("query"))
    if query == "" {
        utils.LogActivity(c, "", "user_search", false, "query missing")
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "query required"})
        return
    }

    page, _ := strconv.Atoi(c.Query("page"))
    if page < 1 {
        page = 1
    }
    limit := config.Get().UserSearchLimit
    offset := (page - 1) * limit

    db, err := config.OpenDB()
    if err != nil {
        utils.LogActivity(c, "", "user_search", false, "db open error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer db.Close()

    rows, err := db.Query(`SELECT user_id, username, avatar_url, created_at FROM users WHERE username LIKE ? ORDER BY created_at DESC LIMIT ? OFFSET ?`, "%"+query+"%", limit, offset)
    if err != nil {
        utils.LogActivity(c, "", "user_search", false, "query error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer rows.Close()

    results := make([]SearchUser, 0)
    for rows.Next() {
        var u SearchUser
        var avatar sql.NullString
        if err := rows.Scan(&u.UserID, &u.Username, &avatar, &u.CreatedAt); err != nil {
            utils.LogActivity(c, "", "user_search", false, "scan error")
            c.JSON(http.StatusInternalServerError, gin.H{"success": false})
            return
        }
        if avatar.Valid {
            u.AvatarURL = &avatar.String
        }
        results = append(results, u)
    }

    utils.LogActivity(c, "", "user_search", true, "")
    c.JSON(http.StatusOK, gin.H{"success": true, "users": results})
}

