package handlers

import (
	"encoding/json"
	"net/http"

	"whatsapp-multi-session/internal/models"
	"whatsapp-multi-session/internal/services"
	"whatsapp-multi-session/pkg/logger"
	"whatsapp-multi-session/pkg/ratelimiter"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	userService *services.UserService
	rateLimiter *ratelimiter.LoginRateLimiter
	logger      *logger.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	userService *services.UserService,
	rateLimiter *ratelimiter.LoginRateLimiter,
	log *logger.Logger,
) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		rateLimiter: rateLimiter,
		logger:      log,
	}
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Get client IP
	clientIP := getClientIP(r)

	// Check rate limiting
	if h.rateLimiter.IsBlocked(clientIP) {
		remaining := h.rateLimiter.GetRemainingTime(clientIP)
		h.logger.Warn("Blocked login attempt from %s, remaining: %v", clientIP, remaining)
		HandleErrorWithMessage(w, http.StatusTooManyRequests, "Too many failed attempts. Please try again later.", models.ErrCodeRateLimited)
		return
	}

	// Parse request
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.rateLimiter.RecordAttempt(clientIP, false)
		HandleErrorWithMessage(w, http.StatusBadRequest, "Invalid request body", models.ErrCodeInvalidInput)
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		h.rateLimiter.RecordAttempt(clientIP, false)
		HandleErrorWithMessage(w, http.StatusBadRequest, "Username and password are required", models.ErrCodeInvalidInput)
		return
	}

	// Attempt login
	response, err := h.userService.Login(&req)
	if err != nil {
		h.rateLimiter.RecordAttempt(clientIP, false)
		h.logger.Warn("Failed login attempt for %s from %s: %v", req.Username, clientIP, err)
		HandleErrorWithMessage(w, http.StatusUnauthorized, "Invalid credentials", models.ErrCodeUnauthorized)
		return
	}

	// Record successful attempt
	h.rateLimiter.RecordAttempt(clientIP, true)

	// Return response
	WriteSuccessResponse(w, "Login successful", response)
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HandleErrorWithMessage(w, http.StatusBadRequest, "Invalid request body", models.ErrCodeInvalidInput)
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		HandleErrorWithMessage(w, http.StatusBadRequest, "Username and password are required", models.ErrCodeInvalidInput)
		return
	}

	if len(req.Username) < 3 {
		HandleErrorWithMessage(w, http.StatusBadRequest, "Username must be at least 3 characters", models.ErrCodeInvalidInput)
		return
	}

	if len(req.Password) < 6 {
		HandleErrorWithMessage(w, http.StatusBadRequest, "Password must be at least 6 characters", models.ErrCodeInvalidInput)
		return
	}

	// Attempt registration
	response, err := h.userService.Register(&req)
	if err != nil {
		h.logger.Warn("Failed registration attempt for %s: %v", req.Username, err)
		HandleErrorWithMessage(w, http.StatusConflict, err.Error(), models.ErrCodeAlreadyExists)
		return
	}

	h.logger.Info("User %s registered successfully", req.Username)

	// Return response
	WriteSuccessResponse(w, "Registration successful", response)
}

// ChangePassword handles password change
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// TODO: Get user from JWT token context
	// For now, return method not implemented
	http.Error(w, "Method not implemented", http.StatusNotImplemented)
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP from the comma-separated list
		if idx := len(xff); idx > 0 {
			for i := 0; i < len(xff); i++ {
				if xff[i] == ',' {
					return xff[:i]
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