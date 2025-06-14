package user

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"

	"toolcenter/config"
	"toolcenter/utils"
)

func UploadAvatar(c *gin.Context) {
	uid, _, _, _, err := utils.Check(c, utils.CheckOpts{
		RequireToken:     true,
		RequireVerified:  true,
		RequireNotBanned: true,
	})
	if err != nil {
		code := http.StatusInternalServerError
		switch err {
		case utils.ErrMissingToken, utils.ErrInvalidToken, utils.ErrExpiredToken:
			code = http.StatusUnauthorized
		case utils.ErrEmailNotVerified, utils.ErrAccountBanned:
			code = http.StatusForbidden
		}
		utils.LogActivity(c, uid, "upload_avatar", false, err.Error())
		c.JSON(code, gin.H{"success": false, "message": err.Error()})
		return
	}

	db, err := config.OpenDB()
	if err != nil {
		utils.LogActivity(c, uid, "upload_avatar", false, "db open error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	defer db.Close()

	cooldown := time.Duration(config.Get().AvatarCooldownHours) * time.Hour
	var lastChangedAt time.Time
	err = db.QueryRow(`SELECT avatar_changed_at FROM users WHERE user_id = ?`, uid).Scan(&lastChangedAt)
	if err == nil && !lastChangedAt.IsZero() && time.Since(lastChangedAt) < cooldown {
		retryAt := lastChangedAt.Add(cooldown)
		utils.LogActivity(c, uid, "upload_avatar", false, "cooldown")
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success":  false,
			"message":  "Vous ne pouvez changer votre photo de profil qu'une fois par jour.",
			"retry_at": retryAt.Format(time.RFC3339),
		})
		return
	}

	if c.PostForm("avatar") == "delete" {
		path := "/var/www/toolcenter/storage/avatars/" + uid + ".webp"
		_ = os.Remove(path)
		_, _ = db.Exec(`UPDATE users SET avatar_url = NULL, avatar_changed_at = NOW() WHERE user_id = ?`, uid)
		utils.LogActivity(c, uid, "upload_avatar", true, "delete")
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 5<<20)
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		utils.LogActivity(c, uid, "upload_avatar", false, "file missing")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "fichier manquant"})
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		utils.LogActivity(c, uid, "upload_avatar", false, "bad format")
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"success": false, "message": "format non supporté"})
		return
	}

	tmp := "/tmp/" + rnd() + ext
	out, err := os.Create(tmp)
	if err != nil {
		utils.LogActivity(c, uid, "upload_avatar", false, "tmp create")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	if _, err = io.Copy(out, file); err != nil {
		out.Close()
		os.Remove(tmp)
		utils.LogActivity(c, uid, "upload_avatar", false, "tmp copy")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	out.Close()

	img, err := imaging.Open(tmp)
	os.Remove(tmp)
	if err != nil {
		utils.LogActivity(c, uid, "upload_avatar", false, "invalid image")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "image invalide"})
		return
	}
	img = imaging.Fill(img, 512, 512, imaging.Center, imaging.Lanczos)

	dir := "/var/www/toolcenter/storage/avatars"
	_ = os.MkdirAll(dir, 0755)
	finalPath := dir + "/" + uid + ".webp"
	fp, err := os.Create(finalPath)
	if err != nil {
		utils.LogActivity(c, uid, "upload_avatar", false, "create final")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	if err = webp.Encode(fp, img, &webp.Options{Lossless: true}); err != nil {
		fp.Close()
		utils.LogActivity(c, uid, "upload_avatar", false, "encode error")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	fp.Close()

	rel := "/avatars/" + uid + ".webp"
	_, _ = db.Exec(`UPDATE users SET avatar_url = ?, avatar_changed_at = NOW() WHERE user_id = ?`, rel, uid)
	utils.LogActivity(c, uid, "upload_avatar", true, "")

	base := config.Get().URLweb
	c.JSON(http.StatusOK, gin.H{"success": true, "avatar_url": base + rel})
}

func rnd() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
