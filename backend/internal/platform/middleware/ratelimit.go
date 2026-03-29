package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"jetistik/internal/platform/response"
)

type visitor struct {
	tokens   float64
	lastSeen time.Time
}

// RateLimit creates middleware that limits requests per IP using a token bucket.
// rate is the number of requests allowed per interval.
func RateLimit(rate int, interval time.Duration) func(http.Handler) http.Handler {
	var mu sync.Mutex
	visitors := make(map[string]*visitor)
	maxTokens := float64(rate)
	refillRate := maxTokens / interval.Seconds()

	// Background cleanup of stale visitors
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > 10*time.Minute {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			// Check X-Forwarded-For for proxied requests
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				ip = forwarded
			}

			mu.Lock()
			v, exists := visitors[ip]
			now := time.Now()

			if !exists {
				v = &visitor{tokens: maxTokens, lastSeen: now}
				visitors[ip] = v
			}

			// Refill tokens based on elapsed time
			elapsed := now.Sub(v.lastSeen).Seconds()
			v.tokens += elapsed * refillRate
			if v.tokens > maxTokens {
				v.tokens = maxTokens
			}
			v.lastSeen = now

			if v.tokens < 1 {
				mu.Unlock()
				w.Header().Set("Retry-After", "60")
				response.Error(w, http.StatusTooManyRequests, "RATE_LIMITED", "too many requests, please try again later")
				return
			}

			v.tokens--
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
