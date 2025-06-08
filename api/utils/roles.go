package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireRole(roles ...string) gin.HandlerFunc {
	roleSet := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		roleSet[r] = struct{}{}
	}
	return func(c *gin.Context) {
		uid, _, userRole, _, err := Check(c, CheckOpts{
			RequireToken:     true,
			RequireVerified:  true,
			RequireNotBanned: true,
		})
		if err != nil {
			code := http.StatusInternalServerError
			switch err {
			case ErrMissingToken, ErrInvalidToken, ErrExpiredToken:
				code = http.StatusUnauthorized
			case ErrEmailNotVerified, ErrAccountBanned:
				code = http.StatusForbidden
			}
			c.JSON(code, gin.H{"success": false, "message": err.Error()})
			c.Abort()
			return
		}
		if _, ok := roleSet[userRole]; !ok {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "access denied"})
			c.Abort()
			return
		}
		c.Set("user_id", uid)
		c.Set("role", userRole)
		c.Next()
	}
}
