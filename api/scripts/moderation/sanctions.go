package moderation

import (
	"database/sql"
	"net/http"
	"time"

	"toolcenter/config"
	"toolcenter/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type SanctionRequest struct {
	Type           string `json:"type"`
	Reason         string `json:"reason"`
	DurationHours  int    `json:"duration_hours"`
	NotContestable bool   `json:"not_contestable"`
}

func CreateSanctionHandler(c *gin.Context) {
	modID, _ := c.Get("user_id")
	moderatorID, _ := modID.(string)

	uid := c.Param("id")
	if uid == "" {
		utils.LogActivity(c, moderatorID, "create_sanction", false, "id manquant")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "id manquant"})
		return
	}

	var req SanctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogActivity(c, moderatorID, "create_sanction", false, "bad request")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "requete invalide"})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, moderatorID, "create_sanction", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	uuidV7, _ := uuid.NewV7()
	sanctionID := uuidV7.String()

	var expires *time.Time
	if req.DurationHours > 0 {
		t := time.Now().Add(time.Duration(req.DurationHours) * time.Hour)
		expires = &t
	}

	_, err = db.Exec(`INSERT INTO sanctions (sanction_id,user_id,moderator_id,type,reason,not_contestable,expires_at) VALUES (?,?,?,?,?,?,?)`,
		sanctionID, uid, moderatorID, req.Type, req.Reason, req.NotContestable, expires)
	if err != nil {
		utils.LogActivity(c, moderatorID, "create_sanction", false, "insert error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	utils.LogActivity(c, moderatorID, "create_sanction", true, "")
	c.JSON(http.StatusOK, gin.H{"success": true, "sanction_id": sanctionID})
}

func ListSanctionsHandler(c *gin.Context) {
	modID, _ := c.Get("user_id")
	moderatorID, _ := modID.(string)

	uid := c.Param("id")
	if uid == "" {
		utils.LogActivity(c, moderatorID, "list_sanctions", false, "id manquant")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "id manquant"})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, moderatorID, "list_sanctions", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	rows, err := db.Query(`SELECT sanction_id,type,reason,not_contestable,created_at,expires_at FROM sanctions WHERE user_id = ? ORDER BY created_at DESC`, uid)
	if err != nil {
		utils.LogActivity(c, moderatorID, "list_sanctions", false, "query error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer rows.Close()

	sanctions := make([]gin.H, 0)
	for rows.Next() {
		var (
			sid, typ, reason string
			nc               bool
			created, expires sql.NullTime
		)
		if err := rows.Scan(&sid, &typ, &reason, &nc, &created, &expires); err == nil {
			s := gin.H{
				"sanction_id":     sid,
				"type":            typ,
				"reason":          reason,
				"not_contestable": nc,
			}
			if created.Valid {
				s["created_at"] = created.Time
			}
			if expires.Valid {
				s["expires_at"] = expires.Time
			}
			sanctions = append(sanctions, s)
		}
	}

	utils.LogActivity(c, moderatorID, "list_sanctions", true, "")
	c.JSON(http.StatusOK, gin.H{"success": true, "sanctions": sanctions})
}

func AppealSanctionHandler(c *gin.Context) {
	uid, _, _, _, err := utils.Check(c, utils.CheckOpts{RequireToken: true, RequireVerified: true, RequireNotBanned: true})
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

	sanctionID := c.Param("id")
	if sanctionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "id manquant"})
		return
	}

	var payload struct {
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil || payload.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "requete invalide"})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	var contestable bool
	err = db.QueryRow(`SELECT not_contestable FROM sanctions WHERE sanction_id = ? AND user_id = ?`, sanctionID, uid).Scan(&contestable)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "sanction inconnue"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	if contestable {
		uuidV7, _ := uuid.NewV7()
		appealID := uuidV7.String()
		_, err = db.Exec(`INSERT INTO sanction_appeals (appeal_id,sanction_id,user_id,message) VALUES (?,?,?,?)`, appealID, sanctionID, uid, payload.Message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "appeal_id": appealID})
	} else {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "cette sanction n'est pas contestable"})
	}
}

func ListAppealsHandler(c *gin.Context) {
	modID, _ := c.Get("user_id")
	moderatorID, _ := modID.(string)

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, moderatorID, "list_appeals", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	rows, err := db.Query(`SELECT appeal_id,sanction_id,user_id,message,status,response,created_at FROM sanction_appeals ORDER BY created_at DESC`)
	if err != nil {
		utils.LogActivity(c, moderatorID, "list_appeals", false, "query error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer rows.Close()

	appeals := make([]gin.H, 0)
	for rows.Next() {
		var id, sid, uid, message, status, response string
		var created time.Time
		if err := rows.Scan(&id, &sid, &uid, &message, &status, &response, &created); err == nil {
			appeals = append(appeals, gin.H{
				"appeal_id":   id,
				"sanction_id": sid,
				"user_id":     uid,
				"message":     message,
				"status":      status,
				"response":    response,
				"created_at":  created,
			})
		}
	}

	utils.LogActivity(c, moderatorID, "list_appeals", true, "")
	c.JSON(http.StatusOK, gin.H{"success": true, "appeals": appeals})
}
