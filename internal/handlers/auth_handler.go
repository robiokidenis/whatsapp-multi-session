package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"whatsapp-multi-session/internal/middleware"
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
	// Get user claims from context
	claims, ok := middleware.GetUserClaims(r)
	if !ok {
		HandleErrorWithMessage(w, http.StatusUnauthorized, "Unauthorized", models.ErrCodeUnauthorized)
		return
	}

	// Parse request
	var req models.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HandleErrorWithMessage(w, http.StatusBadRequest, "Invalid request body", models.ErrCodeInvalidInput)
		return
	}

	// Validate input
	if req.OldPassword == "" || req.NewPassword == "" {
		HandleErrorWithMessage(w, http.StatusBadRequest, "Old password and new password are required", models.ErrCodeInvalidInput)
		return
	}

	if len(req.NewPassword) < 6 {
		HandleErrorWithMessage(w, http.StatusBadRequest, "New password must be at least 6 characters", models.ErrCodeInvalidInput)
		return
	}

	// Change password
	if err := h.userService.ChangePassword(claims.UserID, &req); err != nil {
		h.logger.Warn("Failed password change for user %d: %v", claims.UserID, err)
		HandleErrorWithMessage(w, http.StatusBadRequest, err.Error(), models.ErrCodeInvalidInput)
		return
	}

	h.logger.Info("Password changed successfully for user %d", claims.UserID)
	WriteSuccessResponse(w, "Password changed successfully", nil)
}

// GenerateAPIKey generates a new API key for the authenticated user
func (h *AuthHandler) GenerateAPIKey(w http.ResponseWriter, r *http.Request) {
	// Get user claims from context
	claims, ok := middleware.GetUserClaims(r)
	if !ok {
		HandleErrorWithMessage(w, http.StatusUnauthorized, "Unauthorized", models.ErrCodeUnauthorized)
		return
	}

	// Generate API key
	response, err := h.userService.GenerateAPIKey(claims.UserID)
	if err != nil {
		h.logger.Error("Failed to generate API key for user %d: %v", claims.UserID, err)
		HandleErrorWithMessage(w, http.StatusInternalServerError, "Failed to generate API key", models.ErrCodeInternalServer)
		return
	}

	h.logger.Info("Generated API key for user %d", claims.UserID)
	WriteSuccessResponse(w, "API key generated successfully", response)
}

// RevokeAPIKey revokes the API key for the authenticated user
func (h *AuthHandler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	// Get user claims from context
	claims, ok := middleware.GetUserClaims(r)
	if !ok {
		HandleErrorWithMessage(w, http.StatusUnauthorized, "Unauthorized", models.ErrCodeUnauthorized)
		return
	}

	// Revoke API key
	if err := h.userService.RevokeAPIKey(claims.UserID); err != nil {
		h.logger.Error("Failed to revoke API key for user %d: %v", claims.UserID, err)
		HandleErrorWithMessage(w, http.StatusInternalServerError, "Failed to revoke API key", models.ErrCodeInternalServer)
		return
	}

	h.logger.Info("Revoked API key for user %d", claims.UserID)
	WriteSuccessResponse(w, "API key revoked successfully", nil)
}

// GetAPIKeyInfo returns information about the user's API key (without the key itself)
func (h *AuthHandler) GetAPIKeyInfo(w http.ResponseWriter, r *http.Request) {
	// Get user claims from context
	claims, ok := middleware.GetUserClaims(r)
	if !ok {
		HandleErrorWithMessage(w, http.StatusUnauthorized, "Unauthorized", models.ErrCodeUnauthorized)
		return
	}

	// Get API key info
	info, err := h.userService.GetAPIKeyInfo(claims.UserID)
	if err != nil {
		h.logger.Error("Failed to get API key info for user %d: %v", claims.UserID, err)
		HandleErrorWithMessage(w, http.StatusInternalServerError, "Failed to get API key info", models.ErrCodeInternalServer)
		return
	}

	WriteSuccessResponse(w, "API key info retrieved successfully", info)
}

// AdminGenerateAPIKey generates API key for any user (admin only)
func (h *AuthHandler) AdminGenerateAPIKey(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL params
	vars := mux.Vars(r)
	userIDStr := vars["userId"]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		HandleErrorWithMessage(w, http.StatusBadRequest, "Invalid user ID", models.ErrCodeInvalidInput)
		return
	}

	// Generate API key
	response, err := h.userService.GenerateAPIKey(userID)
	if err != nil {
		h.logger.Error("Failed to generate API key for user %d: %v", userID, err)
		HandleErrorWithMessage(w, http.StatusInternalServerError, "Failed to generate API key", models.ErrCodeInternalServer)
		return
	}

	h.logger.Info("Admin generated API key for user %d", userID)
	WriteSuccessResponse(w, "API key generated successfully", response)
}

// AdminRevokeAPIKey revokes API key for any user (admin only)
func (h *AuthHandler) AdminRevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL params
	vars := mux.Vars(r)
	userIDStr := vars["userId"]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		HandleErrorWithMessage(w, http.StatusBadRequest, "Invalid user ID", models.ErrCodeInvalidInput)
		return
	}

	// Revoke API key
	if err := h.userService.RevokeAPIKey(userID); err != nil {
		h.logger.Error("Failed to revoke API key for user %d: %v", userID, err)
		HandleErrorWithMessage(w, http.StatusInternalServerError, "Failed to revoke API key", models.ErrCodeInternalServer)
		return
	}

	h.logger.Info("Admin revoked API key for user %d", userID)
	WriteSuccessResponse(w, "API key revoked successfully", nil)
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