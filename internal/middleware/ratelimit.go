package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ahmadasror/txsurvey/pkg/response"
)

// RateLimit is a simple per-IP fixed-window limiter — sufficient for the public
// runner endpoints at this scale (5-10 creators, single instance). It bounds
// abuse of the anonymous submit/fetch endpoints without an external store.
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	rl := &limiter{hits: make(map[string]*counter), limit: limit, window: window}
	return func(c *gin.Context) {
		ok, retryAfter := rl.allow(c.ClientIP())
		if !ok {
			c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
			response.Error(c, http.StatusTooManyRequests, "RATE_LIMITED", "too many requests, please slow down")
			c.Abort()
			return
		}
		c.Next()
	}
}

type counter struct {
	count int
	reset time.Time
}

type limiter struct {
	mu     sync.Mutex
	hits   map[string]*counter
	limit  int
	window time.Duration
}

func (l *limiter) allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()

	// Opportunistic prune so the map can't grow unbounded.
	if len(l.hits) > 10_000 {
		for k, c := range l.hits {
			if now.After(c.reset) {
				delete(l.hits, k)
			}
		}
	}

	c, ok := l.hits[key]
	if !ok || now.After(c.reset) {
		l.hits[key] = &counter{count: 1, reset: now.Add(l.window)}
		return true, 0
	}
	if c.count >= l.limit {
		return false, time.Until(c.reset)
	}
	c.count++
	return true, 0
}
