package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gupta/leetcode-judge/internal/common"
	"golang.org/x/time/rate"
)
type RateLimiterConfig struct {
	RequestsPerSecond float64 // How many tokens refill per second
	BurstSize         int     // Max tokens in the bucket (handles spikes)
}

type ipRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rps      rate.Limit
	burst    int
}

func (rl *ipRateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[ip]
	rl.mu.RUnlock()

	if exists {
		return limiter
	}

	rl.mu.Lock()
	limiter = rate.NewLimiter(rl.rps, rl.burst)
	rl.limiters[ip] = limiter
	rl.mu.Unlock()

	return limiter
}

func RateLimiter(cfg RateLimiterConfig) gin.HandlerFunc {
	limiter := &ipRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rate.Limit(cfg.RequestsPerSecond),
		burst:    cfg.BurstSize,
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		l := limiter.getLimiter(ip)

		if !l.Allow() {
			common.Logger.Warnf("rate limit exceeded for IP: %s", ip)
			common.Error(c, http.StatusTooManyRequests, "rate limit exceeded", "RATE_LIMITED")
			c.Abort()
			return
		}

		c.Next()
	}
}
