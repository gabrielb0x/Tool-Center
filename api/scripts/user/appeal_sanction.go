package user

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"toolcenter/config"
	"toolcenter/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type appealRequest struct {
	Message string `json:"message"`
}

func buildAppealEmail(username, appealID string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="fr">
<head>
<meta charset="UTF-8">
<title>Contestation de sanction Tool Center</title>
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
    <h1>Contestation enregistrée</h1>
    <p>Bonjour <span class="highlight">%s</span>,</p>
    <p><strong>ID de contestation&nbsp;:</strong> %s</p>
    <p>Votre contestation de sanction a bien été reçue.<br>
    Notre équipe va l'examiner dans les plus brefs délais et vous tiendra informé(e) par email.</p>
    <div class="divider"></div>
    <p>Si vous n'êtes pas à l'origine de cette demande, vous pouvez ignorer cet email.</p>
    </div>
    <div class="footer">
    <p>© Tool Center %d. Tous droits réservés.</p>
    </div>
</div>
</body>
</html>
`, username, appealID, time.Now().Year())
}

func AppealSanctionHandler(c *gin.Context) {
	uid, _, _, _, err := utils.Check(c, utils.CheckOpts{RequireToken: true})
	if err != nil {
		code := http.StatusInternalServerError
		if err == utils.ErrMissingToken || err == utils.ErrInvalidToken || err == utils.ErrExpiredToken {
			code = http.StatusUnauthorized
		}
		c.JSON(code, gin.H{"success": false, "message": err.Error()})
		return
	}

	sid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "id invalide"})
		return
	}

	var req appealRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "message manquant"})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	var exists int
	err = db.QueryRow(`SELECT COUNT(*) FROM moderation_actions WHERE action_id=? AND user_id=?`, sid, uid).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	if exists == 0 {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "sanction introuvable"})
		return
	}

	_ = db.QueryRow(`SELECT COUNT(*) FROM sanction_appeals WHERE action_id=? AND status='Pending'`, sid).Scan(&exists)
	if exists > 0 {
		c.JSON(http.StatusConflict, gin.H{"success": false, "message": "déjà contestée"})
		return
	}

	uuidV7, _ := uuid.NewV7()
	appealID := uuidV7.String()
	_, err = db.Exec(`INSERT INTO sanction_appeals (appeal_id, action_id, user_id, message) VALUES (?,?,?,?)`, appealID, sid, uid, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	var username, email string
	_ = db.QueryRow(`SELECT username,email FROM users WHERE user_id=?`, uid).Scan(&username, &email)
	if email != "" {
		body := buildAppealEmail(username, appealID)
		_ = utils.QueueEmail(db, email, "Contestation reçue", body)
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
