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
		SignInSecret   string `json:"signin_secret"`
		SignUpSecret   string `json:"signup_secret"`
		TimeoutSeconds int    `json:"timeout_seconds"`
	} `json:"turnstile"`
	Cleanup struct {
		CheckInterval int `json:"check_interval"`
		GracePeriod   int `json:"grace_period"`
	} `json:"cleanup"`
	Storage struct {
		AvatarDir     string `json:"avatar_dir"`
		ToolsImageDir string `json:"tools_image_dir"`
	} `json:"storage"`
	Moderation struct {
		MaxBanDays int  `json:"max_ban_days"`
		AutoUnban  bool `json:"auto_unban"`
	} `json:"moderation"`
	Cooldowns struct {
		EmailChangeDays    int `json:"email_change_days"`
		UsernameChangeDays int `json:"username_change_days"`
		ToolPostHours      int `json:"tool_post_hours"`
		AvatarChangeHours  int `json:"avatar_change_hours"`
	} `json:"cooldowns"`
	TwoFactor struct {
		Issuer string `json:"issuer"`
	} `json:"two_factor"`
	PasswordReset struct {
		TokenExpiryMinutes int `json:"token_expiry_minutes"`
	} `json:"password_reset"`
        RateLimit struct {
                Requests      int `json:"requests"`
                WindowSeconds int `json:"window_seconds"`
        } `json:"rate_limit"`
        AntiSpam struct {
                Enabled            bool    `json:"enabled"`
                RequestThreshold   int     `json:"request_threshold"`
                IntervalSeconds    int     `json:"interval_seconds"`
                InitialBlockSeconds int    `json:"initial_block_seconds"`
                MaxStrike          int     `json:"max_strike"`
                StatusDecrease     int     `json:"status_decrease"`
                BanHours           int     `json:"ban_hours"`
                ProxyMultiplier    int     `json:"proxy_multiplier"`
        } `json:"anti_spam"`
	StatusBanner struct {
		ErrorRateThreshold  float64 `json:"error_rate_threshold"`
		WindowMinutes       int     `json:"window_minutes"`
		Link                string  `json:"link"`
		Message             string  `json:"message"`
		ActivatedForTesting bool    `json:"activated_for_testing"`
	} `json:"status_banner"`
	PrivateNews struct {
		Password   string `json:"password"`
		TokenHours int    `json:"token_hours"`
	} `json:"private_news"`
	UserSearchLimit      int `json:"user_search_limit"`
	UserPublicToolsLimit int `json:"user_public_tools_limit"`
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
	if cfg.Moderation.MaxBanDays == 0 {
		cfg.Moderation.MaxBanDays = 30
	}
	if !cfg.Moderation.AutoUnban {
		cfg.Moderation.AutoUnban = true
	}
	if cfg.TwoFactor.Issuer == "" {
		cfg.TwoFactor.Issuer = "ToolCenter"
	}
	if cfg.PasswordReset.TokenExpiryMinutes == 0 {
		cfg.PasswordReset.TokenExpiryMinutes = 15
	}
	if cfg.RateLimit.Requests == 0 {
		cfg.RateLimit.Requests = 60
	}
	if cfg.RateLimit.WindowSeconds == 0 {
		cfg.RateLimit.WindowSeconds = 60
	}
	if cfg.StatusBanner.WindowMinutes == 0 {
		cfg.StatusBanner.WindowMinutes = 10
	}
	if cfg.StatusBanner.ErrorRateThreshold == 0 {
		cfg.StatusBanner.ErrorRateThreshold = 0.3
	}
	if cfg.StatusBanner.Message == "" {
		cfg.StatusBanner.Message = "Des perturbations sont en cours."
	}
	if cfg.Turnstile.TimeoutSeconds == 0 {
		cfg.Turnstile.TimeoutSeconds = 5
	}
	if cfg.PrivateNews.TokenHours == 0 {
		cfg.PrivateNews.TokenHours = 24
	}
	if cfg.UserSearchLimit <= 0 || cfg.UserSearchLimit > 20 {
		cfg.UserSearchLimit = 10
	}
        if cfg.UserPublicToolsLimit <= 0 || cfg.UserPublicToolsLimit > 10 {
                cfg.UserPublicToolsLimit = 3
        }
       if cfg.AntiSpam.RequestThreshold == 0 {
               cfg.AntiSpam.RequestThreshold = 10
       }
       if cfg.AntiSpam.IntervalSeconds == 0 {
               cfg.AntiSpam.IntervalSeconds = 1
       }
       if cfg.AntiSpam.InitialBlockSeconds == 0 {
               cfg.AntiSpam.InitialBlockSeconds = 30
       }
       if cfg.AntiSpam.MaxStrike == 0 {
               cfg.AntiSpam.MaxStrike = 3
       }
       if cfg.AntiSpam.StatusDecrease == 0 {
               cfg.AntiSpam.StatusDecrease = 2
       }
       if cfg.AntiSpam.BanHours == 0 {
               cfg.AntiSpam.BanHours = 24
       }
       if cfg.AntiSpam.ProxyMultiplier == 0 {
               cfg.AntiSpam.ProxyMultiplier = 2
       }
       if !cfg.AntiSpam.Enabled {
               cfg.AntiSpam.Enabled = true
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
