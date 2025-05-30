package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
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
}

type DiscordWebhookPayload struct {
	Username  string         `json:"username"`
	AvatarURL string         `json:"avatar_url"`
	Embeds    []DiscordEmbed `json:"embeds"`
	Content   string         `json:"content,omitempty"`
}

func verifyGitHubSignature(secret string, body []byte, sigHeader string) bool {
	if !strings.HasPrefix(sigHeader, "sha256=") || secret == "" {
		return false
	}
	signature := sigHeader[len("sha256="):]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func DiscordWebhookNewsHandler(c *gin.Context) {
	cfg := config.Get()

	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to read body"})
		return
	}

	if !verifyGitHubSignature(cfg.WebhookSecret, rawBody, c.GetHeader("X-Hub-Signature-256")) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))

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
	if err := json.Unmarshal(rawBody, &githubPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	var diff strings.Builder
	if len(githubPayload.HeadCommit.Added) > 0 {
		diff.WriteString("### ✅ Fichiers ajoutés :\n```diff\n")
		for _, f := range githubPayload.HeadCommit.Added {
			diff.WriteString("+ " + f + "\n")
		}
		diff.WriteString("```\n")
	}
	if len(githubPayload.HeadCommit.Removed) > 0 {
		diff.WriteString("### 🚫 Fichiers retirés :\n```diff\n")
		for _, f := range githubPayload.HeadCommit.Removed {
			diff.WriteString("- " + f + "\n")
		}
		diff.WriteString("```\n")
	}
	if len(githubPayload.HeadCommit.Modified) > 0 {
		diff.WriteString("### 💱 Fichiers modifiés :\n```diff\n")
		for _, f := range githubPayload.HeadCommit.Modified {
			diff.WriteString("~ " + f + "\n")
		}
		diff.WriteString("```\n")
	}

	lockEmoji := ""
	if githubPayload.Repository.Private {
		lockEmoji = " 🔒"
	}
	description := fmt.Sprintf(
		"# 📨 Message sur le commit :\n```%s```\n%s\n### 👤 Auteur du push\n> ### %s\n### 🔗 Lien vers le commit\n> ### [Clique ici](%s)",
		githubPayload.HeadCommit.Message,
		strings.TrimSpace(diff.String()),
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

	/* 7. Send to Discord */
	jsonPayload, _ := json.Marshal(payload)
	if cfg.DiscordWebhookURL != "" {
		r, err := http.Post(cfg.DiscordWebhookURL, "application/json", bytes.NewBuffer(jsonPayload))
		if err != nil || r.StatusCode >= 300 {
			log.Println("❌ Discord webhook error:", err, r.Status)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send webhook"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "message envoyé à Discord avec succès"})
}
