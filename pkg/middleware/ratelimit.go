package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	// Requests per window
	RequestsPerWindow int
	// Time window duration
	WindowDuration time.Duration
	// Enable/disable rate limiting
	Enabled bool
}

// RateLimiter tracks request rates
type RateLimiter struct {
	config   RateLimitConfig
	visitors map[string]*visitor
	mu       sync.RWMutex
}

// visitor tracks requests for a single IP
type visitor struct {
	requests  int
	firstSeen time.Time
	mu        sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	if config.RequestsPerWindow == 0 {
		config.RequestsPerWindow = 100
	}
	if config.WindowDuration == 0 {
		config.WindowDuration = time.Minute
	}

	rl := &RateLimiter{
		config:   config,
		visitors: make(map[string]*visitor),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Middleware returns the rate limiting middleware handler
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip rate limiting if disabled
		if !rl.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Get client IP
		ip := getClientIP(r)

		// Check rate limit
		if !rl.allow(ip) {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(rl.config.WindowDuration.Seconds())))
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// allow checks if a request from the IP is allowed
func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{
			requests:  0,
			firstSeen: time.Now(),
		}
		rl.visitors[ip] = v
	}
	rl.mu.Unlock()

	v.mu.Lock()
	defer v.mu.Unlock()

	// Reset window if expired
	if time.Since(v.firstSeen) > rl.config.WindowDuration {
		v.requests = 0
		v.firstSeen = time.Now()
	}

	// Check if limit exceeded
	if v.requests >= rl.config.RequestsPerWindow {
		return false
	}

	// Increment counter
	v.requests++
	return true
}

// cleanup removes old visitors periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, v := range rl.visitors {
			v.mu.Lock()
			if now.Sub(v.firstSeen) > 2*rl.config.WindowDuration {
				delete(rl.visitors, ip)
			}
			v.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP if multiple are present
		if idx := len(xff) - 1; idx >= 0 {
			for i := idx; i >= 0; i-- {
				if xff[i] == ',' {
					return xff[i+1:]
				}
			}
			return xff
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}