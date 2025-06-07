package tools

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"toolcenter/config"
	"toolcenter/utils"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

const toolImageDir = "/var/www/toolcenter/storage/tools_images"
const toolImageRelPath = "/tools_images/"

func rnd() string {
	b := make([]byte, 16)
	_, _ = io.ReadFull(rand.Reader, b)
	return hex.EncodeToString(b)
}

func SubmitToolHandler(c *gin.Context) {
	uid, _, _, err := utils.Check(c, utils.CheckOpts{
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
		c.JSON(code, gin.H{"success": false, "message": err.Error()})
		return
	}

	title := c.PostForm("title")
	desc := c.PostForm("description")
	category := c.PostForm("category")
	url := c.PostForm("url")
	tagsRaw := c.PostForm("tags")

	var lastPosted sql.NullTime
	dbCheck, err := config.OpenDB()
	if err == nil {
		defer dbCheck.Close()
		err = dbCheck.QueryRow("SELECT last_tool_posted FROM users WHERE user_id = ?", uid).Scan(&lastPosted)
	}
	if err == nil && lastPosted.Valid {
		cooldown := 24 * 60 * 60
		remaining := int(time.Until(lastPosted.Time.Add(time.Duration(cooldown) * time.Second)).Seconds())
		if remaining > 0 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success":             false,
				"message":             "Vous devez attendre avant de soumettre un nouvel outil.",
				"retry_after_seconds": remaining,
			})
			return
		}
	}

	if title == "" || desc == "" || category == "" || url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "champs manquants"})
		return
	}

	var imageRel string
	if file, header, err := c.Request.FormFile("image"); err == nil {
		defer file.Close()
		ext := strings.ToLower(filepath.Ext(header.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
			c.JSON(http.StatusUnsupportedMediaType, gin.H{"success": false, "message": "format non supporté"})
			return
		}
		tmp := filepath.Join(os.TempDir(), rnd()+ext)
		out, err := os.Create(tmp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false})
			return
		}
		if _, err = io.Copy(out, file); err != nil {
			out.Close()
			os.Remove(tmp)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false})
			return
		}
		out.Close()
		img, err := imaging.Open(tmp)
		os.Remove(tmp)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "image invalide"})
			return
		}
		img = imaging.Fill(img, 1200, 630, imaging.Center, imaging.Lanczos)
		_ = os.MkdirAll(toolImageDir, 0755)
		filename := rnd() + ".webp"
		final := filepath.Join(toolImageDir, filename)
		fp, err := os.Create(final)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false})
			return
		}
		if err = webp.Encode(fp, img, &webp.Options{Lossless: true}); err != nil {
			fp.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false})
			return
		}
		fp.Close()
		imageRel = toolImageRelPath + filename
	}

	db, err := config.OpenDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erreur DB."})
		return
	}
	defer db.Close()

	uuidV7, _ := uuid.NewV7()
	toolID := uuidV7.String()

	_, err = db.Exec(`INSERT INTO tools (tool_id, user_id, title, description, content_url, thumbnail_url, status) VALUES (?, ?, ?, ?, ?, ?, 'Moderated')`,
		toolID, uid, title, desc, url, imageRel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erreur DB."})
		return
	}

	tags := []string{category}
	if tagsRaw != "" {
		for _, t := range strings.Split(tagsRaw, ",") {
			if tt := strings.TrimSpace(t); tt != "" {
				tags = append(tags, tt)
			}
		}
	}
	for _, tag := range tags {
		var tagID int
		err := db.QueryRow(`SELECT tag_id FROM tags WHERE name = ?`, tag).Scan(&tagID)
		if err == sql.ErrNoRows {
			r, err2 := db.Exec(`INSERT INTO tags (name) VALUES (?)`, tag)
			if err2 != nil {
				continue
			}
			id, _ := r.LastInsertId()
			tagID = int(id)
		} else if err != nil {
			continue
		}
		_, _ = db.Exec(`INSERT INTO tool_tags (tool_id, tag_id) VALUES (?, ?)`, toolID, tagID)
	}

	_, _ = db.Exec(`UPDATE users SET last_tool_posted = NOW() WHERE user_id = ?`, uid)

	base := config.Get().URLweb
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"tool_id": toolID,
		"image_url": func() string {
			if imageRel != "" {
				return base + imageRel
			}
			return ""
		}(),
	})
}
