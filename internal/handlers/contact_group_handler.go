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

type ContactGroupHandler struct {
	groupRepo *repository.ContactGroupRepository
	logger    *logger.Logger
}

func NewContactGroupHandler(
	groupRepo *repository.ContactGroupRepository,
	logger *logger.Logger,
) *ContactGroupHandler {
	return &ContactGroupHandler{
		groupRepo: groupRepo,
		logger:    logger,
	}
}

// GetContactGroups handles GET /api/contact-groups
func (h *ContactGroupHandler) GetContactGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := h.groupRepo.GetContactGroups()
	if err != nil {
		h.logger.Error("Failed to get contact groups: %v", err)
		http.Error(w, "Failed to get contact groups", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

// CreateContactGroup handles POST /api/contact-groups
func (h *ContactGroupHandler) CreateContactGroup(w http.ResponseWriter, r *http.Request) {
	var req models.CreateContactGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Convert request to ContactGroup model
	group := &models.ContactGroup{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		IsActive:    true,
	}
	
	if err := h.groupRepo.CreateContactGroup(group); err != nil {
		h.logger.Error("Failed to create contact group: %v", err)
		http.Error(w, "Failed to create contact group", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(group)
}

// UpdateContactGroup handles PUT /api/contact-groups/{id}
func (h *ContactGroupHandler) UpdateContactGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}
	
	var req models.UpdateContactGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if err := h.groupRepo.UpdateContactGroup(groupID, req); err != nil {
		h.logger.Error("Failed to update contact group: %v", err)
		http.Error(w, "Failed to update contact group", http.StatusInternalServerError)
		return
	}
	
	// Get updated group
	group, err := h.groupRepo.GetContactGroup(groupID)
	if err != nil {
		h.logger.Error("Failed to get updated contact group: %v", err)
		http.Error(w, "Failed to get updated contact group", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}

// DeleteContactGroup handles DELETE /api/contact-groups/{id}
func (h *ContactGroupHandler) DeleteContactGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}
	
	if err := h.groupRepo.DeleteContactGroup(groupID); err != nil {
		h.logger.Error("Failed to delete contact group: %v", err)
		http.Error(w, "Failed to delete contact group", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}