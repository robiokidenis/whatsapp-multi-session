package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// JobQueue represents a job in the queue system
type JobQueue struct {
	ID          int                    `json:"id"`
	JobID       string                 `json:"job_id"`
	Type        string                 `json:"type"` // "bulk_message", "scheduled_message", etc.
	Status      string                 `json:"status"` // "pending", "running", "completed", "failed", "cancelled", "scheduled"
	Priority    int                    `json:"priority"` // 1-10, higher number = higher priority
	Payload     map[string]interface{} `json:"payload"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   *time.Time             `json:"updated_at,omitempty"`
}

// JobQueueRequest represents a request to create a job
type JobQueueRequest struct {
	Type        string                 `json:"type" validate:"required"`
	Priority    int                    `json:"priority,omitempty"`
	Payload     map[string]interface{} `json:"payload" validate:"required"`
	MaxAttempts int                    `json:"max_attempts,omitempty"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
}

// JobQueueResponse represents the API response for job operations
type JobQueueResponse struct {
	JobID   string `json:"job_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// JobQueueListResponse represents paginated job list response
type JobQueueListResponse struct {
	Jobs  []JobQueue `json:"jobs"`
	Total int        `json:"total"`
	Page  int        `json:"page"`
	Limit int        `json:"limit"`
	Pages int        `json:"pages"`
}

// BulkMessagePayload represents the payload for bulk message jobs
type BulkMessagePayload struct {
	SessionID    string            `json:"session_id"`
	TemplateID   int               `json:"template_id"`
	ContactIDs   []int             `json:"contact_ids,omitempty"`
	GroupID      *int              `json:"group_id,omitempty"`
	DelayBetween int               `json:"delay_between"`
	RandomDelay  bool              `json:"random_delay"`
	Variables    map[string]string `json:"variables,omitempty"`
}

// ScheduledMessagePayload represents the payload for scheduled message jobs
type ScheduledMessagePayload struct {
	SessionID   string            `json:"session_id"`
	Phone       string            `json:"phone"`
	Message     string            `json:"message"`
	MessageType string            `json:"message_type"`
	MediaURL    string            `json:"media_url,omitempty"`
	Variables   map[string]string `json:"variables,omitempty"`
}

// JobStatistics represents job queue statistics
type JobStatistics struct {
	TotalJobs     int `json:"total_jobs"`
	PendingJobs   int `json:"pending_jobs"`
	RunningJobs   int `json:"running_jobs"`
	CompletedJobs int `json:"completed_jobs"`
	FailedJobs    int `json:"failed_jobs"`
	ScheduledJobs int `json:"scheduled_jobs"`
}

// MarshalPayload converts payload to JSON string for database storage
func (jq *JobQueue) MarshalPayload() ([]byte, error) {
	return json.Marshal(jq.Payload)
}

// UnmarshalPayload converts JSON string from database to payload map
func (jq *JobQueue) UnmarshalPayload(data []byte) error {
	return json.Unmarshal(data, &jq.Payload)
}

// MarshalResult converts result to JSON string for database storage
func (jq *JobQueue) MarshalResult() ([]byte, error) {
	return json.Marshal(jq.Result)
}

// UnmarshalResult converts JSON string from database to result map
func (jq *JobQueue) UnmarshalResult(data []byte) error {
	return json.Unmarshal(data, &jq.Result)
}

// IsScheduled returns true if the job is scheduled for future execution
func (jq *JobQueue) IsScheduled() bool {
	return jq.ScheduledAt != nil && jq.ScheduledAt.After(time.Now())
}

// IsReadyToRun returns true if the job is ready to be executed
func (jq *JobQueue) IsReadyToRun() bool {
	if jq.Status != "pending" && jq.Status != "scheduled" {
		return false
	}
	
	if jq.ScheduledAt != nil {
		return jq.ScheduledAt.Before(time.Now()) || jq.ScheduledAt.Equal(time.Now())
	}
	
	return true
}

// CanRetry returns true if the job can be retried
func (jq *JobQueue) CanRetry() bool {
	return jq.Status == "failed" && jq.Attempts < jq.MaxAttempts
}

// GetBulkMessagePayload converts payload to BulkMessagePayload
func (jq *JobQueue) GetBulkMessagePayload() (*BulkMessagePayload, error) {
	if jq.Type != "bulk_message" {
		return nil, fmt.Errorf("job type is not bulk_message")
	}
	
	payloadBytes, err := json.Marshal(jq.Payload)
	if err != nil {
		return nil, err
	}
	
	var payload BulkMessagePayload
	err = json.Unmarshal(payloadBytes, &payload)
	return &payload, err
}

// GetScheduledMessagePayload converts payload to ScheduledMessagePayload
func (jq *JobQueue) GetScheduledMessagePayload() (*ScheduledMessagePayload, error) {
	if jq.Type != "scheduled_message" {
		return nil, fmt.Errorf("job type is not scheduled_message")
	}
	
	payloadBytes, err := json.Marshal(jq.Payload)
	if err != nil {
		return nil, err
	}
	
	var payload ScheduledMessagePayload
	err = json.Unmarshal(payloadBytes, &payload)
	return &payload, err
}