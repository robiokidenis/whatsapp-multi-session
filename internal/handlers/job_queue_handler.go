package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"whatsapp-multi-session/internal/models"
	"whatsapp-multi-session/internal/services"
	"whatsapp-multi-session/pkg/logger"
)

// JobQueueHandler handles job queue API requests
type JobQueueHandler struct {
	jobQueueService *services.JobQueueService
	logger          *logger.Logger
}

// NewJobQueueHandler creates a new job queue handler
func NewJobQueueHandler(jobQueueService *services.JobQueueService, log *logger.Logger) *JobQueueHandler {
	return &JobQueueHandler{
		jobQueueService: jobQueueService,
		logger:          log,
	}
}

// GetJobs handles GET /api/job-queue
func (h *JobQueueHandler) GetJobs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "all"
	}
	
	jobType := r.URL.Query().Get("type")
	if jobType == "" {
		jobType = "all"
	}
	
	pageStr := r.URL.Query().Get("page")
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	offset := (page - 1) * limit
	
	// Get jobs from service
	jobs, total, err := h.jobQueueService.GetJobs(status, jobType, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get jobs: %v", err)
		http.Error(w, "Failed to get jobs", http.StatusInternalServerError)
		return
	}
	
	// Calculate pagination
	pages := (total + limit - 1) / limit
	
	response := models.JobQueueListResponse{
		Jobs:  make([]models.JobQueue, len(jobs)),
		Total: total,
		Page:  page,
		Limit: limit,
		Pages: pages,
	}
	
	// Convert pointers to values
	for i, job := range jobs {
		response.Jobs[i] = *job
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetJob handles GET /api/job-queue/{jobId}
func (h *JobQueueHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["jobId"]
	
	if jobID == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}
	
	job, err := h.jobQueueService.GetJob(jobID)
	if err != nil {
		h.logger.Error("Failed to get job %s: %v", jobID, err)
		http.Error(w, "Failed to get job", http.StatusInternalServerError)
		return
	}
	
	if job == nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// CreateJob handles POST /api/job-queue
func (h *JobQueueHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req models.JobQueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode job request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate request
	if req.Type == "" {
		http.Error(w, "Job type is required", http.StatusBadRequest)
		return
	}
	
	if req.Payload == nil {
		http.Error(w, "Job payload is required", http.StatusBadRequest)
		return
	}
	
	job, err := h.jobQueueService.EnqueueJob(&req)
	if err != nil {
		h.logger.Error("Failed to enqueue job: %v", err)
		http.Error(w, "Failed to create job", http.StatusInternalServerError)
		return
	}
	
	response := models.JobQueueResponse{
		JobID:   job.JobID,
		Status:  job.Status,
		Message: "Job created successfully",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// CancelJob handles DELETE /api/job-queue/{jobId}
func (h *JobQueueHandler) CancelJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["jobId"]
	
	if jobID == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}
	
	err := h.jobQueueService.CancelJob(jobID)
	if err != nil {
		h.logger.Error("Failed to cancel job %s: %v", jobID, err)
		
		// Check for specific error types
		if err.Error() == "job not found" {
			http.Error(w, "Job not found", http.StatusNotFound)
			return
		}
		
		if err.Error() == "job cannot be cancelled" {
			http.Error(w, "Job cannot be cancelled", http.StatusBadRequest)
			return
		}
		
		http.Error(w, "Failed to cancel job", http.StatusInternalServerError)
		return
	}
	
	response := models.JobQueueResponse{
		JobID:   jobID,
		Status:  "cancelled",
		Message: "Job cancelled successfully",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RetryJob handles POST /api/job-queue/{jobId}/retry
func (h *JobQueueHandler) RetryJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["jobId"]
	
	if jobID == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}
	
	err := h.jobQueueService.RetryJob(jobID)
	if err != nil {
		h.logger.Error("Failed to retry job %s: %v", jobID, err)
		
		// Check for specific error types
		if err.Error() == "job not found" {
			http.Error(w, "Job not found", http.StatusNotFound)
			return
		}
		
		if err.Error() == "job is not in failed state" || err.Error() == "job has exceeded maximum retry attempts" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		http.Error(w, "Failed to retry job", http.StatusInternalServerError)
		return
	}
	
	response := models.JobQueueResponse{
		JobID:   jobID,
		Status:  "pending",
		Message: "Job queued for retry",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetStatistics handles GET /api/job-queue/statistics
func (h *JobQueueHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.jobQueueService.GetStatistics()
	if err != nil {
		h.logger.Error("Failed to get job statistics: %v", err)
		http.Error(w, "Failed to get statistics", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// CleanupJobs handles POST /api/job-queue/cleanup
func (h *JobQueueHandler) CleanupJobs(w http.ResponseWriter, r *http.Request) {
	// Parse cleanup duration (default: 7 days)
	daysStr := r.URL.Query().Get("days")
	days := 7
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}
	
	duration := time.Duration(days) * 24 * time.Hour
	count, err := h.jobQueueService.CleanupOldJobs(duration)
	if err != nil {
		h.logger.Error("Failed to cleanup jobs: %v", err)
		http.Error(w, "Failed to cleanup jobs", http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"message":      "Jobs cleaned up successfully",
		"deleted_jobs": count,
		"older_than":   fmt.Sprintf("%d days", days),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateBulkMessageJob handles POST /api/job-queue/bulk-message
func (h *JobQueueHandler) CreateBulkMessageJob(w http.ResponseWriter, r *http.Request) {
	var req models.BulkMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode bulk message request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Parse scheduled time if provided
	var scheduledAt *time.Time
	scheduledAtStr := r.URL.Query().Get("scheduled_at")
	if scheduledAtStr != "" {
		if t, err := time.Parse(time.RFC3339, scheduledAtStr); err == nil {
			scheduledAt = &t
		}
	}
	
	job, err := h.jobQueueService.EnqueueBulkMessage(&req, scheduledAt)
	if err != nil {
		h.logger.Error("Failed to enqueue bulk message job: %v", err)
		http.Error(w, "Failed to create bulk message job", http.StatusInternalServerError)
		return
	}
	
	response := models.JobQueueResponse{
		JobID:   job.JobID,
		Status:  job.Status,
		Message: "Bulk message job created successfully",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// CreateScheduledMessageJob handles POST /api/job-queue/scheduled-message
func (h *JobQueueHandler) CreateScheduledMessageJob(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID   string            `json:"session_id"`
		Phone       string            `json:"phone"`
		Message     string            `json:"message"`
		MessageType string            `json:"message_type"`
		MediaURL    string            `json:"media_url,omitempty"`
		Variables   map[string]string `json:"variables,omitempty"`
		ScheduledAt time.Time         `json:"scheduled_at"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode scheduled message request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate required fields
	if req.SessionID == "" || req.Phone == "" || req.Message == "" {
		http.Error(w, "session_id, phone, and message are required", http.StatusBadRequest)
		return
	}
	
	job, err := h.jobQueueService.EnqueueScheduledMessage(
		req.SessionID, req.Phone, req.Message, req.MessageType, 
		req.MediaURL, req.Variables, &req.ScheduledAt,
	)
	if err != nil {
		h.logger.Error("Failed to enqueue scheduled message job: %v", err)
		http.Error(w, "Failed to create scheduled message job", http.StatusInternalServerError)
		return
	}
	
	response := models.JobQueueResponse{
		JobID:   job.JobID,
		Status:  job.Status,
		Message: "Scheduled message job created successfully",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}