package models

import (
	"time"
)

// AutoReply represents an auto-reply rule
type AutoReply struct {
	ID          int       `json:"id"`
	SessionID   string    `json:"session_id"`
	Name        string    `json:"name"`
	Trigger     string    `json:"trigger"`     // "keyword", "all", "new_contact", "time_based"
	Keywords    []string  `json:"keywords,omitempty"`
	Response    string    `json:"response"`
	MediaURL    string    `json:"media_url,omitempty"`
	MediaType   string    `json:"media_type,omitempty"`
	IsActive    bool      `json:"is_active"`
	Priority    int       `json:"priority"`    // Higher number = higher priority
	DelayMin    int       `json:"delay_min"`   // Minimum delay in seconds
	DelayMax    int       `json:"delay_max"`   // Maximum delay in seconds
	MaxReplies  int       `json:"max_replies"` // Max replies per contact per day (0 = unlimited)
	TimeStart   string    `json:"time_start,omitempty"` // HH:MM format
	TimeEnd     string    `json:"time_end,omitempty"`   // HH:MM format
	Conditions  []AutoReplyCondition `json:"conditions,omitempty"`
	UsageCount  int       `json:"usage_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

// AutoReplyCondition represents conditions for auto-reply triggers
type AutoReplyCondition struct {
	Field    string `json:"field"`    // "message_type", "contact_group", "time_of_day", "day_of_week"
	Operator string `json:"operator"` // "equals", "contains", "not_equals", "in", "not_in"
	Value    string `json:"value"`
}

// AutoReplyLog represents a log of auto-reply actions
type AutoReplyLog struct {
	ID           int       `json:"id"`
	AutoReplyID  int       `json:"auto_reply_id"`
	SessionID    string    `json:"session_id"`
	ContactPhone string    `json:"contact_phone"`
	TriggerMsg   string    `json:"trigger_msg"`
	Response     string    `json:"response"`
	Success      bool      `json:"success"`
	ErrorMsg     string    `json:"error_msg,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateAutoReplyRequest represents auto-reply creation request
type CreateAutoReplyRequest struct {
	SessionID  string               `json:"session_id" validate:"required"`
	Name       string               `json:"name" validate:"required"`
	Trigger    string               `json:"trigger" validate:"required"`
	Keywords   []string             `json:"keywords,omitempty"`
	Response   string               `json:"response" validate:"required"`
	MediaURL   string               `json:"media_url,omitempty"`
	MediaType  string               `json:"media_type,omitempty"`
	Priority   int                  `json:"priority,omitempty"`
	DelayMin   int                  `json:"delay_min,omitempty"`
	DelayMax   int                  `json:"delay_max,omitempty"`
	MaxReplies int                  `json:"max_replies,omitempty"`
	TimeStart  string               `json:"time_start,omitempty"`
	TimeEnd    string               `json:"time_end,omitempty"`
	Conditions []AutoReplyCondition `json:"conditions,omitempty"`
}

// UpdateAutoReplyRequest represents auto-reply update request
type UpdateAutoReplyRequest struct {
	Name       string               `json:"name,omitempty"`
	Trigger    string               `json:"trigger,omitempty"`
	Keywords   []string             `json:"keywords,omitempty"`
	Response   string               `json:"response,omitempty"`
	MediaURL   string               `json:"media_url,omitempty"`
	MediaType  string               `json:"media_type,omitempty"`
	IsActive   *bool                `json:"is_active,omitempty"`
	Priority   int                  `json:"priority,omitempty"`
	DelayMin   int                  `json:"delay_min,omitempty"`
	DelayMax   int                  `json:"delay_max,omitempty"`
	MaxReplies int                  `json:"max_replies,omitempty"`
	TimeStart  string               `json:"time_start,omitempty"`
	TimeEnd    string               `json:"time_end,omitempty"`
	Conditions []AutoReplyCondition `json:"conditions,omitempty"`
}

// AutoReplyStats represents auto-reply statistics
type AutoReplyStats struct {
	TotalRules     int     `json:"total_rules"`
	ActiveRules    int     `json:"active_rules"`
	TotalTriggers  int     `json:"total_triggers"`
	SuccessRate    float64 `json:"success_rate"`
	AvgResponseTime float64 `json:"avg_response_time"`
}

// AutoReplyTestRequest represents auto-reply test request
type AutoReplyTestRequest struct {
	AutoReplyID int    `json:"auto_reply_id" validate:"required"`
	TestMessage string `json:"test_message" validate:"required"`
	TestPhone   string `json:"test_phone,omitempty"`
}

// AutoReplyTestResponse represents auto-reply test response
type AutoReplyTestResponse struct {
	WouldTrigger bool   `json:"would_trigger"`
	Response     string `json:"response,omitempty"`
	Delay        int    `json:"delay,omitempty"`
	Reason       string `json:"reason,omitempty"`
}