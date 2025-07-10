package utils

import (
    "database/sql"
    "time"
)

var statusSteps = []string{"Good", "Limited", "Very Limited", "At Risk", "Banned"}

// DecreaseStatus lowers the user account_status by n steps.
func DecreaseStatus(db *sql.DB, userID string, n int) (string, error) {
    var current string
    if err := db.QueryRow(`SELECT account_status FROM users WHERE user_id=?`, userID).Scan(&current); err != nil {
        return "", err
    }
    idx := 0
    for i, s := range statusSteps {
        if s == current {
            idx = i
            break
        }
    }
    idx += n
    if idx >= len(statusSteps) {
        idx = len(statusSteps) - 1
    }
    newStatus := statusSteps[idx]
    if newStatus != current {
        if _, err := db.Exec(`UPDATE users SET account_status=? WHERE user_id=?`, newStatus, userID); err != nil {
            return current, err
        }
    }
    return newStatus, nil
}

// QueueEmail inserts a message into the email_queue table.
func QueueEmail(db *sql.DB, to, subject, body string) error {
    _, err := db.Exec(`INSERT INTO email_queue (to_email, subject, body) VALUES (?, ?, ?)`, to, subject, body)
    return err
}

// Helper to compute if a sanction is active given an optional end time.
func SanctionActive(end sql.NullTime) bool {
    if !end.Valid {
        return true
    }
    return time.Now().Before(end.Time)
}

