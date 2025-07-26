package types

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
)

// LoginAttempt tracks login attempts for rate limiting
type LoginAttempt struct {
	Count        int       `json:"count"`
	LastAttempt  time.Time `json:"last_attempt"`
	BlockedUntil time.Time `json:"blocked_until"`
}

// LoginRateLimiter manages login rate limiting
type LoginRateLimiter struct {
	Attempts map[string]*LoginAttempt
	Mu       sync.RWMutex

	// Configuration
	MaxAttempts    int           // Max attempts before blocking
	BlockDuration  time.Duration // How long to block after max attempts
	WindowDuration time.Duration // Time window to count attempts
}

// Session represents a WhatsApp session
type Session struct {
	ID          string             `json:"id"`
	Phone       string             `json:"phone"`
	ActualPhone string             `json:"actual_phone,omitempty"`
	Name        string             `json:"name"`
	Position    int                `json:"position"`
	Connected   bool               `json:"connected"`
	LoggedIn    bool               `json:"logged_in"`
	Client      *whatsmeow.Client  `json:"-"`
	WebhookURL  string             `json:"webhook_url,omitempty"`
	CreatedAt   int64              `json:"created_at"`
	Connecting  bool               `json:"connecting"`
}

// SessionMetadata holds session metadata for database storage
type SessionMetadata struct {
	ID          string `json:"id"`
	Phone       string `json:"phone"`
	ActualPhone string `json:"actual_phone,omitempty"`
	Name        string `json:"name"`
	Position    int    `json:"position"`
	WebhookURL  string `json:"webhook_url,omitempty"`
	CreatedAt   int64  `json:"created_at"`
}

// User represents a user in the system
type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	Password     string `json:"-"` // Don't include in JSON responses
	Role         string `json:"role"`
	SessionLimit int    `json:"session_limit"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    *int64 `json:"updated_at,omitempty"`
}

// Config holds database and application configuration
type Config struct {
	DatabasePath  string // Path to SQLite database file
	JWTSecret     string
	EnableLogging bool   // Whether to enable logging
	LogLevel      string // Log level: "debug", "info", "warn", "error"
}

// Claims represents JWT claims
type Claims struct {
	Username string `json:"username"`
	UserID   int    `json:"user_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// SessionManager manages multiple WhatsApp sessions
type SessionManager struct {
	Sessions   map[string]*Session
	Store      *sqlstore.Container
	MetadataDB *sql.DB
	Mu         sync.RWMutex
}

// Logger provides configurable logging
type Logger struct {
	Enabled bool
	Level   string
	Logger  *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger(enabled bool, level string) *Logger {
	return &Logger{
		Enabled: enabled,
		Level:   level,
		Logger:  log.New(os.Stdout, "", log.LstdFlags),
	}
}

// ShouldLog checks if a message should be logged based on level
func (l *Logger) ShouldLog(level string) bool {
	if !l.Enabled {
		return false
	}

	levels := map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
	}

	currentLevel, ok := levels[l.Level]
	if !ok {
		currentLevel = 1 // default to info
	}

	msgLevel, ok := levels[level]
	if !ok {
		msgLevel = 1 // default to info
	}

	return msgLevel >= currentLevel
}

// Debug logs debug messages
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.ShouldLog("debug") {
		l.Logger.Printf("[DEBUG] "+format, v...)
	}
}

// Info logs info messages
func (l *Logger) Info(format string, v ...interface{}) {
	if l.ShouldLog("info") {
		l.Logger.Printf("[INFO] "+format, v...)
	}
}

// Warn logs warning messages
func (l *Logger) Warn(format string, v ...interface{}) {
	if l.ShouldLog("warn") {
		l.Logger.Printf("[WARN] "+format, v...)
	}
}

// Error logs error messages
func (l *Logger) Error(format string, v ...interface{}) {
	if l.ShouldLog("error") {
		l.Logger.Printf("[ERROR] "+format, v...)
	}
}

// Printf logs formatted messages (for compatibility)
func (l *Logger) Printf(format string, v ...interface{}) {
	l.Info(format, v...)
}

// Println logs messages (for compatibility)
func (l *Logger) Println(v ...interface{}) {
	if l.ShouldLog("info") {
		l.Logger.Println("[INFO]", fmt.Sprint(v...))
	}
}

// Print logs messages without newline (for compatibility)
func (l *Logger) Print(v ...interface{}) {
	if l.ShouldLog("info") {
		l.Logger.Print("[INFO] ", fmt.Sprint(v...))
	}
}

// Fatal logs fatal messages and exits
func (l *Logger) Fatal(v ...interface{}) {
	l.Logger.Fatal(v...)
}

// LoginRateLimiter methods

// IsBlocked checks if an IP is currently blocked
func (l *LoginRateLimiter) IsBlocked(ip string) bool {
	l.Mu.RLock()
	defer l.Mu.RUnlock()

	attempt, exists := l.Attempts[ip]
	if !exists {
		return false
	}

	return time.Now().Before(attempt.BlockedUntil)
}

// RecordAttempt records a login attempt (successful or failed)
func (l *LoginRateLimiter) RecordAttempt(ip string, success bool, logger *Logger) {
	l.Mu.Lock()
	defer l.Mu.Unlock()

	now := time.Now()
	attempt, exists := l.Attempts[ip]

	if !exists {
		attempt = &LoginAttempt{}
		l.Attempts[ip] = attempt
	}

	// Reset if outside window
	if now.Sub(attempt.LastAttempt) > l.WindowDuration {
		attempt.Count = 0
	}

	if success {
		// Reset on successful login
		attempt.Count = 0
		attempt.BlockedUntil = time.Time{}
	} else {
		// Increment failed attempts
		attempt.Count++
		attempt.LastAttempt = now

		// Block if exceeded max attempts
		if attempt.Count >= l.MaxAttempts {
			attempt.BlockedUntil = now.Add(l.BlockDuration)
			if logger != nil {
				logger.Warn("IP %s blocked for %v after %d failed attempts", ip, l.BlockDuration, attempt.Count)
			}
		}
	}
}

// GetRemainingTime returns how much time is left on a block
func (l *LoginRateLimiter) GetRemainingTime(ip string) time.Duration {
	l.Mu.RLock()
	defer l.Mu.RUnlock()

	attempt, exists := l.Attempts[ip]
	if !exists {
		return 0
	}

	remaining := time.Until(attempt.BlockedUntil)
	if remaining < 0 {
		return 0
	}

	return remaining
}

// Cleanup removes old entries periodically
func (l *LoginRateLimiter) Cleanup() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		l.Mu.Lock()
		now := time.Now()
		for ip, attempt := range l.Attempts {
			// Remove entries that are old and not blocked
			if now.Sub(attempt.LastAttempt) > l.WindowDuration && now.After(attempt.BlockedUntil) {
				delete(l.Attempts, ip)
			}
		}
		l.Mu.Unlock()
	}
}