package admin

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

type UserInfo struct {
	UserID     string    `json:"user_id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	Role       string    `json:"role"`
	Status     string    `json:"status"`
	Created    time.Time `json:"created_at"`
	AvatarURL  string    `json:"avatar_url"`
	IsVerified bool      `json:"is_verified"`
	TwoFactor  bool      `json:"two_factor_enabled"`
}

func UserListHandler(c *gin.Context) {
	adminID, _ := c.Get("user_id")
	uid, _ := adminID.(string)

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, uid, "user_list", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	search := strings.TrimSpace(c.Query("search"))
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	limit := 10
	offset := (page - 1) * limit
	var total int

	var rows *sql.Rows
	if search != "" {
		rows, err = db.Query(`SELECT user_id, username, email, role, account_status, created_at, avatar_url, is_verified, authenticator_secret FROM users WHERE username LIKE ? ORDER BY created_at DESC LIMIT ? OFFSET ?`, "%"+search+"%", limit, offset)
		_ = db.QueryRow(`SELECT COUNT(*) FROM users WHERE username LIKE ?`, "%"+search+"%").Scan(&total)
	} else {
		rows, err = db.Query(`SELECT user_id, username, email, role, account_status, created_at, avatar_url, is_verified, authenticator_secret FROM users ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
		_ = db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&total)
	}
	if err != nil {
		utils.LogActivity(c, uid, "user_list", false, "query error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer rows.Close()

	users := make([]UserInfo, 0)
	for rows.Next() {
		var (
			u      UserInfo
			secret sql.NullString
		)
		if err := rows.Scan(&u.UserID, &u.Username, &u.Email, &u.Role, &u.Status, &u.Created, &u.AvatarURL, &u.IsVerified, &secret); err != nil {
			utils.LogActivity(c, uid, "user_list", false, "scan error")
			c.JSON(http.StatusInternalServerError, gin.H{"success": false})
			return
		}
		u.TwoFactor = secret.Valid && secret.String != ""
		users = append(users, u)
	}

	utils.LogActivity(c, uid, "user_list", true, "")

	c.JSON(http.StatusOK, gin.H{"success": true, "users": users, "total": total})
}
