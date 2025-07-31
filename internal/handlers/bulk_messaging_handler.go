package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"whatsapp-multi-session/internal/models"
	"whatsapp-multi-session/internal/services"
	"whatsapp-multi-session/pkg/logger"
)

type BulkMessagingHandler struct {
	jobQueueService *services.JobQueueService
	logger          *logger.Logger
}

func NewBulkMessagingHandler(
	jobQueueService *services.JobQueueService,
	logger *logger.Logger,
) *BulkMessagingHandler {
	return &BulkMessagingHandler{
		jobQueueService: jobQueueService,
		logger:          logger,
	}
}

// StartBulkMessaging handles POST /api/bulk-messages
func (h *BulkMessagingHandler) StartBulkMessaging(w http.ResponseWriter, r *http.Request) {
	var bulkReq models.BulkMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&bulkReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Create bulk message job using job queue service
	job, err := h.jobQueueService.EnqueueBulkMessage(&bulkReq, nil)
	if err != nil {
		h.logger.Error("Failed to enqueue bulk messaging job: %v", err)
		http.Error(w, "Failed to start bulk messaging", http.StatusInternalServerError)
		return
	}
	
	response := models.JobQueueResponse{
		JobID:   job.JobID,
		Status:  job.Status,
		Message: "Bulk messaging job created successfully",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetBulkMessagingJob handles GET /api/bulk-messages/{jobId}
func (h *BulkMessagingHandler) GetBulkMessagingJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["jobId"]
	
	job, err := h.jobQueueService.GetJob(jobID)
	if err != nil {
		h.logger.Error("Failed to get bulk messaging job: %v", err)
		http.Error(w, "Failed to get bulk messaging job", http.StatusInternalServerError)
		return
	}
	
	if job == nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// GetBulkMessagingJobs handles GET /api/bulk-messages
func (h *BulkMessagingHandler) GetBulkMessagingJobs(w http.ResponseWriter, r *http.Request) {
	// Get only bulk_message jobs from the queue
	jobs, total, err := h.jobQueueService.GetJobs("all", "bulk_message", 50, 0)
	if err != nil {
		h.logger.Error("Failed to get bulk messaging jobs: %v", err)
		http.Error(w, "Failed to get bulk messaging jobs", http.StatusInternalServerError)
		return
	}
	
	response := models.JobQueueListResponse{
		Jobs:  make([]models.JobQueue, len(jobs)),
		Total: total,
		Page:  1,
		Limit: 50,
		Pages: (total + 49) / 50,
	}
	
	// Convert pointers to values
	for i, job := range jobs {
		response.Jobs[i] = *job
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CancelBulkMessagingJob handles DELETE /api/bulk-messages/{jobId}
func (h *BulkMessagingHandler) CancelBulkMessagingJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["jobId"]
	
	if err := h.jobQueueService.CancelJob(jobID); err != nil {
		h.logger.Error("Failed to cancel bulk messaging job: %v", err)
		
		// Check for specific error types
		if err.Error() == "job not found" {
			http.Error(w, "Job not found", http.StatusNotFound)
			return
		}
		
		if err.Error() == "job cannot be cancelled" {
			http.Error(w, "Job cannot be cancelled", http.StatusBadRequest)
			return
		}
		
		http.Error(w, "Failed to cancel bulk messaging job", http.StatusInternalServerError)
		return
	}
	
	response := models.JobQueueResponse{
		JobID:   jobID,
		Status:  "cancelled",
		Message: "Bulk messaging job cancelled successfully",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}