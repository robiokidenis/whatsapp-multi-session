package models

import (
	"time"
)

// Contact represents a contact in the CRM system
type Contact struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Phone       string    `json:"phone"`
	Email       string    `json:"email,omitempty"`
	Company     string    `json:"company,omitempty"`
	Position    string    `json:"position,omitempty"`
	GroupID     *int      `json:"group_id,omitempty"`
	Group       *ContactGroup `json:"group,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Notes       string    `json:"notes,omitempty"`
	IsActive    bool      `json:"is_active"`
	LastContact *time.Time `json:"last_contact,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

// ContactGroup represents a contact group for marketing campaigns
type ContactGroup struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Color       string    `json:"color,omitempty"`
	IsActive    bool      `json:"is_active"`
	ContactCount int      `json:"contact_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

// CreateContactRequest represents contact creation request
type CreateContactRequest struct {
	Name     string `json:"name" validate:"required"`
	Phone    string `json:"phone" validate:"required"`
	Email    string `json:"email,omitempty"`
	Company  string `json:"company,omitempty"`
	Position string `json:"position,omitempty"`
	GroupID  *int   `json:"group_id,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Notes    string `json:"notes,omitempty"`
}

// UpdateContactRequest represents contact update request
type UpdateContactRequest struct {
	Name     string   `json:"name,omitempty"`
	Phone    string   `json:"phone,omitempty"`
	Email    string   `json:"email,omitempty"`
	Company  string   `json:"company,omitempty"`
	Position string   `json:"position,omitempty"`
	GroupID  *int     `json:"group_id,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Notes    string   `json:"notes,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}

// CreateContactGroupRequest represents contact group creation request
type CreateContactGroupRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
}

// UpdateContactGroupRequest represents contact group update request
type UpdateContactGroupRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	IsActive    *bool  `json:"is_active,omitempty"`
}

// ContactImportResult represents the result of a contact import operation
type ContactImportResult struct {
	Success     int      `json:"success"`
	Failed      int      `json:"failed"`
	Duplicates  int      `json:"duplicates"`
	Total       int      `json:"total"`
	Errors      []string `json:"errors,omitempty"`
}

// BulkContactRequest represents a bulk contact operation request
type BulkContactRequest struct {
	ContactIDs []int  `json:"contact_ids" validate:"required"`
	Action     string `json:"action" validate:"required"` // "delete", "activate", "deactivate", "move_to_group"
	GroupID    *int   `json:"group_id,omitempty"`
}

// ContactSearchRequest represents contact search parameters
type ContactSearchRequest struct {
	Query   string `json:"query,omitempty"`
	GroupID *int   `json:"group_id,omitempty"`
	Tags    []string `json:"tags,omitempty"`
	IsActive *bool  `json:"is_active,omitempty"`
	Page    int    `json:"page,omitempty"`
	Limit   int    `json:"limit,omitempty"`
}

// ContactListResponse represents paginated contact list response
type ContactListResponse struct {
	Contacts []Contact `json:"contacts"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
	Pages    int       `json:"pages"`
}

// SmartContactDetection represents detected contact information from CSV/text
type SmartContactDetection struct {
	Name        string   `json:"name"`
	Phone       string   `json:"phone"`
	Email       string   `json:"email,omitempty"`
	Company     string   `json:"company,omitempty"`
	Position    string   `json:"position,omitempty"`
	Confidence  float64  `json:"confidence"`
	Source      string   `json:"source"`
	RawData     string   `json:"raw_data"`
}