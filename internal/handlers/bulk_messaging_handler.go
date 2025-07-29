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
	bulkService *services.BulkMessagingService
	logger      *logger.Logger
}

func NewBulkMessagingHandler(
	bulkService *services.BulkMessagingService,
	logger *logger.Logger,
) *BulkMessagingHandler {
	return &BulkMessagingHandler{
		bulkService: bulkService,
		logger:      logger,
	}
}

// StartBulkMessaging handles POST /api/bulk-messages
func (h *BulkMessagingHandler) StartBulkMessaging(w http.ResponseWriter, r *http.Request) {
	var bulkReq models.BulkMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&bulkReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// For now, we'll use the direct bulk message service
	// TODO: Get template and contacts from database based on request
	var template *models.MessageTemplate
	var contacts []models.Contact
	
	// Start bulk messaging
	job, err := h.bulkService.StartBulkMessage(bulkReq, template, contacts)
	if err != nil {
		h.logger.Error("Failed to start bulk messaging: %v", err)
		http.Error(w, "Failed to start bulk messaging", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(job)
}

// GetBulkMessagingJob handles GET /api/bulk-messages/{jobId}
func (h *BulkMessagingHandler) GetBulkMessagingJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["jobId"]
	
	job, err := h.bulkService.GetJob(jobID)
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
	jobs := h.bulkService.GetJobs()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

// CancelBulkMessagingJob handles DELETE /api/bulk-messages/{jobId}
func (h *BulkMessagingHandler) CancelBulkMessagingJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["jobId"]
	
	if err := h.bulkService.CancelJob(jobID); err != nil {
		h.logger.Error("Failed to cancel bulk messaging job: %v", err)
		http.Error(w, "Failed to cancel bulk messaging job", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}