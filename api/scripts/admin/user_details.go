package admin

import (
	"database/sql"
	"net/http"

	"toolcenter/config"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func UserDetailsHandler(c *gin.Context) {
	uid := c.Param("id")
	if uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "id manquant"})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	var (
		username, email, role, status string
		avatar, bio                   sql.NullString
		created                       sql.NullTime
	)

	err = db.QueryRow(`SELECT username,email,role,account_status,avatar_url,bio,created_at FROM users WHERE user_id = ?`, uid).
		Scan(&username, &email, &role, &status, &avatar, &bio, &created)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	var toolsCount int
	_ = db.QueryRow(`SELECT COUNT(*) FROM tools WHERE user_id = ?`, uid).Scan(&toolsCount)

	user := gin.H{
		"user_id":      uid,
		"username":     username,
		"email":        email,
		"role":         role,
		"status":       status,
		"toolsCount":   toolsCount,
		"reportsCount": 0,
	}
	if avatar.Valid {
		user["avatar"] = avatar.String
	}
	if bio.Valid {
		user["bio"] = bio.String
	}
	if created.Valid {
		user["createdAt"] = created.Time
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "user": user})
}
