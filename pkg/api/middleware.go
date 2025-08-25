package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

// ContextKey type for context values
type ContextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey ContextKey = "requestID"
	// UserKey is the context key for authenticated user
	UserKey ContextKey = "user"
)

// WithTimeout adds a timeout to the request context
func WithTimeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Handle context cancellation
			done := make(chan struct{})
			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				// Request completed normally
			case <-ctx.Done():
				// Context cancelled or timed out
				if ctx.Err() == context.DeadlineExceeded {
					Error(w, r, http.StatusRequestTimeout, "Request timeout")
				} else {
					Error(w, r, http.StatusServiceUnavailable, "Request cancelled")
				}
			}
		})
	}
}

// RequestIDMiddleware adds a unique request ID to the context
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// generateRequestID generates a unique request ID using cryptographically secure random bytes
func generateRequestID() string {
	// Generate 16 random bytes (128 bits) for uniqueness
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		return fmt.Sprintf("%d-%x", time.Now().UnixNano(), time.Now().UnixNano())
	}
	
	// Format as hex string with timestamp prefix for sortability
	// Format: YYYYMMDD-HHMMSS-<random-hex>
	return fmt.Sprintf("%s-%s", 
		time.Now().Format("20060102-150405"),
		hex.EncodeToString(b))
}
