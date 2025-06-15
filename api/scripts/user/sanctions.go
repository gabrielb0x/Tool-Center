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

// Sanction represents a moderation action applied to a user.
type Sanction struct {
	ID          string  `json:"sanction_id"`
	Type        string  `json:"type"`
	Reason      string  `json:"reason"`
	Appealable  bool    `json:"appealable"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date"`
	Status      string  `json:"status"`
	ModeratorID *string `json:"moderator_id"`
}

// MySanctionsHandler returns sanctions for the authenticated user.
func MySanctionsHandler(c *gin.Context) {
	uid, _, _, _, err := utils.Check(c, utils.CheckOpts{
		RequireToken:     true,
		RequireVerified:  false,
		RequireNotBanned: false,
	})
	if err != nil {
		code := http.StatusInternalServerError
		switch err {
		case utils.ErrMissingToken, utils.ErrInvalidToken, utils.ErrExpiredToken:
			code = http.StatusUnauthorized
		case utils.ErrEmailNotVerified, utils.ErrAccountBanned:
			code = http.StatusForbidden
		}
		c.JSON(code, gin.H{"success": false, "message": err.Error()})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, uid, "get_sanctions", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	rows, err := db.Query(`SELECT sanction_id, type, reason, appealable, start_date, end_date, status, moderator_id FROM sanctions WHERE user_id = ? ORDER BY start_date DESC`, uid)
	if err != nil {
		utils.LogActivity(c, uid, "get_sanctions", false, "query error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer rows.Close()

	sanctions := make([]Sanction, 0)
	for rows.Next() {
		var s Sanction
		var end sql.NullTime
		var mod sql.NullString
		if err := rows.Scan(&s.ID, &s.Type, &s.Reason, &s.Appealable, &s.StartDate, &end, &s.Status, &mod); err != nil {
			utils.LogActivity(c, uid, "get_sanctions", false, "scan error")
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
		sanctions = append(sanctions, s)
	}

	utils.LogActivity(c, uid, "get_sanctions", true, "")
	c.JSON(http.StatusOK, gin.H{"success": true, "sanctions": sanctions})
}
