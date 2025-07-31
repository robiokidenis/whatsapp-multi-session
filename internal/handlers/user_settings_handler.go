package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"whatsapp-multi-session/pkg/logger"
	"whatsapp-multi-session/internal/models"
	"whatsapp-multi-session/internal/repository"
)

// UserSettingsHandler handles user settings API requests
type UserSettingsHandler struct {
	settingsRepo *repository.UserSettingsRepository
	logger       *logger.Logger
}

// NewUserSettingsHandler creates a new user settings handler
func NewUserSettingsHandler(settingsRepo *repository.UserSettingsRepository, log *logger.Logger) *UserSettingsHandler {
	return &UserSettingsHandler{
		settingsRepo: settingsRepo,
		logger:       log,
	}
}

// GetUserSettings retrieves user settings
func (h *UserSettingsHandler) GetUserSettings(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by JWT middleware)
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		h.logger.Error("Failed to get user ID from context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get settings from database
	settings, err := h.settingsRepo.GetByUserID(userID)
	if err != nil {
		h.logger.Error("Failed to get user settings for user %d: %v", userID, err)
		http.Error(w, "Failed to get user settings", http.StatusInternalServerError)
		return
	}

	// If no settings found, return defaults
	if settings == nil {
		defaultSettings := models.GetDefaultSettings()
		defaultSettings.UserID = userID
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(defaultSettings.ToResponse())
		return
	}

	// Return settings
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings.ToResponse())
}

// UpdateUserSettings creates or updates user settings
func (h *UserSettingsHandler) UpdateUserSettings(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by JWT middleware)
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		h.logger.Error("Failed to get user ID from context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req models.CreateUserSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate timezone (basic validation)
	if req.Timezone == "" {
		req.Timezone = "UTC"
	}
	if req.DateFormat == "" {
		req.DateFormat = "YYYY-MM-DD"
	}
	if req.TimeFormat == "" {
		req.TimeFormat = "24h"
	}
	if req.Language == "" {
		req.Language = "en"
	}

	// Create or update settings
	settings, err := h.settingsRepo.CreateOrUpdate(userID, &req)
	if err != nil {
		h.logger.Error("Failed to save user settings for user %d: %v", userID, err)
		http.Error(w, "Failed to save user settings", http.StatusInternalServerError)
		return
	}

	h.logger.Info("User settings updated for user %d, timezone: %s", userID, settings.Timezone)

	// Return updated settings
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings.ToResponse())
}

// PatchUserSettings partially updates user settings
func (h *UserSettingsHandler) PatchUserSettings(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by JWT middleware)
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		h.logger.Error("Failed to get user ID from context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req models.UpdateUserSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update settings
	settings, err := h.settingsRepo.Update(userID, &req)
	if err != nil {
		h.logger.Error("Failed to update user settings for user %d: %v", userID, err)
		http.Error(w, "Failed to update user settings", http.StatusInternalServerError)
		return
	}

	if settings == nil {
		http.Error(w, "User settings not found", http.StatusNotFound)
		return
	}

	h.logger.Info("User settings partially updated for user %d", userID)

	// Return updated settings
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings.ToResponse())
}

// DeleteUserSettings removes user settings (reset to defaults)
func (h *UserSettingsHandler) DeleteUserSettings(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by JWT middleware)
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		h.logger.Error("Failed to get user ID from context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete settings
	err := h.settingsRepo.Delete(userID)
	if err != nil {
		h.logger.Error("Failed to delete user settings for user %d: %v", userID, err)
		http.Error(w, "Failed to delete user settings", http.StatusInternalServerError)
		return
	}

	h.logger.Info("User settings deleted for user %d", userID)

	w.WriteHeader(http.StatusNoContent)
}

// GetUserSettingsByID retrieves user settings by user ID (admin only)
func (h *UserSettingsHandler) GetUserSettingsByID(w http.ResponseWriter, r *http.Request) {
	// Check if user is admin (you might want to implement role-based access)
	userRole, ok := r.Context().Value("user_role").(string)
	if !ok || userRole != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get user ID from URL path
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get settings from database
	settings, err := h.settingsRepo.GetByUserID(userID)
	if err != nil {
		h.logger.Error("Failed to get user settings by ID for user %d: %v", userID, err)
		http.Error(w, "Failed to get user settings", http.StatusInternalServerError)
		return
	}

	// If no settings found, return defaults
	if settings == nil {
		defaultSettings := models.GetDefaultSettings()
		defaultSettings.UserID = userID
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(defaultSettings.ToResponse())
		return
	}

	// Return settings
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings.ToResponse())
}