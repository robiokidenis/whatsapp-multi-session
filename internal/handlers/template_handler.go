package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"whatsapp-multi-session/internal/models"
	"whatsapp-multi-session/internal/repository"
	"whatsapp-multi-session/pkg/logger"
)

type TemplateHandler struct {
	templateRepo *repository.TemplateRepository
	contactRepo  *repository.ContactRepository
	logger       *logger.Logger
}

func NewTemplateHandler(
	templateRepo *repository.TemplateRepository,
	contactRepo *repository.ContactRepository,
	logger *logger.Logger,
) *TemplateHandler {
	return &TemplateHandler{
		templateRepo: templateRepo,
		contactRepo:  contactRepo,
		logger:       logger,
	}
}

// GetMessageTemplates handles GET /api/message-templates
func (h *TemplateHandler) GetMessageTemplates(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	
	// Check for is_active filter
	isActiveStr := r.URL.Query().Get("is_active")
	var isActive *bool
	if isActiveStr == "true" {
		active := true
		isActive = &active
	} else if isActiveStr == "false" {
		active := false
		isActive = &active
	}
	
	templates, err := h.templateRepo.GetMessageTemplates(userID, isActive)
	if err != nil {
		h.logger.Error("Failed to get message templates: %v", err)
		http.Error(w, "Failed to get message templates", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

// GetMessageTemplateCategories handles GET /api/message-templates/categories
func (h *TemplateHandler) GetMessageTemplateCategories(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	
	categories, err := h.templateRepo.GetTemplateCategories(userID)
	if err != nil {
		h.logger.Error("Failed to get template categories: %v", err)
		http.Error(w, "Failed to get template categories", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// CreateMessageTemplate handles POST /api/message-templates
func (h *TemplateHandler) CreateMessageTemplate(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	
	var template models.MessageTemplate
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	template.UserID = userID
	
	if err := h.templateRepo.CreateMessageTemplate(&template); err != nil {
		h.logger.Error("Failed to create message template: %v", err)
		http.Error(w, "Failed to create message template", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(template)
}

// UpdateMessageTemplate handles PUT /api/message-templates/{id}
func (h *TemplateHandler) UpdateMessageTemplate(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	vars := mux.Vars(r)
	templateID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid template ID", http.StatusBadRequest)
		return
	}
	
	var template models.MessageTemplate
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	template.ID = templateID
	template.UserID = userID
	
	if err := h.templateRepo.UpdateMessageTemplate(&template); err != nil {
		h.logger.Error("Failed to update message template: %v", err)
		http.Error(w, "Failed to update message template", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

// DeleteMessageTemplate handles DELETE /api/message-templates/{id}
func (h *TemplateHandler) DeleteMessageTemplate(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	vars := mux.Vars(r)
	templateID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid template ID", http.StatusBadRequest)
		return
	}
	
	if err := h.templateRepo.DeleteMessageTemplate(userID, templateID); err != nil {
		h.logger.Error("Failed to delete message template: %v", err)
		http.Error(w, "Failed to delete message template", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// PreviewMessageTemplate handles POST /api/message-templates/preview
func (h *TemplateHandler) PreviewMessageTemplate(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	
	var request struct {
		TemplateID int                    `json:"template_id"`
		ContactID  *int                   `json:"contact_id,omitempty"`
		Variables  map[string]interface{} `json:"variables,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Get template
	template, err := h.templateRepo.GetMessageTemplate(userID, request.TemplateID)
	if err != nil {
		h.logger.Error("Failed to get message template: %v", err)
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}
	
	// Get contact data if provided
	var contact *models.Contact
	if request.ContactID != nil {
		contact, err = h.contactRepo.GetContact(*request.ContactID)
		if err != nil {
			h.logger.Error("Failed to get contact: %v", err)
			// Don't fail if contact not found, just continue without contact data
		}
	}
	
	// Apply variables to template
	content := template.Content
	
	// Apply contact variables if available
	if contact != nil {
		content = strings.ReplaceAll(content, "[name]", contact.Name)
		content = strings.ReplaceAll(content, "[phone]", contact.Phone)
		if contact.Email != "" {
			content = strings.ReplaceAll(content, "[email]", contact.Email)
		}
		if contact.Company != "" {
			content = strings.ReplaceAll(content, "[company]", contact.Company)
		}
		if contact.Position != "" {
			content = strings.ReplaceAll(content, "[position]", contact.Position)
		}
	}
	
	// Apply custom variables
	if request.Variables != nil {
		for key, value := range request.Variables {
			placeholder := "[" + key + "]"
			if strValue, ok := value.(string); ok {
				content = strings.ReplaceAll(content, placeholder, strValue)
			}
		}
	}
	
	response := map[string]interface{}{
		"content": content,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}