package models

import (
	"time"
)

// Campaign represents a messaging campaign
type Campaign struct {
	ID              int                `json:"id"`
	Name            string             `json:"name"`
	Description     string             `json:"description,omitempty"`
	TemplateID      int                `json:"template_id"`
	Template        *MessageTemplate   `json:"template,omitempty"`
	GroupID         *int               `json:"group_id,omitempty"`
	Group           *ContactGroup      `json:"group,omitempty"`
	ContactIDs      []int              `json:"contact_ids,omitempty"`
	SessionID       string             `json:"session_id"`
	Status          string             `json:"status"` // "draft", "scheduled", "running", "paused", "completed", "failed"
	DelayBetween    int                `json:"delay_between"` // seconds between messages
	RandomDelay     bool               `json:"random_delay"`  // add random delay variation
	ScheduledAt     *time.Time         `json:"scheduled_at,omitempty"`
	StartedAt       *time.Time         `json:"started_at,omitempty"`
	CompletedAt     *time.Time         `json:"completed_at,omitempty"`
	TotalContacts   int                `json:"total_contacts"`
	SentCount       int                `json:"sent_count"`
	FailedCount     int                `json:"failed_count"`
	PendingCount    int                `json:"pending_count"`
	Variables       map[string]string  `json:"variables,omitempty"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       *time.Time         `json:"updated_at,omitempty"`
}

// CampaignMessage represents a message in a campaign
type CampaignMessage struct {
	ID         int        `json:"id"`
	CampaignID int        `json:"campaign_id"`
	ContactID  int        `json:"contact_id"`
	Contact    *Contact   `json:"contact,omitempty"`
	Content    string     `json:"content"`
	Status     string     `json:"status"` // "pending", "sending", "sent", "failed", "delivered", "read"
	ErrorMsg   string     `json:"error_msg,omitempty"`
	MessageID  string     `json:"message_id,omitempty"`
	SentAt     *time.Time `json:"sent_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// CreateCampaignRequest represents campaign creation request
type CreateCampaignRequest struct {
	Name         string            `json:"name" validate:"required"`
	Description  string            `json:"description,omitempty"`
	TemplateID   int               `json:"template_id" validate:"required"`
	GroupID      *int              `json:"group_id,omitempty"`
	ContactIDs   []int             `json:"contact_ids,omitempty"`
	SessionID    string            `json:"session_id" validate:"required"`
	DelayBetween int               `json:"delay_between,omitempty"`
	RandomDelay  bool              `json:"random_delay,omitempty"`
	ScheduledAt  *time.Time        `json:"scheduled_at,omitempty"`
	Variables    map[string]string `json:"variables,omitempty"`
}

// UpdateCampaignRequest represents campaign update request
type UpdateCampaignRequest struct {
	Name         string            `json:"name,omitempty"`
	Description  string            `json:"description,omitempty"`
	TemplateID   int               `json:"template_id,omitempty"`
	GroupID      *int              `json:"group_id,omitempty"`
	ContactIDs   []int             `json:"contact_ids,omitempty"`
	SessionID    string            `json:"session_id,omitempty"`
	DelayBetween int               `json:"delay_between,omitempty"`
	RandomDelay  bool              `json:"random_delay,omitempty"`
	ScheduledAt  *time.Time        `json:"scheduled_at,omitempty"`
	Variables    map[string]string `json:"variables,omitempty"`
	Status       string            `json:"status,omitempty"`
}

// CampaignStats represents campaign statistics
type CampaignStats struct {
	TotalCampaigns    int     `json:"total_campaigns"`
	ActiveCampaigns   int     `json:"active_campaigns"`
	CompletedCampaigns int    `json:"completed_campaigns"`
	TotalMessagesSent int     `json:"total_messages_sent"`
	SuccessRate       float64 `json:"success_rate"`
	AvgDeliveryTime   float64 `json:"avg_delivery_time"`
}

// CampaignListResponse represents paginated campaign list response
type CampaignListResponse struct {
	Campaigns []Campaign `json:"campaigns"`
	Total     int        `json:"total"`
	Page      int        `json:"page"`
	Limit     int        `json:"limit"`
	Pages     int        `json:"pages"`
}

// BulkMessageRequest represents a direct bulk message request (without campaigns)
type BulkMessageRequest struct {
	SessionID    string            `json:"session_id" validate:"required"`
	TemplateID   *int              `json:"template_id,omitempty"` // Optional - can send direct message without template
	Message      string            `json:"message,omitempty"`      // Direct message content when not using template
	ContactIDs   []int             `json:"contact_ids,omitempty"`
	GroupID      *int              `json:"group_id,omitempty"`
	DelayBetween int               `json:"delay_between,omitempty"`
	RandomDelay  bool              `json:"random_delay,omitempty"`
	Variables    map[string]string `json:"variables,omitempty"`
}

// BulkMessageResponse represents bulk message operation response
type BulkMessageResponse struct {
	JobID         string `json:"job_id"`
	TotalContacts int    `json:"total_contacts"`
	Status        string `json:"status"`
	Message       string `json:"message"`
}