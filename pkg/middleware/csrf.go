package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"
	"sync"
	"time"
)

// CSRFConfig holds CSRF protection configuration
type CSRFConfig struct {
	// Token length in bytes
	TokenLength int
	// Cookie name for CSRF token
	CookieName string
	// Header name for CSRF token
	HeaderName string
	// Form field name for CSRF token
	FieldName string
	// Enable/disable CSRF protection
	Enabled bool
}

// CSRFMiddleware provides CSRF protection
type CSRFMiddleware struct {
	config CSRFConfig
	tokens map[string]time.Time
	mu     sync.RWMutex
}

// NewCSRFMiddleware creates a new CSRF protection middleware
func NewCSRFMiddleware(config CSRFConfig) *CSRFMiddleware {
	if config.TokenLength == 0 {
		config.TokenLength = 32
	}
	if config.CookieName == "" {
		config.CookieName = "csrf_token"
	}
	if config.HeaderName == "" {
		config.HeaderName = "X-CSRF-Token"
	}
	if config.FieldName == "" {
		config.FieldName = "csrf_token"
	}

	return &CSRFMiddleware{
		config: config,
		tokens: make(map[string]time.Time),
	}
}

// Middleware returns the CSRF middleware handler
func (c *CSRFMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF protection if disabled
		if !c.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Skip CSRF for safe methods and static assets
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" ||
			strings.HasPrefix(r.URL.Path, "/css/") ||
			strings.HasPrefix(r.URL.Path, "/js/") ||
			strings.HasPrefix(r.URL.Path, "/favicon.ico") {
			// Set CSRF token for GET requests
			c.setCSRFToken(w, r)
			next.ServeHTTP(w, r)
			return
		}

		// Verify CSRF token for state-changing methods
		if !c.verifyCSRFToken(r) {
			http.Error(w, "CSRF token validation failed", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// setCSRFToken sets a CSRF token cookie
func (c *CSRFMiddleware) setCSRFToken(w http.ResponseWriter, r *http.Request) {
	// Check if token already exists
	if cookie, err := r.Cookie(c.config.CookieName); err == nil && cookie.Value != "" {
		return
	}

	// Generate new token
	token := c.generateCSRFToken()

	// Store token with expiration
	c.mu.Lock()
	c.tokens[token] = time.Now().Add(24 * time.Hour)
	// Clean up expired tokens
	now := time.Now()
	for t, exp := range c.tokens {
		if exp.Before(now) {
			delete(c.tokens, t)
		}
	}
	c.mu.Unlock()

	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     c.config.CookieName,
		Value:    token,
		HttpOnly: false, // Must be readable by JavaScript
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   86400, // 24 hours
	})
}

// verifyCSRFToken verifies the CSRF token from the request
func (c *CSRFMiddleware) verifyCSRFToken(r *http.Request) bool {
	// Get token from cookie
	cookie, err := r.Cookie(c.config.CookieName)
	if err != nil || cookie.Value == "" {
		return false
	}

	// Check if token exists and is not expired
	c.mu.RLock()
	expTime, exists := c.tokens[cookie.Value]
	c.mu.RUnlock()

	if !exists || expTime.Before(time.Now()) {
		return false
	}

	// Get submitted token from header or form
	var submittedToken string

	// Check header first
	submittedToken = r.Header.Get(c.config.HeaderName)

	// If not in header, check form data
	if submittedToken == "" {
		submittedToken = r.FormValue(c.config.FieldName)
	}

	// Constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(submittedToken)) == 1
}

// generateCSRFToken generates a new CSRF token
func (c *CSRFMiddleware) generateCSRFToken() string {
	b := make([]byte, c.config.TokenLength)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// GetCSRFToken retrieves the CSRF token from the request
func (c *CSRFMiddleware) GetCSRFToken(r *http.Request) string {
	if cookie, err := r.Cookie(c.config.CookieName); err == nil {
		return cookie.Value
	}
	return ""
}
