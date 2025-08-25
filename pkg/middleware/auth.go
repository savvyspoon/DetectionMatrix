package middleware

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"
	"sync"
	"time"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	// Basic auth credentials (for simple protection)
	Username string
	Password string
	// Session duration
	SessionDuration time.Duration
	// Enable/disable authentication
	Enabled bool
}

// Session represents a user session
type Session struct {
	Token     string
	Username  string
	ExpiresAt time.Time
}

// AuthMiddleware provides basic authentication
type AuthMiddleware struct {
	config   AuthConfig
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(config AuthConfig) *AuthMiddleware {
	if config.SessionDuration == 0 {
		config.SessionDuration = 24 * time.Hour
	}
	return &AuthMiddleware{
		config:   config,
		sessions: make(map[string]*Session),
	}
}

// Middleware returns the authentication middleware handler
func (a *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication if disabled
		if !a.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Allow public assets
		if strings.HasPrefix(r.URL.Path, "/css/") ||
			strings.HasPrefix(r.URL.Path, "/js/") ||
			strings.HasPrefix(r.URL.Path, "/favicon.ico") {
			next.ServeHTTP(w, r)
			return
		}

		// Check for session cookie
		cookie, err := r.Cookie("session")
		if err == nil {
			a.mu.RLock()
			session, exists := a.sessions[cookie.Value]
			a.mu.RUnlock()

			if exists && session.ExpiresAt.After(time.Now()) {
				// Valid session, add user to context
				ctx := context.WithValue(r.Context(), "user", session.Username)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// Check for Basic Auth
		username, password, ok := r.BasicAuth()
		if !ok || !a.validateCredentials(username, password) {
			w.Header().Set("WWW-Authenticate", `Basic realm="RiskMatrix"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Create session
		session := a.createSession(username)
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    session.Token,
			Expires:  session.ExpiresAt,
			HttpOnly: true,
			Secure:   r.TLS != nil,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
		})

		// Add user to context
		ctx := context.WithValue(r.Context(), "user", username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateCredentials checks if the provided credentials are valid
func (a *AuthMiddleware) validateCredentials(username, password string) bool {
	// Use constant-time comparison to prevent timing attacks
	userMatch := subtle.ConstantTimeCompare([]byte(username), []byte(a.config.Username)) == 1
	passMatch := subtle.ConstantTimeCompare([]byte(password), []byte(a.config.Password)) == 1
	return userMatch && passMatch
}

// createSession creates a new session for the user
func (a *AuthMiddleware) createSession(username string) *Session {
	token := generateToken()
	session := &Session{
		Token:     token,
		Username:  username,
		ExpiresAt: time.Now().Add(a.config.SessionDuration),
	}

	a.mu.Lock()
	a.sessions[token] = session
	// Clean up expired sessions
	for t, s := range a.sessions {
		if s.ExpiresAt.Before(time.Now()) {
			delete(a.sessions, t)
		}
	}
	a.mu.Unlock()

	return session
}

// generateToken generates a secure random token
func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// GetUser retrieves the authenticated user from the request context
func GetUser(r *http.Request) (string, bool) {
	user, ok := r.Context().Value("user").(string)
	return user, ok
}
