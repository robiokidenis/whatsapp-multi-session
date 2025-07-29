package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"whatsapp-multi-session/internal/repository"
	"whatsapp-multi-session/pkg/logger"
)

// LogHandler handles log-related HTTP requests
type LogHandler struct {
	logRepo *repository.LogRepository
	logger  *logger.Logger
}

// NewLogHandler creates a new log handler
func NewLogHandler(logRepo *repository.LogRepository, logger *logger.Logger) *LogHandler {
	return &LogHandler{
		logRepo: logRepo,
		logger:  logger,
	}
}

// LogsResponse represents the response structure for logs
type LogsResponse struct {
	Logs       []repository.LogEntry `json:"logs"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

// GetLogs retrieves logs with filtering and pagination
func (h *LogHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	
	filter := repository.LogFilter{}
	
	if level := query.Get("level"); level != "" {
		filter.Level = level
	}
	
	if component := query.Get("component"); component != "" {
		filter.Component = component
	}
	
	if sessionID := query.Get("session_id"); sessionID != "" {
		filter.SessionID = sessionID
	}
	
	if userIDStr := query.Get("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			filter.UserID = &userID
		}
	}
	
	if startTimeStr := query.Get("start_time"); startTimeStr != "" {
		if startTime, err := strconv.ParseInt(startTimeStr, 10, 64); err == nil {
			filter.StartTime = startTime
		}
	}
	
	if endTimeStr := query.Get("end_time"); endTimeStr != "" {
		if endTime, err := strconv.ParseInt(endTimeStr, 10, 64); err == nil {
			filter.EndTime = endTime
		}
	}
	
	// Pagination
	page := 1
	if pageStr := query.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	pageSize := 50
	if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 1000 {
			pageSize = ps
		}
	}
	
	filter.Limit = pageSize
	filter.Offset = (page - 1) * pageSize
	
	// Get logs
	logs, err := h.logRepo.GetLogs(filter)
	if err != nil {
		h.logger.Error("Failed to get logs: %v", err)
		http.Error(w, "Failed to retrieve logs", http.StatusInternalServerError)
		return
	}
	
	// Get total count
	total, err := h.logRepo.GetLogCount(filter)
	if err != nil {
		h.logger.Error("Failed to get log count: %v", err)
		http.Error(w, "Failed to retrieve log count", http.StatusInternalServerError)
		return
	}
	
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	
	response := LogsResponse{
		Logs:       logs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode logs response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetLogLevels returns available log levels
func (h *LogHandler) GetLogLevels(w http.ResponseWriter, r *http.Request) {
	levels := []string{
		logger.LevelDebug,
		logger.LevelInfo,
		logger.LevelWarn,
		logger.LevelError,
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string][]string{"levels": levels}); err != nil {
		h.logger.Error("Failed to encode log levels response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetLogComponents returns available log components
func (h *LogHandler) GetLogComponents(w http.ResponseWriter, r *http.Request) {
	// Get distinct components from database
	filter := repository.LogFilter{Limit: 1000}
	logs, err := h.logRepo.GetLogs(filter)
	if err != nil {
		h.logger.Error("Failed to get logs for components: %v", err)
		http.Error(w, "Failed to retrieve components", http.StatusInternalServerError)
		return
	}
	
	componentSet := make(map[string]bool)
	for _, log := range logs {
		if log.Component != "" {
			componentSet[log.Component] = true
		}
	}
	
	components := make([]string, 0, len(componentSet))
	for component := range componentSet {
		components = append(components, component)
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string][]string{"components": components}); err != nil {
		h.logger.Error("Failed to encode components response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// DeleteOldLogs deletes logs older than specified days
func (h *LogHandler) DeleteOldLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	daysStr := vars["days"]
	
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		http.Error(w, "Invalid days parameter", http.StatusBadRequest)
		return
	}
	
	cutoffTime := time.Now().AddDate(0, 0, -days).Unix()
	
	deletedCount, err := h.logRepo.DeleteOldLogs(cutoffTime)
	if err != nil {
		h.logger.Error("Failed to delete old logs: %v", err)
		http.Error(w, "Failed to delete old logs", http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"deleted_count": deletedCount,
		"cutoff_days":   days,
		"message":       "Old logs deleted successfully",
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode delete response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// ClearAllLogs deletes all logs
func (h *LogHandler) ClearAllLogs(w http.ResponseWriter, r *http.Request) {
	deletedCount, err := h.logRepo.DeleteOldLogs(time.Now().Unix() + 1) // Delete all logs (including current time + 1 second)
	if err != nil {
		h.logger.Error("Failed to clear all logs: %v", err)
		http.Error(w, "Failed to clear all logs", http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"deleted_count": deletedCount,
		"message":       "All logs cleared successfully",
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode clear response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}