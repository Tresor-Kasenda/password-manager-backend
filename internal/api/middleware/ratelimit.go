package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
}

type visitor struct {
	lastSeen time.Time
	count    int
}

var limiter = &rateLimiter{
	visitors: make(map[string]*visitor),
}

func RateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			limiter.cleanup()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		limiter.mu.Lock()
		v, exists := limiter.visitors[ip]
		if !exists {
			limiter.visitors[ip] = &visitor{
				lastSeen: time.Now(),
				count:    1,
			}
			limiter.mu.Unlock()
			c.Next()
			return
		}

		if time.Since(v.lastSeen) > time.Minute {
			v.count = 1
			v.lastSeen = time.Now()
			limiter.mu.Unlock()
			c.Next()
			return
		}

		if v.count >= requestsPerMinute {
			limiter.mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		v.count++
		limiter.mu.Unlock()

		c.Next()
	}
}

func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for ip, v := range rl.visitors {
		if time.Since(v.lastSeen) > 10*time.Minute {
			delete(rl.visitors, ip)
		}
	}
}
