package admin

import (
	"fmt"
	"net/http"
	"time"
	"toolcenter/config"
	"toolcenter/utils"

	"github.com/gin-gonic/gin"
)

func buildDecisionEmail(username, msg string) string {
	return fmt.Sprintf(`
    <!DOCTYPE html>
    <html lang="fr">
    <head>
    <meta charset="UTF-8">
    <title>Décision de contestation Tool Center</title>
    <link href="https://fonts.googleapis.com/css2?family=Poppins:wght@300;400;500;600;700&display=swap" rel="stylesheet">
    <style>
    body{
        margin:0;
        padding:0;
        background-color:#121212;
        font-family:'Poppins',sans-serif;
        color:#e0e0e0;
    }
    .container{
        max-width:600px;
        margin:30px auto;
        background:#1e1e1e;
        border-radius:12px;
        overflow:hidden;
        box-shadow:0 10px 30px rgba(0,0,0,0.3);
        border:1px solid #333;
    }
    .header{
        padding:30px;
        text-align:center;
        background:linear-gradient(135deg,#3000FF 0%%,#6200EA 100%%);
    }
    .logo{
        max-width:180px;
        height:auto;
        transition:transform .3s ease;
    }
    .logo:hover{transform:scale(1.05);}
    .content{padding:30px;}
    h1{
        color:#fff;
        font-size:28px;
        margin-bottom:20px;
        font-weight:600;
        text-align:center;
    }
    p{
        font-size:16px;
        line-height:1.6;
        margin-bottom:20px;
        color:#b0b0b0;
    }
    .highlight{color:#fff;font-weight:500;}
    .divider{
        height:1px;
        background:linear-gradient(90deg,transparent,#333,transparent);
        margin:20px 0;
    }
    .footer{
        padding:20px;
        text-align:center;
        background:#121212;
        border-top:1px solid #333;
        font-size:12px;
        color:#666;
    }
    .security-note{
        background:rgba(48,0,255,.1);
        padding:15px;
        border-radius:8px;
        border-left:4px solid #3000FF;
        font-size:14px;
        color:#fff;
        margin-top:25px;
    }
    @media only screen and (max-width:600px){
        .container{margin:0;border-radius:0;}
        .content{padding:25px;}
    }
    </style>
    </head>
    <body>
    <div class="container">
        <div class="header">
        <img src="https://tool-center.fr/assets/tc_logo.webp" alt="Tool Center Logo"
            style="border-radius:9999px;display:block;margin:auto;" class="logo">
        </div>
        <div class="content">
        <h1>Décision de contestation</h1>
        <p>Bonjour <span class="highlight">%s</span>,</p>
        <p>%s</p>
        <div class="divider"></div>
        <div class="security-note">
            <strong>Conseil&nbsp;:</strong> Si vous avez des questions concernant cette décision, vous pouvez répondre à ce mail ou contacter le support Tool Center.<br>
            Ne partagez jamais d'informations sensibles par email.
        </div>
        </div>
        <div class="footer">
        <p>© Tool Center %d. Tous droits réservés.</p>
        </div>
    </div>
    </body>
    </html>
    `, username, msg, time.Now().Year())
}

func ListAppealsHandler(c *gin.Context) {
	db, err := config.OpenDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	rows, err := db.Query(`SELECT appeal_id, action_id, user_id, message, status, created_at FROM sanction_appeals ORDER BY created_at DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer rows.Close()

	appeals := make([]gin.H, 0)
	for rows.Next() {
		var id, action, uid, msg, status string
		var created string
		if err := rows.Scan(&id, &action, &uid, &msg, &status, &created); err == nil {
			appeals = append(appeals, gin.H{
				"appeal_id":  id,
				"action_id":  action,
				"user_id":    uid,
				"message":    msg,
				"status":     status,
				"created_at": created,
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "appeals": appeals})
}

func ReviewAppealHandler(c *gin.Context) {
	adminID := c.GetString("user_id")
	appealID := c.Param("id")

	var req struct {
		Approve bool   `json:"approve"`
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	var actionID, userID string
	err = db.QueryRow(`SELECT action_id, user_id FROM sanction_appeals WHERE appeal_id=? AND status='Pending'`, appealID).Scan(&actionID, &userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false})
		return
	}

	status := "Rejected"
	if req.Approve {
		status = "Approved"
	}
	var prev string
	_ = db.QueryRow(`SELECT previous_status FROM moderation_actions WHERE action_id=?`, actionID).Scan(&prev)
	if prev != "" {
		_, _ = db.Exec(`UPDATE users SET account_status=? WHERE user_id=?`, prev, userID)
	}
	_, _ = db.Exec(`UPDATE moderation_actions SET end_date=NOW() WHERE action_id=?`, actionID)
	_, _ = db.Exec(`UPDATE sanction_appeals SET status=?, reviewed_by=?, reviewed_at=NOW() WHERE appeal_id=?`, status, adminID, appealID)

	var username, email string
	_ = db.QueryRow(`SELECT username,email FROM users WHERE user_id=?`, userID).Scan(&username, &email)
	if email != "" {
		body := buildDecisionEmail(username, req.Message)
		_ = utils.QueueEmail(db, email, "Contestation", body)
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
