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

func nStr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}
func nTime(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

func MeHandler(c *gin.Context) {
	uid, _, _, _, err := utils.Check(c, utils.CheckOpts{
		RequireToken:     true,
		RequireVerified:  true,
		RequireNotBanned: true,
		UpdateLastLogin:  true,
	})
	if err != nil {
		code := http.StatusInternalServerError
		switch err {
		case utils.ErrMissingToken, utils.ErrInvalidToken, utils.ErrExpiredToken:
			code = http.StatusUnauthorized
		case utils.ErrEmailNotVerified, utils.ErrAccountBanned:
			code = http.StatusForbidden
		}
		utils.LogActivity(c, "", "me", false, err.Error())
		c.JSON(code, gin.H{"success": false, "message": err.Error()})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, uid, "me", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erreur DB."})
		return
	}
	defer db.Close()

	var (
		username, role, status               string
		email                                sql.NullString
		avatar, banner, bio                  sql.NullString
		createdAt, updatedAt                 time.Time
		userChg, mailChg, avaChg, banChg     sql.NullTime
		passChg, lastLogin, lastPost, lastUp sql.NullTime
		blueCheck                            bool
	)

	const qry = `
	SELECT username,email,avatar_url,banner_url,role,account_status,bio,
		   created_at,updated_at,
		   username_changed_at,email_changed_at,avatar_changed_at,banner_changed_at,
		   password_changed_at,last_login,last_tool_posted,last_tool_updated,
		   is_verified
	FROM users WHERE user_id = ? LIMIT 1`

	if err = db.QueryRow(qry, uid).Scan(
		&username, &email, &avatar, &banner, &role, &status, &bio,
		&createdAt, &updatedAt,
		&userChg, &mailChg, &avaChg, &banChg,
		&passChg, &lastLogin, &lastPost, &lastUp,
		&blueCheck,
	); err != nil {
		utils.LogActivity(c, uid, "me", false, "query error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erreur interne."})
		return
	}

	var st struct{ Tools, Cmts, LG, LR, Fav int }
	_ = db.QueryRow(`
SELECT tools_posted_count, comments_count, likes_given, likes_received, favorites_count
FROM user_stats WHERE user_id = ?`, uid).
		Scan(&st.Tools, &st.Cmts, &st.LG, &st.LR, &st.Fav)

	sanctions := make([]gin.H, 0)

	if status != "Good" {
		rows, err := db.Query(`
            SELECT warn_id, reason, warn_date
            FROM user_warns
            WHERE user_id = ?
            ORDER BY warn_date DESC`, uid)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var (
					warnID   int
					reason   string
					warnDate time.Time
				)
				if err = rows.Scan(&warnID, &reason, &warnDate); err == nil {
					sanctions = append(sanctions, gin.H{
						"warn_id":   warnID,
						"reason":    reason,
						"warn_date": warnDate,
					})
				}
			}
		} else {
			utils.LogActivity(c, uid, "me", false, "warns query error")
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erreur interne."})
			return
		}
	}

	resp := gin.H{
		"user_id":             uid,
		"username":            username,
		"email":               nStr(email),
		"avatar_url":          nStr(avatar),
		"banner_url":          nStr(banner),
		"bio":                 nStr(bio),
		"is_verified":         blueCheck,
		"account_status":      status,
		"created_at":          createdAt,
		"updated_at":          updatedAt,
		"username_changed_at": nTime(userChg),
		"email_changed_at":    nTime(mailChg),
		"avatar_changed_at":   nTime(avaChg),
		"banner_changed_at":   nTime(banChg),
		"password_changed_at": nTime(passChg),
		"last_login":          nTime(lastLogin),
		"last_tool_posted":    nTime(lastPost),
		"last_tool_updated":   nTime(lastUp),
		"stats": gin.H{
			"tools_posted":   st.Tools,
			"comments":       st.Cmts,
			"likes_given":    st.LG,
			"likes_received": st.LR,
			"favorites":      st.Fav,
		},
	}
	if role != "User" {
		resp["role"] = role
	}
	if len(sanctions) > 0 {
		resp["sanctions"] = sanctions
	}

	utils.LogActivity(c, uid, "me", true, "")
	c.JSON(http.StatusOK, gin.H{"success": true, "user": resp})
}
