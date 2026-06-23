package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// tokenBucket implements a simple token bucket rate limiter per IP.
type tokenBucket struct {
	tokens   float64
	capacity float64
	rate     float64 // tokens per second
	lastFill time.Time
	mu       sync.Mutex
}

func newBucket(capacity float64, rate float64) *tokenBucket {
	return &tokenBucket{
		tokens:   capacity,
		capacity: capacity,
		rate:     rate,
		lastFill: time.Now(),
	}
}

// take attempts to consume one token. Returns true if allowed.
func (b *tokenBucket) take() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastFill).Seconds()
	b.tokens = min(b.capacity, b.tokens+elapsed*b.rate)
	b.lastFill = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// ipLimiter manages one bucket per client IP.
type ipLimiter struct {
	buckets  sync.Map
	capacity float64
	rate     float64
}

func newIPLimiter(capacity float64, rate float64) *ipLimiter {
	l := &ipLimiter{capacity: capacity, rate: rate}
	// Periodically clean up stale buckets to prevent memory leaks
	go l.cleanup()
	return l
}

func (l *ipLimiter) allow(ip string) bool {
	v, _ := l.buckets.LoadOrStore(ip, newBucket(l.capacity, l.rate))
	return v.(*tokenBucket).take()
}

// cleanup removes buckets that have been full for more than 10 minutes
// (meaning the IP has been idle and its bucket is no longer needed).
func (l *ipLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		l.buckets.Range(func(k, v any) bool {
			b := v.(*tokenBucket)
			b.mu.Lock()
			idle := time.Since(b.lastFill) > 10*time.Minute
			b.mu.Unlock()
			if idle {
				l.buckets.Delete(k)
			}
			return true
		})
	}
}

// RateLimiter returns a Gin middleware that limits requests per IP.
//
// capacity: max burst size (tokens)
// rate:     sustained requests per second
//
// Example: RateLimiter(10, 2) allows bursts of 10 with a sustained rate of 2 req/s.
func RateLimiter(capacity float64, rate float64) gin.HandlerFunc {
	limiter := newIPLimiter(capacity, rate)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests — please slow down",
			})
			return
		}
		c.Next()
	}
}
