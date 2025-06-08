package admin

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"toolcenter/config"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type UserInfo struct {
	UserID   string    `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	Status   string    `json:"status"`
	Created  time.Time `json:"created_at"`
}

func UserListHandler(c *gin.Context) {
	db, err := config.OpenDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	search := strings.TrimSpace(c.Query("search"))
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	limit := 50
	offset := (page - 1) * limit

	var rows *sql.Rows
	if search != "" {
		rows, err = db.Query(`SELECT user_id, username, email, role, account_status, created_at FROM users WHERE username LIKE ? ORDER BY created_at DESC LIMIT ? OFFSET ?`, "%"+search+"%", limit, offset)
	} else {
		rows, err = db.Query(`SELECT user_id, username, email, role, account_status, created_at FROM users ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer rows.Close()

	users := make([]UserInfo, 0)
	for rows.Next() {
		var u UserInfo
		if err := rows.Scan(&u.UserID, &u.Username, &u.Email, &u.Role, &u.Status, &u.Created); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false})
			return
		}
		users = append(users, u)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "users": users})
}
