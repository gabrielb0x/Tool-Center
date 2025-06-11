package admin

import (
	"database/sql"
	"net/http"

	"toolcenter/config"
	"toolcenter/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func UserDetailsHandler(c *gin.Context) {
	adminID, _ := c.Get("user_id")
	modID, _ := adminID.(string)

	uid := c.Param("id")
	if uid == "" {
		utils.LogActivity(c, modID, "user_details", false, "id manquant")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "id manquant"})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, modID, "user_details", false, "db open error")
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
		utils.LogActivity(c, modID, "user_details", false, "not found")
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "not found"})
		return
	}
	if err != nil {
		utils.LogActivity(c, modID, "user_details", false, "query error")
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

	utils.LogActivity(c, modID, "user_details", true, "")

	c.JSON(http.StatusOK, gin.H{"success": true, "user": user})
}
