package admin

import (
	"net/http"
	"time"

	"toolcenter/config"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func percentChange(curr, prev int) int {
	if prev == 0 {
		if curr == 0 {
			return 0
		}
		return 100
	}
	return int(float64(curr-prev) / float64(prev) * 100)
}

func StatsHandler(c *gin.Context) {
	db, err := config.OpenDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	var total, moderators, banned int
	_ = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&total)
	_ = db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'Moderator'").Scan(&moderators)
	_ = db.QueryRow("SELECT COUNT(*) FROM users WHERE account_status = 'Banned'").Scan(&banned)

	var monthUsers, prevMonthUsers int
	_ = db.QueryRow("SELECT COUNT(*) FROM users WHERE created_at >= DATE_FORMAT(NOW(), '%Y-%m-01')").Scan(&monthUsers)
	_ = db.QueryRow("SELECT COUNT(*) FROM users WHERE created_at >= DATE_SUB(DATE_FORMAT(NOW(), '%Y-%m-01'), INTERVAL 1 MONTH) AND created_at < DATE_FORMAT(NOW(), '%Y-%m-01')").Scan(&prevMonthUsers)

	var newUsers, prevNewUsers int
	_ = db.QueryRow("SELECT COUNT(*) FROM users WHERE created_at >= NOW() - INTERVAL 7 DAY").Scan(&newUsers)
	_ = db.QueryRow("SELECT COUNT(*) FROM users WHERE created_at >= NOW() - INTERVAL 14 DAY AND created_at < NOW() - INTERVAL 7 DAY").Scan(&prevNewUsers)

	var bannedMonth, bannedPrevMonth int
	_ = db.QueryRow("SELECT COUNT(*) FROM users WHERE account_status = 'Banned' AND created_at >= DATE_FORMAT(NOW(), '%Y-%m-01')").Scan(&bannedMonth)
	_ = db.QueryRow("SELECT COUNT(*) FROM users WHERE account_status = 'Banned' AND created_at >= DATE_SUB(DATE_FORMAT(NOW(), '%Y-%m-01'), INTERVAL 1 MONTH) AND created_at < DATE_FORMAT(NOW(), '%Y-%m-01')").Scan(&bannedPrevMonth)

	c.JSON(http.StatusOK, gin.H{
		"success":           true,
		"totalUsers":        total,
		"userGrowth":        percentChange(monthUsers, prevMonthUsers),
		"newUsers":          newUsers,
		"newUsersGrowth":    percentChange(newUsers, prevNewUsers),
		"bannedUsers":       banned,
		"bannedUsersChange": percentChange(bannedMonth, bannedPrevMonth),
		"moderators":        moderators,
		"timestamp":         time.Now(),
	})
}
