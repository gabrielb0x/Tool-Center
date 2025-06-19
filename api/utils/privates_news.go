package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"toolcenter/config"
)

var lastToken string
var lastTokenExpiration time.Time

// Pour ajouter des tags à un article, il suffit d'ajouter les chaînes désirées dans le tableau.
// Exemple : []string{"pinned", "urgent", "important", "todo", "inprogress", "done", "obsolete"},

func loadPrivateArticles() ([]gin.H, error) {
	exePath, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	filePath := filepath.Join(exePath, "utils", "privates_articles.json")
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var articles []gin.H
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&articles); err != nil {
		return nil, err
	}
	return articles, nil
}

func generateToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func PrivateNewsHandler(c *gin.Context) {
	var req struct {
		Password    string      `json:"password"`
		Token       string      `json:"token"`
		BrowserInfo interface{} `json:"browserInfo"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Requête invalide"})
		return
	}

	articles, err := loadPrivateArticles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erreur de lecture des articles privés"})
		return
	}

	if req.Password != "" {
		if req.Password == config.Get().PrivateNewsPassword {
			lastToken = generateToken(32)
			lastTokenExpiration = time.Now().Add(24 * time.Hour)
			c.JSON(http.StatusOK, gin.H{
				"success":  true,
				"token":    lastToken,
				"articles": articles,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Mot de passe incorrect"})
		return
	}

	if req.Token != "" {
		if req.Token == lastToken && time.Now().Before(lastTokenExpiration) {
			c.JSON(http.StatusOK, gin.H{
				"success":  true,
				"articles": articles,
			})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Token invalide ou expiré"})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Aucun identifiant fourni"})
}
