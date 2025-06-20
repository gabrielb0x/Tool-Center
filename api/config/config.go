package config

import (
	"encoding/json"
	"os"
	"sync"
)

var (
	mu      sync.RWMutex
	Current Config
)

type Config struct {
	GinMode           string `json:"gin_mode"`
	Port              int    `json:"port"`
	Version           string `json:"version"`
	URLVersion        string `json:"URL_version"`
	URLapi            string `json:"URL_api"`
	URLweb            string `json:"URL_web"`
	CorsAllowedOrigin string `json:"cors_allowed_origin"`
	Debug             bool   `json:"debug"`
	Logs              struct {
		Enabled bool   `json:"enabled"`
		Path    string `json:"path"`
	} `json:"logs"`
	Database struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		DBName   string `json:"dbname"`
	} `json:"database"`
	Email struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		From     string `json:"from"`
	} `json:"email"`
	DiscordWebhookURL string `json:"discord_webhook_url"`
	WebhookSecret     string `json:"webhook_secret"`
	Turnstile         struct {
		SignInSecret string `json:"signin_secret"`
		SignUpSecret string `json:"signup_secret"`
	} `json:"turnstile"`
	Cleanup struct {
		CheckInterval int `json:"check_interval"`
		GracePeriod   int `json:"grace_period"`
	} `json:"cleanup"`
	Storage struct {
		AvatarDir     string `json:"avatar_dir"`
		ToolsImageDir string `json:"tools_image_dir"`
	} `json:"storage"`
	Cooldowns struct {
		EmailChangeDays    int `json:"email_change_days"`
		UsernameChangeDays int `json:"username_change_days"`
		ToolPostHours      int `json:"tool_post_hours"`
		AvatarChangeHours  int `json:"avatar_change_hours"`
	} `json:"cooldowns"`
	PrivateNewsPassword string `json:"private_news_password"`
}

func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}
	mu.Lock()
	Current = cfg
	mu.Unlock()
	return nil
}

func Get() Config {
	mu.RLock()
	defer mu.RUnlock()
	return Current
}
