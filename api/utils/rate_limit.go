package utils

import (
        "sync"
        "time"
        "strings"

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
        ipBlocks = make(map[string]time.Time)
        ipAttempts = make(map[string]int)
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
                for ip, until := range ipBlocks {
                        if time.Now().After(until) {
                                delete(ipBlocks, ip)
                                delete(ipAttempts, ip)
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
                rlMu.Lock()
                if until, ok := ipBlocks[ip]; ok && time.Now().Before(until) {
                        rlMu.Unlock()
                        c.AbortWithStatusJSON(429, gin.H{"success": false, "message": "Trop de requêtes, veuillez patienter."})
                        return
                }
                rlMu.Unlock()
                limiter := getLimiter(ip)
                if !limiter.Allow() {
                        rlMu.Lock()
                        ipAttempts[ip]++
                        attempts := ipAttempts[ip]
                        cfg := config.Get()
                        base := cfg.SpamProtection.BlockInitialSeconds
                        if base <= 0 {
                                base = 60
                        }
                        mult := cfg.SpamProtection.BlockMultiplier
                        if mult <= 1 {
                                mult = 2
                        }
                        dur := time.Duration(base) * time.Second
                        for i := 1; i < attempts; i++ {
                                dur = time.Duration(float64(dur) * mult)
                        }
                        ipBlocks[ip] = time.Now().Add(dur)
                        rlMu.Unlock()

                        if strings.Contains(c.GetHeader("X-Forwarded-For"), ",") {
                                ipAttempts[ip]++
                        }

                        if uid, _, _, _, err := Check(c, CheckOpts{RequireToken: true}); err == nil {
                                thr := cfg.SpamProtection.SanctionThreshold
                                if thr == 0 {
                                        thr = 3
                                }
                                if attempts >= thr {
                                        drop := cfg.SpamProtection.StatusDrop
                                        if drop <= 0 {
                                                drop = 2
                                        }
                                        ApplyStatusDrop(uid, drop)
                                        db, err := config.OpenDB()
                                        if err == nil {
                                                var email string
                                                _ = db.QueryRow(`SELECT email FROM users WHERE user_id=?`, uid).Scan(&email)
                                                db.Close()
                                                if email != "" {
                                                        _ = SendEmail(email, "Sanction ToolCenter", "Vous avez été sanctionné pour spammer l'API.")
                                                }
                                        }
                                        ipAttempts[ip] = 0
                                }
                        }

                        c.AbortWithStatusJSON(429, gin.H{"success": false, "message": "Trop de requêtes, veuillez patienter."})
                        return
                }
                c.Next()
        }
}
