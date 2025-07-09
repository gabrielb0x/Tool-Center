package utils

import (
	"sync"
	"time"

	"toolcenter/config"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	rlMu     sync.Mutex
	ipLimits = make(map[string]*limiterEntry)
)

func getLimiter(ip string) *rate.Limiter {
	rlMu.Lock()
	defer rlMu.Unlock()
	entry, ok := ipLimits[ip]
	if !ok {
		cfg := config.Get()
		if cfg.RateLimit.Requests <= 0 {
			cfg.RateLimit.Requests = 60
		}
		if cfg.RateLimit.WindowSeconds <= 0 {
			cfg.RateLimit.WindowSeconds = 60
		}
		r := rate.Every(time.Duration(cfg.RateLimit.WindowSeconds) * time.Second / time.Duration(cfg.RateLimit.Requests))
		entry = &limiterEntry{
			limiter:  rate.NewLimiter(r, cfg.RateLimit.Requests),
			lastSeen: time.Now(),
		}
		ipLimits[ip] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter
}

func cleanupLimiters() {
	for {
		time.Sleep(5 * time.Minute)
		rlMu.Lock()
		for ip, entry := range ipLimits {
			if time.Since(entry.lastSeen) > 10*time.Minute {
				delete(ipLimits, ip)
			}
		}
		rlMu.Unlock()
	}
}

func init() {
	go cleanupLimiters()
}

// RateLimitMiddleware restricts the number of requests per IP
// based on the RateLimit settings in config.json.
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if ip == "" {
			ip = "unknown"
		}
		limiter := getLimiter(ip)
		if !limiter.Allow() {
			c.AbortWithStatusJSON(429, gin.H{"success": false, "message": "Trop de requêtes, veuillez patienter."})
			return
		}
		c.Next()
	}
}
