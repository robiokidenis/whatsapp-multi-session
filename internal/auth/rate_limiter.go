package auth

import (
	"time"

	"whatsapp-multi-session/internal/types"
)

// NewLoginRateLimiter creates a new rate limiter with default settings
func NewLoginRateLimiter() *types.LoginRateLimiter {
	limiter := &types.LoginRateLimiter{
		Attempts:       make(map[string]*types.LoginAttempt),
		MaxAttempts:    5,
		BlockDuration:  15 * time.Minute,
		WindowDuration: 1 * time.Hour,
	}

	// Start cleanup goroutine
	go limiter.Cleanup()

	return limiter
}