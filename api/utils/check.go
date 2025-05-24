package utils

import (
	"database/sql"
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

func Check(c *gin.Context, o CheckOpts) (string, bool, string, error) {
	var bearer string

	if o.RequireToken {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			return "", false, "", ErrMissingToken
		}
		bearer = strings.TrimPrefix(h, "Bearer ")
	}

	db, err := config.OpenDB()
	if err != nil {
		return "", false, "", err
	}
	defer db.Close()

	var (
		uid           string
		expires       sql.NullTime
		verifiedAt    sql.NullTime
		accountStatus string
	)

	if o.RequireToken {
		err = db.QueryRow(`
			SELECT u.user_id, ut.expires_at, u.email_verified_at, u.account_status
			FROM user_tokens ut
			JOIN users u ON u.user_id = ut.user_id
			WHERE ut.token = ? LIMIT 1`, bearer).
			Scan(&uid, &expires, &verifiedAt, &accountStatus)
	} else {
		var payload struct {
			UserID string `json:"user_id"`
		}
		_ = c.ShouldBindJSON(&payload)
		uid = payload.UserID
		err = db.QueryRow(`SELECT email_verified_at, account_status FROM users WHERE user_id = ?`, uid).
			Scan(&verifiedAt, &accountStatus)
	}

	switch {
	case err == sql.ErrNoRows:
		return "", false, "", ErrInvalidToken
	case err != nil:
		return "", false, "", err
	}

	if o.RequireToken && expires.Valid && time.Now().After(expires.Time) {
		return "", false, "", ErrExpiredToken
	}
	if o.RequireVerified && !verifiedAt.Valid {
		return "", false, "", ErrEmailNotVerified
	}
	if o.RequireNotBanned && accountStatus == "Banned" {
		return "", false, "", ErrAccountBanned
	}

	if o.UpdateLastLogin {
		_, _ = db.Exec(`UPDATE users SET last_login = NOW() WHERE user_id = ?`, uid)
	}

	return uid, verifiedAt.Valid, accountStatus, nil
}
