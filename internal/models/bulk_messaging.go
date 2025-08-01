package models

import "time"

// BulkMessageResult represents the result of a bulk message job

type BulkMessageResult struct {
	JobID         string     `json:"job_id"`
	TotalContacts int        `json:"total_contacts"`
	SentCount     int        `json:"sent_count"`
	FailedCount   int        `json:"failed_count"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}
