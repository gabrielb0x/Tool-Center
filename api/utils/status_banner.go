package utils

import (
	"sync"
	"time"

	"toolcenter/config"

	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	mu       sync.Mutex
	reqTimes []time.Time
	errTimes []time.Time
)

func recordRequest(isError bool) {
	mu.Lock()
	defer mu.Unlock()
	now := time.Now()
	reqTimes = append(reqTimes, now)
	if isError {
		errTimes = append(errTimes, now)
	}
	cleanup()
}

func cleanup() {
	cfg := config.Get()
	window := cfg.StatusBanner.WindowMinutes
	if window == 0 {
		window = 10
	}
	cutoff := time.Now().Add(-time.Duration(window) * time.Minute)
	for len(reqTimes) > 0 && reqTimes[0].Before(cutoff) {
		reqTimes = reqTimes[1:]
	}
	for len(errTimes) > 0 && errTimes[0].Before(cutoff) {
		errTimes = errTimes[1:]
	}
}

type Status struct {
	ErrorRate    float64 `json:"error_rate"`
	RequestCount int     `json:"request_count"`
	ErrorCount   int     `json:"error_count"`
}

func MonitorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		recordRequest(c.Writer.Status() >= 500)
	}
}

func getStatus() Status {
	mu.Lock()
	defer mu.Unlock()
	cleanup()
	total := len(reqTimes)
	errors := len(errTimes)
	rate := 0.0
	if total > 0 {
		rate = float64(errors) / float64(total)
	}
	return Status{
		ErrorRate:    rate,
		RequestCount: total,
		ErrorCount:   errors,
	}
}

func StatusHandler(c *gin.Context) {
	st := getStatus()
	cfg := config.Get()
	show := st.ErrorRate >= cfg.StatusBanner.ErrorRateThreshold && st.RequestCount > 0

	if cfg.StatusBanner.ActivatedForTesting {
		msg := fmt.Sprintf("%s (%.0f%% d'erreurs)", cfg.StatusBanner.Message, st.ErrorRate*100)
		c.JSON(http.StatusOK, gin.H{
			"show_banner": true,
			"message":     msg,
			"link":        cfg.StatusBanner.Link,
		})
		return
	}

	if show {
		msg := fmt.Sprintf("%s (%.0f%% d'erreurs)", cfg.StatusBanner.Message, st.ErrorRate*100)
		c.JSON(http.StatusOK, gin.H{
			"show_banner": true,
			"message":     msg,
			"link":        cfg.StatusBanner.Link,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"show_banner": false})
}
