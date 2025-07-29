package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"whatsapp-multi-session/internal/models"
	"whatsapp-multi-session/internal/repository"
	"whatsapp-multi-session/internal/services"
	"whatsapp-multi-session/pkg/logger"
)

type ContactHandler struct {
	contactRepo    *repository.ContactRepository
	groupRepo      *repository.ContactGroupRepository
	detectionSvc   *services.ContactDetectionService
	logger         *logger.Logger
}

func NewContactHandler(
	contactRepo *repository.ContactRepository,
	groupRepo *repository.ContactGroupRepository,
	detectionSvc *services.ContactDetectionService,
	logger *logger.Logger,
) *ContactHandler {
	return &ContactHandler{
		contactRepo:  contactRepo,
		groupRepo:    groupRepo,
		detectionSvc: detectionSvc,
		logger:       logger,
	}
}

// GetContacts handles GET /api/contacts
func (h *ContactHandler) GetContacts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query().Get("query")
	groupIDStr := r.URL.Query().Get("group_id")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	
	var groupID *int
	if groupIDStr != "" {
		if gid, err := strconv.Atoi(groupIDStr); err == nil {
			groupID = &gid
		}
	}
	
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	searchReq := models.ContactSearchRequest{
		Query:   query,
		GroupID: groupID,
		Page:    page,
		Limit:   limit,
	}
	
	response, err := h.contactRepo.GetContacts(searchReq)
	if err != nil {
		h.logger.Error("Failed to get contacts: %v", err)
		http.Error(w, "Failed to get contacts", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateContact handles POST /api/contacts
func (h *ContactHandler) CreateContact(w http.ResponseWriter, r *http.Request) {
	var contact models.Contact
	if err := json.NewDecoder(r.Body).Decode(&contact); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if err := h.contactRepo.CreateContact(&contact); err != nil {
		h.logger.Error("Failed to create contact: %v", err)
		http.Error(w, "Failed to create contact", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(contact)
}

// UpdateContact handles PUT /api/contacts/{id}
func (h *ContactHandler) UpdateContact(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contactID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}
	
	var updateReq models.UpdateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if err := h.contactRepo.UpdateContact(contactID, updateReq); err != nil {
		h.logger.Error("Failed to update contact: %v", err)
		http.Error(w, "Failed to update contact", http.StatusInternalServerError)
		return
	}
	
	// Return updated contact
	contact, err := h.contactRepo.GetContact(contactID)
	if err != nil {
		h.logger.Error("Failed to get updated contact: %v", err)
		http.Error(w, "Failed to get updated contact", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contact)
}

// DeleteContact handles DELETE /api/contacts/{id}
func (h *ContactHandler) DeleteContact(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contactID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}
	
	if err := h.contactRepo.DeleteContact(contactID); err != nil {
		h.logger.Error("Failed to delete contact: %v", err)
		http.Error(w, "Failed to delete contact", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// BulkActions handles POST /api/contacts/bulk
func (h *ContactHandler) BulkActions(w http.ResponseWriter, r *http.Request) {
	var request models.BulkContactRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if err := h.contactRepo.BulkUpdateContacts(request); err != nil {
		h.logger.Error("Failed to perform bulk action: %v", err)
		http.Error(w, "Failed to perform bulk action", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// DetectContacts handles POST /api/contacts/detect
func (h *ContactHandler) DetectContacts(w http.ResponseWriter, r *http.Request) {
	var detectedContacts []models.SmartContactDetection
	var err error
	
	contentType := r.Header.Get("Content-Type")
	
	if contentType == "application/json" {
		// Handle JSON request (text data)
		var request struct {
			Type string `json:"type"`
			Data string `json:"data"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		if request.Type == "text" {
			detectedContacts, err = h.detectionSvc.DetectFromText(request.Data)
		} else {
			http.Error(w, "Invalid type for JSON request", http.StatusBadRequest)
			return
		}
	} else {
		// Handle multipart form (file upload)
		if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB limit
			http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
			return
		}
		
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Failed to get file from request", http.StatusBadRequest)
			return
		}
		defer file.Close()
		
		detectedContacts, err = h.detectionSvc.DetectFromCSV(file)
	}
	
	if err != nil {
		h.logger.Error("Failed to detect contacts: %v", err)
		http.Error(w, "Failed to detect contacts", http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"contacts": detectedContacts,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ImportContacts handles POST /api/contacts/import
func (h *ContactHandler) ImportContacts(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Contacts []models.Contact `json:"contacts"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Use bulk create to import contacts
	result, err := h.contactRepo.BulkCreateContacts(request.Contacts)
	if err != nil {
		h.logger.Error("Failed to import contacts: %v", err)
		http.Error(w, "Failed to import contacts", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}