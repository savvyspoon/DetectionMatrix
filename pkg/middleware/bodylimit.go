package middleware

import (
	"net/http"
)

// BodyLimitConfig holds body size limit configuration
type BodyLimitConfig struct {
	// Maximum request body size in bytes
	MaxBodySize int64
	// Enable/disable body size limiting
	Enabled bool
}

// BodyLimitMiddleware limits request body size
type BodyLimitMiddleware struct {
	config BodyLimitConfig
}

// NewBodyLimitMiddleware creates a new body limit middleware
func NewBodyLimitMiddleware(config BodyLimitConfig) *BodyLimitMiddleware {
	if config.MaxBodySize == 0 {
		config.MaxBodySize = 10 * 1024 * 1024 // 10MB default
	}
	return &BodyLimitMiddleware{
		config: config,
	}
}

// Middleware returns the body limit middleware handler
func (b *BodyLimitMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip body limiting if disabled
		if !b.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Limit request body size
		r.Body = http.MaxBytesReader(w, r.Body, b.config.MaxBodySize)

		next.ServeHTTP(w, r)
	})
}