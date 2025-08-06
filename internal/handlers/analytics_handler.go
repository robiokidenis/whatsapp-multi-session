package handlers

import (
	"encoding/json"
	"net/http"
	
	"whatsapp-multi-session/internal/middleware"
	"whatsapp-multi-session/internal/services"
	"whatsapp-multi-session/pkg/logger"
)

type AnalyticsHandler struct {
	analyticsService *services.AnalyticsService
	log              *logger.Logger
}

func NewAnalyticsHandler(analyticsService *services.AnalyticsService, log *logger.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		log:              log,
	}
}

// GetAnalytics handles GET /api/analytics
func (h *AnalyticsHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	// Get user context from middleware using the proper context key
	userClaims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.Claims)
	if !ok {
		h.log.Error("Failed to get user claims from context")
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	
	// Get time range from query params
	timeRange := r.URL.Query().Get("timeRange")
	if timeRange == "" {
		timeRange = "week" // Default to week
	}
	
	// Validate time range
	validRanges := map[string]bool{
		"today": true,
		"week":  true,
		"month": true,
		"year":  true,
	}
	
	if !validRanges[timeRange] {
		respondWithError(w, http.StatusBadRequest, "Invalid time range. Must be one of: today, week, month, year")
		return
	}
	
	// Check if user is admin
	isAdmin := userClaims.Role == "admin"
	
	// Get analytics data
	analytics, err := h.analyticsService.GetAnalytics(int64(userClaims.UserID), isAdmin, timeRange)
	if err != nil {
		h.log.Error("Failed to get analytics: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve analytics")
		return
	}
	
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    analytics,
	})
}

// GetMessageStats handles GET /api/analytics/messages
func (h *AnalyticsHandler) GetMessageStats(w http.ResponseWriter, r *http.Request) {
	// Get user context from middleware using the proper context key
	userClaims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.Claims)
	if !ok {
		h.log.Error("Failed to get user claims from context")
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	
	// Get time range from query params
	timeRange := r.URL.Query().Get("timeRange")
	if timeRange == "" {
		timeRange = "week"
	}
	
	// Check if user is admin
	isAdmin := userClaims.Role == "admin"
	
	// Get message stats
	stats, err := h.analyticsService.GetMessageStats(int64(userClaims.UserID), isAdmin, timeRange)
	if err != nil {
		h.log.Error("Failed to get message stats: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve message statistics")
		return
	}
	
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// GetSessionStats handles GET /api/analytics/sessions
func (h *AnalyticsHandler) GetSessionStats(w http.ResponseWriter, r *http.Request) {
	// Get user context from middleware using the proper context key
	userClaims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.Claims)
	if !ok {
		h.log.Error("Failed to get user claims from context")
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	
	// Check if user is admin
	isAdmin := userClaims.Role == "admin"
	
	// Get session stats
	stats, err := h.analyticsService.GetSessionStats(int64(userClaims.UserID), isAdmin)
	if err != nil {
		h.log.Error("Failed to get session stats: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve session statistics")
		return
	}
	
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// Helper functions for JSON responses
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]interface{}{
		"success": false,
		"error":   message,
	})
}