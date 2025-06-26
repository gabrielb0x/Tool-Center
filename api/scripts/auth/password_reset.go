package auth

import (
    "crypto/rand"
    "crypto/sha256"
    "database/sql"
    "encoding/hex"
    "fmt"
    "net/http"
    "time"

    "toolcenter/config"
    "toolcenter/utils"

    "github.com/gin-gonic/gin"
    _ "github.com/go-sql-driver/mysql"
    "golang.org/x/crypto/bcrypt"
    "gopkg.in/gomail.v2"
)

type resetRequest struct {
    Email          string `json:"email"`
    TurnstileToken string `json:"turnstile_token"`
}

type resetConfirmRequest struct {
    Token       string `json:"token"`
    NewPassword string `json:"new_password"`
}

func randToken(n int) string {
    b := make([]byte, n)
    rand.Read(b)
    return hex.EncodeToString(b)
}

func hashTok(t string) string {
    h := sha256.Sum256([]byte(t))
    return hex.EncodeToString(h[:])
}

func RequestPasswordResetHandler(c *gin.Context) {
    var req resetRequest
    if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" || req.TurnstileToken == "" {
        utils.LogActivity(c, "", "pwd_reset_request", false, "bad request")
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Requête invalide."})
        return
    }
    ok, err := utils.VerifyTurnstile(req.TurnstileToken, config.Get().Turnstile.SignInSecret, c.ClientIP())
    if err != nil || !ok {
        utils.LogActivity(c, "", "pwd_reset_request", false, "captcha invalid")
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Captcha invalide"})
        return
    }
    db, err := config.OpenDB()
    if err != nil {
        utils.LogActivity(c, "", "pwd_reset_request", false, "db error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erreur DB."})
        return
    }
    defer db.Close()

    var uid, username string
    err = db.QueryRow("SELECT user_id, username FROM users WHERE email=?", req.Email).Scan(&uid, &username)
    if err == sql.ErrNoRows {
        // avoid enumeration
        c.JSON(http.StatusOK, gin.H{"success": true})
        return
    }
    if err != nil {
        utils.LogActivity(c, "", "pwd_reset_request", false, "query error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erreur interne."})
        return
    }

    token := randToken(32)
    tokenHash := hashTok(token)
    expires := time.Now().Add(time.Duration(config.Get().PasswordReset.TokenExpiryMinutes) * time.Minute)
    _, err = db.Exec(`INSERT INTO password_resets (user_id, token, expires_at) VALUES (?,?,?)`, uid, tokenHash, expires)
    if err != nil {
        utils.LogActivity(c, uid, "pwd_reset_request", false, "insert error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    go sendPasswordResetEmail(req.Email, username, token)
    utils.LogActivity(c, uid, "pwd_reset_request", true, "")
    c.JSON(http.StatusOK, gin.H{"success": true})
}

func sendPasswordResetEmail(email, username, token string) {
    cfg := config.Get()
    link := fmt.Sprintf("%s/reset.html?token=%s", cfg.URLweb, token)
    body := fmt.Sprintf("Bonjour %s,<br/>Cliquez sur le lien suivant pour réinitialiser votre mot de passe : <a href=\"%s\">Réinitialiser</a><br/>Ce lien expirera dans %d minutes.", username, link, cfg.PasswordReset.TokenExpiryMinutes)
    msg := gomail.NewMessage()
    msg.SetHeader("From", cfg.Email.From)
    msg.SetHeader("To", email)
    msg.SetHeader("Subject", "Réinitialisation de votre mot de passe")
    msg.SetBody("text/html", body)
    d := gomail.NewDialer(cfg.Email.Host, cfg.Email.Port, cfg.Email.Username, cfg.Email.Password)
    d.SSL = true
    _ = d.DialAndSend(msg)
}

func ResetPasswordHandler(c *gin.Context) {
    var req resetConfirmRequest
    if err := c.ShouldBindJSON(&req); err != nil || req.Token == "" || len(req.NewPassword) < 7 {
        utils.LogActivity(c, "", "pwd_reset", false, "bad request")
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Requête invalide."})
        return
    }
    tokenHash := hashTok(req.Token)
    db, err := config.OpenDB()
    if err != nil {
        utils.LogActivity(c, "", "pwd_reset", false, "db error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    defer db.Close()

    var uid string
    var expires time.Time
    err = db.QueryRow(`SELECT user_id, expires_at FROM password_resets WHERE token=?`, tokenHash).Scan(&uid, &expires)
    if err == sql.ErrNoRows {
        c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Token invalide"})
        return
    }
    if err != nil {
        utils.LogActivity(c, "", "pwd_reset", false, "query error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    if time.Now().After(expires) {
        db.Exec(`DELETE FROM password_resets WHERE token=?`, tokenHash)
        c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Token expiré"})
        return
    }
    pwHash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
    _, err = db.Exec(`UPDATE users SET password_hash=? WHERE user_id=?`, pwHash, uid)
    if err != nil {
        utils.LogActivity(c, uid, "pwd_reset", false, "update error")
        c.JSON(http.StatusInternalServerError, gin.H{"success": false})
        return
    }
    db.Exec(`DELETE FROM password_resets WHERE user_id=?`, uid)
    utils.LogActivity(c, uid, "pwd_reset", true, "")
    c.JSON(http.StatusOK, gin.H{"success": true})
}

