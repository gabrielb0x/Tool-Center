package utils

import (
    "time"
    "toolcenter/config"
)

var statusOrder = []string{"Good", "Limited", "Very Limited", "At Risk", "Banned"}

func indexOfStatus(s string) int {
    for i, v := range statusOrder {
        if v == s {
            return i
        }
    }
    return -1
}

// ApplyStatusDrop decreases a user's account_status by the given number of steps.
// If the resulting status is "Banned", a temporary ban is recorded.
func ApplyStatusDrop(userID string, drop int) (string, error) {
    db, err := config.OpenDB()
    if err != nil {
        return "", err
    }
    defer db.Close()

    var current string
    if err := db.QueryRow(`SELECT account_status FROM users WHERE user_id=?`, userID).Scan(&current); err != nil {
        return "", err
    }
    idx := indexOfStatus(current)
    if idx == -1 {
        idx = 0
    }
    newIdx := idx + drop
    if newIdx >= len(statusOrder)-1 {
        // Ban user
        banDuration := config.Get().SpamProtection.BanDays
        if banDuration <= 0 {
            banDuration = 7
        }
        end := time.Now().Add(time.Duration(banDuration) * 24 * time.Hour)
        if _, err := db.Exec(`UPDATE users SET account_status='Banned' WHERE user_id=?`, userID); err != nil {
            return "", err
        }
        _, _ = db.Exec(`INSERT INTO moderation_actions (user_id, action_type, reason, start_date, end_date) VALUES (?, 'Ban', 'Spam', NOW(), ?)`, userID, end)
        return "Banned", nil
    }
    newStatus := statusOrder[newIdx]
    _, err = db.Exec(`UPDATE users SET account_status=? WHERE user_id=?`, newStatus, userID)
    return newStatus, err
}
