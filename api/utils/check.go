package utils

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"

        "toolcenter/config"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type CheckOpts struct {
	RequireToken     bool
	RequireVerified  bool
	RequireNotBanned bool
	UpdateLastLogin  bool
}

var (
	ErrMissingToken     = errors.New("vous devez vous connecter")
	ErrInvalidToken     = errors.New("votre session a expiré, veuillez vous reconnecter")
	ErrExpiredToken     = errors.New("reconnectez-vous, la session a expiré")
	ErrEmailNotVerified = errors.New("vous devez vérifier votre adresse e-mail")
        ErrAccountBanned    = errors.New("compte banni")
)

func LiftExpiredBan(db *sql.DB, uid string) (bool, error) {
        cfg := config.Get()
        if !cfg.Moderation.AutoUnban {
                return false, nil
        }
        var end sql.NullTime
        err := db.QueryRow(`SELECT end_date FROM moderation_actions WHERE user_id=? AND action_type='Ban' ORDER BY action_date DESC LIMIT 1`, uid).Scan(&end)
        if err != nil {
                if err == sql.ErrNoRows {
                        return false, nil
                }
                return false, err
        }
        if end.Valid && time.Now().After(end.Time) {
                if _, err := db.Exec(`UPDATE users SET account_status='Good' WHERE user_id=?`, uid); err == nil {
                        return true, nil
                }
        }
        return false, nil
}

func Check(c *gin.Context, o CheckOpts) (string, bool, string, string, error) {
	var bearer string

	if o.RequireToken {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			return "", false, "", "", ErrMissingToken
		}
		bearer = strings.TrimPrefix(h, "Bearer ")
		hash := sha256.Sum256([]byte(bearer))
		bearer = hex.EncodeToString(hash[:])
	}

	db, err := config.OpenDB()
	if err != nil {
		return "", false, "", "", err
	}
	defer db.Close()

	var (
		uid           string
		expires       sql.NullTime
		verifiedAt    sql.NullTime
		accountStatus string
		role          string
	)

	if o.RequireToken {
		err = db.QueryRow(`
                        SELECT u.user_id, ut.expires_at, u.email_verified_at, u.account_status, u.role
                        FROM user_tokens ut
                        JOIN users u ON u.user_id = ut.user_id
                        WHERE ut.token = ? LIMIT 1`, bearer).
			Scan(&uid, &expires, &verifiedAt, &accountStatus, &role)
	} else {
		if v, ok := c.Get("user_id_override"); ok {
			if s, ok := v.(string); ok {
				uid = s
			}
		}
		if uid == "" {
			var payload struct {
				UserID string `json:"user_id"`
			}
			_ = c.ShouldBindJSON(&payload)
			uid = payload.UserID
		}
		err = db.QueryRow(`SELECT role, email_verified_at, account_status FROM users WHERE user_id = ?`, uid).
			Scan(&role, &verifiedAt, &accountStatus)
	}

	switch {
	case err == sql.ErrNoRows:
		return "", false, "", "", ErrInvalidToken
	case err != nil:
		return "", false, "", "", err
	}

        if o.RequireToken && expires.Valid && time.Now().After(expires.Time) {
                return "", false, "", "", ErrExpiredToken
        }
        if o.RequireVerified && !verifiedAt.Valid {
                return "", false, "", "", ErrEmailNotVerified
        }

        if accountStatus == "Banned" {
                if lifted, _ := LiftExpiredBan(db, uid); lifted {
                        accountStatus = "Good"
                }
        }
        if o.RequireNotBanned && accountStatus == "Banned" {
                return "", false, "", "", ErrAccountBanned
        }

	if o.UpdateLastLogin {
		_, _ = db.Exec(`UPDATE users SET last_login = NOW() WHERE user_id = ?`, uid)
	}

	return uid, verifiedAt.Valid, role, accountStatus, nil
}
