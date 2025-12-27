package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/anurag-327/neuron/internal/util/response"
	"github.com/gin-gonic/gin"
)

// RateLimiter implements a simple in-memory token bucket rate limiter
type RateLimiter struct {
	mu       sync.RWMutex
	visitors map[string]*visitor
	rate     int           // requests per window
	window   time.Duration // time window
}

type visitor struct {
	tokens     int
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}

	// Cleanup old visitors every 5 minutes
	go rl.cleanupVisitors()

	return rl
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[key]
	now := time.Now()

	if !exists {
		rl.visitors[key] = &visitor{
			tokens:     rl.rate - 1,
			lastRefill: now,
		}
		return true
	}

	// Refill tokens if window has passed
	if now.Sub(v.lastRefill) >= rl.window {
		v.tokens = rl.rate
		v.lastRefill = now
	}

	if v.tokens > 0 {
		v.tokens--
		return true
	}

	return false
}

// cleanupVisitors removes old visitors to prevent memory leak
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, v := range rl.visitors {
			if now.Sub(v.lastRefill) > rl.window*2 {
				delete(rl.visitors, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Global rate limiters for different endpoints
var (
	// General API rate limiter: 60 requests per minute
	generalLimiter = NewRateLimiter(60, time.Minute)

	// Auth endpoints: 10 requests per minute (stricter)
	authLimiter = NewRateLimiter(10, time.Minute)

	// Code submission: 30 requests per minute
	submissionLimiter = NewRateLimiter(30, time.Minute)
)

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use user ID if authenticated, otherwise use IP
		key := c.ClientIP()

		if userID, exists := c.Get("user_id"); exists {
			key = userID.(string)
		}

		if !limiter.Allow(key) {
			response.Error(c, http.StatusTooManyRequests, "rate limit exceeded, please try again later")
			c.Abort()
			return
		}

		c.Next()
	}
}

// GeneralRateLimit applies general rate limiting
func GeneralRateLimit() gin.HandlerFunc {
	return RateLimitMiddleware(generalLimiter)
}

// AuthRateLimit applies stricter rate limiting for auth endpoints
func AuthRateLimit() gin.HandlerFunc {
	return RateLimitMiddleware(authLimiter)
}

// SubmissionRateLimit applies rate limiting for code submissions
func SubmissionRateLimit() gin.HandlerFunc {
	return RateLimitMiddleware(submissionLimiter)
}
