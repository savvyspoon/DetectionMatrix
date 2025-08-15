package api

import (
	"context"
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
					http.Error(w, "Request timeout", http.StatusRequestTimeout)
				} else {
					http.Error(w, "Request cancelled", http.StatusServiceUnavailable)
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

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of the specified length
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}