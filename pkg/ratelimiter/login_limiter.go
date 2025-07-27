package ratelimiter

import (
	"sync"
	"time"
)

// LoginAttempt tracks login attempts for rate limiting
type LoginAttempt struct {
	Count        int       `json:"count"`
	LastAttempt  time.Time `json:"last_attempt"`
	BlockedUntil time.Time `json:"blocked_until"`
}

// LoginRateLimiter manages login rate limiting
type LoginRateLimiter struct {
	attempts map[string]*LoginAttempt
	mu       sync.RWMutex

	// Configuration
	MaxAttempts     int           // Max attempts before blocking
	BlockDuration   time.Duration // How long to block after max attempts
	WindowDuration  time.Duration // Time window to count attempts
	CleanupInterval time.Duration // How often to clean up old attempts
}

// NewLoginRateLimiter creates a new rate limiter
func NewLoginRateLimiter() *LoginRateLimiter {
	limiter := &LoginRateLimiter{
		attempts:        make(map[string]*LoginAttempt),
		MaxAttempts:     5,                // 5 attempts
		BlockDuration:   15 * time.Minute, // Block for 15 minutes
		WindowDuration:  5 * time.Minute,  // 5 minute window
		CleanupInterval: 30 * time.Minute, // Cleanup every 30 minutes
	}

	// Start cleanup routine
	go limiter.cleanup()

	return limiter
}

// IsBlocked checks if an IP is currently blocked
func (l *LoginRateLimiter) IsBlocked(ip string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	attempt, exists := l.attempts[ip]
	if !exists {
		return false
	}

	return time.Now().Before(attempt.BlockedUntil)
}

// RecordAttempt records a login attempt for an IP
func (l *LoginRateLimiter) RecordAttempt(ip string, success bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	attempt, exists := l.attempts[ip]

	if !exists {
		attempt = &LoginAttempt{}
		l.attempts[ip] = attempt
	}

	if success {
		// Reset on successful login
		attempt.Count = 0
		attempt.BlockedUntil = time.Time{}
		return
	}

	// Check if we're in a new window
	if now.Sub(attempt.LastAttempt) > l.WindowDuration {
		attempt.Count = 1
	} else {
		attempt.Count++
	}

	attempt.LastAttempt = now

	// Block if max attempts reached
	if attempt.Count >= l.MaxAttempts {
		attempt.BlockedUntil = now.Add(l.BlockDuration)
	}
}

// GetRemainingTime returns how long until the IP is unblocked
func (l *LoginRateLimiter) GetRemainingTime(ip string) time.Duration {
	l.mu.RLock()
	defer l.mu.RUnlock()

	attempt, exists := l.attempts[ip]
	if !exists {
		return 0
	}

	remaining := attempt.BlockedUntil.Sub(time.Now())
	if remaining < 0 {
		return 0
	}

	return remaining
}

// cleanup removes old entries periodically
func (l *LoginRateLimiter) cleanup() {
	ticker := time.NewTicker(l.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		now := time.Now()

		for ip, attempt := range l.attempts {
			// Remove if not blocked and last attempt was long ago
			if now.After(attempt.BlockedUntil) && now.Sub(attempt.LastAttempt) > l.WindowDuration*2 {
				delete(l.attempts, ip)
			}
		}

		l.mu.Unlock()
	}
}