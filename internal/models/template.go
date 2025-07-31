package models

import (
	"time"
)

// MessageTemplate represents a message template for campaigns
type MessageTemplate struct {
	ID          int                    `json:"id"`
	UserID      int                    `json:"user_id"`
	Name        string                 `json:"name"`
	Content     string                 `json:"content"`
	Type        string                 `json:"type"` // "text", "image", "document", "location"
	Variables   []TemplateVariable     `json:"variables,omitempty"`
	MediaURL    string                 `json:"media_url,omitempty"`
	MediaType   string                 `json:"media_type,omitempty"`
	Category    string                 `json:"category,omitempty"`
	IsActive    bool                   `json:"is_active"`
	UsageCount  int                    `json:"usage_count"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   *time.Time             `json:"updated_at,omitempty"`
}

// TemplateVariable represents a variable in a message template
type TemplateVariable struct {
	Name         string `json:"name"`
	Placeholder  string `json:"placeholder"`
	DefaultValue string `json:"default_value,omitempty"`
	Required     bool   `json:"required"`
}

// CreateTemplateRequest represents template creation request
type CreateTemplateRequest struct {
	Name      string             `json:"name" validate:"required"`
	Content   string             `json:"content" validate:"required"`
	Type      string             `json:"type" validate:"required"`
	Variables []TemplateVariable `json:"variables,omitempty"`
	MediaURL  string             `json:"media_url,omitempty"`
	MediaType string             `json:"media_type,omitempty"`
	Category  string             `json:"category,omitempty"`
}

// UpdateTemplateRequest represents template update request
type UpdateTemplateRequest struct {
	Name      string             `json:"name,omitempty"`
	Content   string             `json:"content,omitempty"`
	Type      string             `json:"type,omitempty"`
	Variables []TemplateVariable `json:"variables,omitempty"`
	MediaURL  string             `json:"media_url,omitempty"`
	MediaType string             `json:"media_type,omitempty"`
	Category  string             `json:"category,omitempty"`
	IsActive  *bool              `json:"is_active,omitempty"`
}

// TemplatePreviewRequest represents template preview request
type TemplatePreviewRequest struct {
	TemplateID int                    `json:"template_id" validate:"required"`
	Variables  map[string]string      `json:"variables,omitempty"`
	ContactID  *int                   `json:"contact_id,omitempty"`
}

// TemplatePreviewResponse represents template preview response
type TemplatePreviewResponse struct {
	Content   string `json:"content"`
	Type      string `json:"type"`
	MediaURL  string `json:"media_url,omitempty"`
	MediaType string `json:"media_type,omitempty"`
}