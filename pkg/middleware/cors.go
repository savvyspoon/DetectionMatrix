package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSConfig holds configuration for CORS handling
type CORSConfig struct {
	Enabled          bool
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAgeSeconds    int
}

// CORSMiddleware adds CORS headers and handles preflight requests
type CORSMiddleware struct {
	config CORSConfig
}

// NewCORSMiddleware creates a CORS middleware with the given config
func NewCORSMiddleware(config CORSConfig) *CORSMiddleware {
	// Reasonable defaults
	if len(config.AllowedMethods) == 0 {
		config.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}
	if len(config.AllowedHeaders) == 0 {
		config.AllowedHeaders = []string{"Content-Type", "Authorization", "X-CSRF-Token"}
	}
	if config.MaxAgeSeconds <= 0 {
		config.MaxAgeSeconds = 600
	}
	return &CORSMiddleware{config: config}
}

// Middleware wraps the handler adding CORS logic
func (c *CORSMiddleware) Middleware(next http.Handler) http.Handler {
	if !c.config.Enabled {
		return next
	}

	methods := strings.Join(c.config.AllowedMethods, ", ")
	headers := strings.Join(c.config.AllowedHeaders, ", ")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Compute Access-Control-Allow-Origin
		if origin != "" {
			if contains(c.config.AllowedOrigins, "*") {
				// If credentials are allowed, echo origin instead of *
				if c.config.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				} else {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				}
			} else if contains(c.config.AllowedOrigins, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			if c.config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}

		// Always set varying headers for caches
		if origin != "" {
			w.Header().Add("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.Header().Add("Vary", "Access-Control-Request-Method")
			w.Header().Add("Vary", "Access-Control-Request-Headers")
		}

		// For both preflight and actual requests
		w.Header().Set("Access-Control-Allow-Methods", methods)
		w.Header().Set("Access-Control-Allow-Headers", headers)
		w.Header().Set("Access-Control-Max-Age", strconv.Itoa(c.config.MaxAgeSeconds))

		// Handle preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func contains(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

// no-op helper removed; using strconv.Itoa for conversions
