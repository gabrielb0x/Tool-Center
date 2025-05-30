package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"toolcenter/config"

	"github.com/gin-gonic/gin"
)

type DiscordEmbed struct {
	Description string `json:"description"`
	Color       int    `json:"color"`
	Timestamp   string `json:"timestamp"`
	Footer      struct {
		Text string `json:"text"`
	} `json:"footer"`
	Author struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"author"`
	Thumbnail struct {
		URL string `json:"url"`
	} `json:"thumbnail"`
	// Fields removed
}

type DiscordWebhookPayload struct {
	Username  string         `json:"username"`
	AvatarURL string         `json:"avatar_url"`
	Embeds    []DiscordEmbed `json:"embeds"`
	Content   string         `json:"content,omitempty"`
}

func DiscordWebhookNewsHandler(c *gin.Context) {
	var githubPayload struct {
		HeadCommit struct {
			Message string `json:"message"`
			URL     string `json:"url"`
			Author  struct {
				Name string `json:"name"`
			} `json:"author"`
			Added    []string `json:"added"`
			Removed  []string `json:"removed"`
			Modified []string `json:"modified"`
		} `json:"head_commit"`
		Pusher struct {
			Name string `json:"name"`
		} `json:"pusher"`
		Repository struct {
			FullName string `json:"full_name"`
			HTMLURL  string `json:"html_url"`
			Private  bool   `json:"private"`
		} `json:"repository"`
		Sender struct {
			AvatarURL string `json:"avatar_url"`
		} `json:"sender"`
	}

	if err := c.ShouldBindJSON(&githubPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid GitHub payload"})
		return
	}

	cfg := config.Get()

	diff := ""
	if len(githubPayload.HeadCommit.Added) > 0 || len(githubPayload.HeadCommit.Removed) > 0 || len(githubPayload.HeadCommit.Modified) > 0 {
	}
	if len(githubPayload.HeadCommit.Added) > 0 {
		diff += "### ✅ Fichiers ajoutés :\n"
		diff += "```diff\n"
		for _, f := range githubPayload.HeadCommit.Added {
			diff += fmt.Sprintf("+ %s\n", f)
		}
		diff += "```\n"
	}
	if len(githubPayload.HeadCommit.Removed) > 0 {
		diff += "### 🚫 Fichiers retirés :\n"
		diff += "```diff\n"
		for _, f := range githubPayload.HeadCommit.Removed {
			diff += fmt.Sprintf("- %s\n", f)
		}
		diff += "```\n"
	}
	if len(githubPayload.HeadCommit.Modified) > 0 {
		diff += "### 💱 Fichiers modifiés :\n"
		diff += "```diff\n"
		for _, f := range githubPayload.HeadCommit.Modified {
			diff += fmt.Sprintf("~ %s\n", f)
		}
		diff += "```\n"
	}

	lockEmoji := ""
	if githubPayload.Repository.Private {
		lockEmoji = " 🔒"
	}

	description := fmt.Sprintf(
		"# 📨 Message sur le commit :\n```%s```\n%s\n### 👤 Auteur du push\n> ### %s\n### 🔗 Lien vers le commit\n> ### [Clique ici](%s)",
		githubPayload.HeadCommit.Message,
		strings.TrimSpace(diff),
		githubPayload.Pusher.Name,
		githubPayload.HeadCommit.URL,
	)

	embed := DiscordEmbed{
		Description: description,
		Color:       0x00b0f4,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	embed.Footer.Text = "ℹ️ Webhook envoyé automatiquement par ToolCenter"
	embed.Thumbnail.URL = githubPayload.Sender.AvatarURL

	payload := DiscordWebhookPayload{
		Username:  "ToolCenter Notifier",
		AvatarURL: "https://tool-center.fr/assets/tc_logo.webp",
		Embeds:    []DiscordEmbed{embed},
		Content:   fmt.Sprintf("## 🛠️ Nouveau commit sur `%s`%s !", githubPayload.Repository.FullName, lockEmoji),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Println("❌ Erreur marshalling payload Discord:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	discordURL := cfg.DiscordWebhookURL
	if discordURL == "" {
		c.Status(http.StatusNoContent)
		return
	}
	resp, err := http.Post(discordURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil || resp.StatusCode >= 300 {
		log.Println("❌ Erreur envoi vers Discord:", err, "Status:", resp.Status)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send webhook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "message envoyé à Discord avec succès"})
}
