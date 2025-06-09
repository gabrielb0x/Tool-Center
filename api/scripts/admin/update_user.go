package admin

import (
	"net/http"

	"toolcenter/config"
	"toolcenter/utils"

	"github.com/gin-gonic/gin"
)

type updateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	Bio      string `json:"bio"`
}

func UpdateUserHandler(c *gin.Context) {
	uid := c.Param("id")
	if uid == "" {
		utils.LogActivity(c, "", "update_user", false, "id manquant")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "id manquant"})
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogActivity(c, "", "update_user", false, "invalid data")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "données invalides"})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, "", "update_user", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	_, err = db.Exec(`UPDATE users SET username=?, email=?, role=?, account_status=?, bio=? WHERE user_id=?`,
		req.Username, req.Email, req.Role, req.Status, req.Bio, uid)
	if err != nil {
		utils.LogActivity(c, "", "update_user", false, "update error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	utils.LogActivity(c, "", "update_user", true, "")
	c.JSON(http.StatusOK, gin.H{"success": true})
}
