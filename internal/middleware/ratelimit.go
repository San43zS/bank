package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type tokenBucket struct {
	tokens   float64
	last     time.Time
	lastSeen time.Time
}

type ipRateLimiter struct {
	rps   float64
	burst float64

	mu      sync.Mutex
	buckets map[string]*tokenBucket
}

func newIPRateLimiter(limitPerSecond int, burst int) *ipRateLimiter {
	if limitPerSecond <= 0 {
		limitPerSecond = 1
	}
	if burst <= 0 {
		burst = 1
	}
	rl := &ipRateLimiter{
		rps:     float64(limitPerSecond),
		burst:   float64(burst),
		buckets: make(map[string]*tokenBucket),
	}
	go rl.cleanupLoop(10*time.Minute, 30*time.Minute)
	return rl
}

func (r *ipRateLimiter) cleanupLoop(every time.Duration, maxAge time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	for now := range t.C {
		cutoff := now.Add(-maxAge)
		r.mu.Lock()
		for k, b := range r.buckets {
			if b.lastSeen.Before(cutoff) {
				delete(r.buckets, k)
			}
		}
		r.mu.Unlock()
	}
}

func (r *ipRateLimiter) allow(key string) bool {
	now := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	b, ok := r.buckets[key]
	if !ok {
		b = &tokenBucket{
			tokens:   r.burst,
			last:     now,
			lastSeen: now,
		}
		r.buckets[key] = b
	}

	elapsed := now.Sub(b.last).Seconds()
	if elapsed < 0 {
		elapsed = 0
	}
	b.tokens += elapsed * r.rps
	if b.tokens > r.burst {
		b.tokens = r.burst
	}
	b.last = now
	b.lastSeen = now

	if b.tokens < 1 {
		return false
	}
	b.tokens -= 1
	return true
}

func RateLimitMiddleware(limitPerSecond int, burst int) gin.HandlerFunc {
	if limitPerSecond <= 0 || burst <= 0 {
		return func(c *gin.Context) { c.Next() }
	}

	rl := newIPRateLimiter(limitPerSecond, burst)
	return func(c *gin.Context) {
		key := c.ClientIP()
		if key == "" {
			key = "unknown"
		}
		if !rl.allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate_limited"})
			c.Abort()
			return
		}
		c.Next()
	}
}
