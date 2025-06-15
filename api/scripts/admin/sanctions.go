package admin

import (
	"database/sql"
	"net/http"
	"time"

	"toolcenter/config"
	"toolcenter/scripts/user"
	"toolcenter/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type sanctionRequest struct {
	Type       string  `json:"type"`
	Reason     string  `json:"reason"`
	Appealable bool    `json:"appealable"`
	EndDate    *string `json:"end_date"`
}

// AddSanctionHandler creates a new sanction for a user.
func AddSanctionHandler(c *gin.Context) {
	targetID := c.Param("id")
	adminID, _ := c.Get("user_id")
	uid, _ := adminID.(string)

	var req sanctionRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Type == "" {
		utils.LogActivity(c, uid, "add_sanction", false, "invalid payload")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "données invalides"})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, uid, "add_sanction", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	id, _ := uuid.NewV7()
	sanctionID := id.String()

	var end interface{}
	if req.EndDate != nil {
		if t, err := time.Parse(time.RFC3339, *req.EndDate); err == nil {
			end = t
		}
	}

	_, err = db.Exec(`INSERT INTO sanctions (sanction_id, user_id, moderator_id, type, reason, appealable, end_date) VALUES (?,?,?,?,?,?,?)`,
		sanctionID, targetID, uid, req.Type, req.Reason, req.Appealable, end)
	if err != nil {
		utils.LogActivity(c, uid, "add_sanction", false, "insert error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	utils.LogActivity(c, uid, "add_sanction", true, "")
	c.JSON(http.StatusOK, gin.H{"success": true, "sanction_id": sanctionID})
}

// ListSanctionsHandler returns sanctions for the given user.
func ListSanctionsHandler(c *gin.Context) {
	targetID := c.Param("id")
	adminID, _ := c.Get("user_id")
	uid, _ := adminID.(string)

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, uid, "list_sanctions", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	rows, err := db.Query(`SELECT sanction_id, type, reason, appealable, start_date, end_date, status, moderator_id FROM sanctions WHERE user_id = ? ORDER BY start_date DESC`, targetID)
	if err != nil {
		utils.LogActivity(c, uid, "list_sanctions", false, "query error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer rows.Close()

	list := make([]user.Sanction, 0)
	for rows.Next() {
		var s user.Sanction
		var end sql.NullTime
		var mod sql.NullString
		if err := rows.Scan(&s.ID, &s.Type, &s.Reason, &s.Appealable, &s.StartDate, &end, &s.Status, &mod); err != nil {
			utils.LogActivity(c, uid, "list_sanctions", false, "scan error")
			c.JSON(http.StatusInternalServerError, gin.H{"success": false})
			return
		}
		if end.Valid {
			tm := end.Time.Format(time.RFC3339)
			s.EndDate = &tm
		}
		if mod.Valid {
			s.ModeratorID = &mod.String
		}
		list = append(list, s)
	}

	utils.LogActivity(c, uid, "list_sanctions", true, "")
	c.JSON(http.StatusOK, gin.H{"success": true, "sanctions": list})
}
