package user

import (
    "database/sql"
    "net/http"
    "time"

    "toolcenter/config"
    "toolcenter/utils"

    "github.com/gin-gonic/gin"
    _ "github.com/go-sql-driver/mysql"
)

// UserProfile contains the public information of a user.
type UserProfile struct {
    Username  string    `json:"username"`
    AvatarURL *string   `json:"avatar_url,omitempty"`
    Bio       *string   `json:"bio,omitempty"`
    Tools     int       `json:"tools_count"`
    JoinedAt  time.Time `json:"joined_at"`
}

// ProfileHandler returns the public profile of a user by username.
func ProfileHandler(c *gin.Context) {
    username := c.Param("username")
    if username == "" {
        utils.LogActivity(c, "", "user_profile", false, "missing username")
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "username required"})
        return
    }

    db, err := config.OpenDB()
    if err != nil {
        utils.LogActivity(c, "", "user_profile", false, "db open error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer db.Close()

    var (
        id     string
        avatar sql.NullString
        bio    sql.NullString
        created time.Time
    )
    err = db.QueryRow(`SELECT user_id, avatar_url, bio, created_at FROM users WHERE username = ? LIMIT 1`, username).Scan(&id, &avatar, &bio, &created)
    switch {
    case err == sql.ErrNoRows:
        utils.LogActivity(c, "", "user_profile", false, "not found")
        c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "not found"})
        return
    case err != nil:
        utils.LogActivity(c, "", "user_profile", false, "query error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }

    var toolsCount int
    _ = db.QueryRow(`SELECT tools_posted_count FROM user_stats WHERE user_id = ?`, id).Scan(&toolsCount)

    profile := UserProfile{
        Username: username,
        Tools:    toolsCount,
        JoinedAt: created,
    }
    if avatar.Valid {
        profile.AvatarURL = &avatar.String
    }
    if bio.Valid {
        profile.Bio = &bio.String
    }

    utils.LogActivity(c, id, "user_profile", true, "")
    c.JSON(http.StatusOK, gin.H{"success": true, "profile": profile})
}

