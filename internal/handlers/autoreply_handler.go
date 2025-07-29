package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"whatsapp-multi-session/internal/models"
	"whatsapp-multi-session/internal/repository"
	"whatsapp-multi-session/pkg/logger"
)

type AutoReplyHandler struct {
	autoReplyRepo *repository.AutoReplyRepository
	logger        *logger.Logger
}

func NewAutoReplyHandler(
	autoReplyRepo *repository.AutoReplyRepository,
	logger *logger.Logger,
) *AutoReplyHandler {
	return &AutoReplyHandler{
		autoReplyRepo: autoReplyRepo,
		logger:        logger,
	}
}

// GetAutoReplies handles GET /api/auto-replies
func (h *AutoReplyHandler) GetAutoReplies(w http.ResponseWriter, r *http.Request) {
	// Get session ID from query parameter
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "session_id parameter is required", http.StatusBadRequest)
		return
	}
	
	autoReplies, err := h.autoReplyRepo.GetAutoRepliesBySession(sessionID)
	if err != nil {
		h.logger.Error("Failed to get auto replies: %v", err)
		http.Error(w, "Failed to get auto replies", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(autoReplies)
}

// CreateAutoReply handles POST /api/auto-replies
func (h *AutoReplyHandler) CreateAutoReply(w http.ResponseWriter, r *http.Request) {
	var autoReply models.AutoReply
	if err := json.NewDecoder(r.Body).Decode(&autoReply); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if err := h.autoReplyRepo.CreateAutoReply(&autoReply); err != nil {
		h.logger.Error("Failed to create auto reply: %v", err)
		http.Error(w, "Failed to create auto reply", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(autoReply)
}

// UpdateAutoReply handles PUT /api/auto-replies/{id}
func (h *AutoReplyHandler) UpdateAutoReply(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	autoReplyID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid auto reply ID", http.StatusBadRequest)
		return
	}
	
	var updateReq models.UpdateAutoReplyRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if err := h.autoReplyRepo.UpdateAutoReply(autoReplyID, updateReq); err != nil {
		h.logger.Error("Failed to update auto reply: %v", err)
		http.Error(w, "Failed to update auto reply", http.StatusInternalServerError)
		return
	}
	
	// Get updated auto reply
	autoReply, err := h.autoReplyRepo.GetAutoReply(autoReplyID)
	if err != nil {
		h.logger.Error("Failed to get updated auto reply: %v", err)
		http.Error(w, "Failed to get updated auto reply", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(autoReply)
}

// DeleteAutoReply handles DELETE /api/auto-replies/{id}
func (h *AutoReplyHandler) DeleteAutoReply(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	autoReplyID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid auto reply ID", http.StatusBadRequest)
		return
	}
	
	if err := h.autoReplyRepo.DeleteAutoReply(autoReplyID); err != nil {
		h.logger.Error("Failed to delete auto reply: %v", err)
		http.Error(w, "Failed to delete auto reply", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}