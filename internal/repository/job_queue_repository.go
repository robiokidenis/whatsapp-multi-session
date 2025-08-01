package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"whatsapp-multi-session/internal/models"
)

// JobQueueRepository handles job queue database operations
type JobQueueRepository struct {
	db *sql.DB
}

// NewJobQueueRepository creates a new job queue repository
func NewJobQueueRepository(db *sql.DB) *JobQueueRepository {
	return &JobQueueRepository{db: db}
}

// Create adds a new job to the queue
func (r *JobQueueRepository) Create(req *models.JobQueueRequest) (*models.JobQueue, error) {
	now := time.Now()
	jobID := uuid.New().String()
	
	// Marshal payload
	payloadJSON, err := json.Marshal(req.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}
	
	// Set defaults
	maxAttempts := req.MaxAttempts
	if maxAttempts == 0 {
		maxAttempts = 3
	}
	
	priority := req.Priority
	if priority == 0 {
		priority = 5
	}
	
	// Determine initial status
	status := "pending"
	if req.ScheduledAt != nil && req.ScheduledAt.After(now) {
		status = "scheduled"
	}
	
	var scheduledAtUnix *int64
	if req.ScheduledAt != nil {
		unix := req.ScheduledAt.Unix()
		scheduledAtUnix = &unix
	}
	
	query := `
		INSERT INTO job_queue (
			job_id, type, status, priority, payload, max_attempts, 
			scheduled_at, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	
	_, err = r.db.Exec(query, 
		jobID, req.Type, status, priority, string(payloadJSON), maxAttempts,
		scheduledAtUnix, now.Unix())
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %v", err)
	}
	
	return r.GetByJobID(jobID)
}

// GetByJobID retrieves a job by its job ID
func (r *JobQueueRepository) GetByJobID(jobID string) (*models.JobQueue, error) {
	query := `
		SELECT id, job_id, type, status, priority, payload, result, error,
			   attempts, max_attempts, scheduled_at, started_at, completed_at,
			   created_at, updated_at
		FROM job_queue 
		WHERE job_id = ?`
	
	return r.scanJob(r.db.QueryRow(query, jobID))
}

// GetByID retrieves a job by its database ID
func (r *JobQueueRepository) GetByID(id int) (*models.JobQueue, error) {
	query := `
		SELECT id, job_id, type, status, priority, payload, result, error,
			   attempts, max_attempts, scheduled_at, started_at, completed_at,
			   created_at, updated_at
		FROM job_queue 
		WHERE id = ?`
	
	return r.scanJob(r.db.QueryRow(query, id))
}

// GetPendingJobs retrieves jobs ready to be processed
func (r *JobQueueRepository) GetPendingJobs(limit int) ([]*models.JobQueue, error) {
	query := `
		SELECT id, job_id, type, status, priority, payload, result, error,
			   attempts, max_attempts, scheduled_at, started_at, completed_at,
			   created_at, updated_at
		FROM job_queue 
		WHERE (status = 'pending' OR (status = 'scheduled' AND scheduled_at <= ?))
		  AND attempts < max_attempts
		ORDER BY priority DESC, created_at ASC
		LIMIT ?`
	
	rows, err := r.db.Query(query, time.Now().Unix(), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending jobs: %v", err)
	}
	defer rows.Close()
	
	var jobs []*models.JobQueue
	for rows.Next() {
		job, err := r.scanJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	
	return jobs, nil
}

// GetJobs retrieves jobs with filtering and pagination
func (r *JobQueueRepository) GetJobs(status string, jobType string, limit, offset int) ([]*models.JobQueue, int, error) {
	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	
	if status != "" && status != "all" {
		whereClause += " AND status = ?"
		args = append(args, status)
	}
	
	if jobType != "" && jobType != "all" {
		whereClause += " AND type = ?"
		args = append(args, jobType)
	}
	
	// Get total count
	countQuery := "SELECT COUNT(*) FROM job_queue " + whereClause
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count jobs: %v", err)
	}
	
	// Get jobs
	query := `
		SELECT id, job_id, type, status, priority, payload, result, error,
			   attempts, max_attempts, scheduled_at, started_at, completed_at,
			   created_at, updated_at
		FROM job_queue ` + whereClause + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`
	
	args = append(args, limit, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get jobs: %v", err)
	}
	defer rows.Close()
	
	var jobs []*models.JobQueue
	for rows.Next() {
		job, err := r.scanJob(rows)
		if err != nil {
			return nil, 0, err
		}
		jobs = append(jobs, job)
	}
	
	return jobs, total, nil
}

// UpdateStatus updates a job's status and related fields
func (r *JobQueueRepository) UpdateStatus(jobID string, status string, result map[string]interface{}, errorMsg string) error {
	now := time.Now()
	
	var resultJSON *string
	if result != nil {
		resultBytes, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %v", err)
		}
		resultStr := string(resultBytes)
		resultJSON = &resultStr
	}
	
	query := `
		UPDATE job_queue 
		SET status = ?, result = ?, error = ?, updated_at = ?`
	args := []interface{}{status, resultJSON, errorMsg, now.Unix()}
	
	// Add status-specific fields
	if status == "running" {
		query += ", started_at = ?"
		args = append(args, now.Unix())
	} else if status == "completed" || status == "failed" {
		query += ", completed_at = ?"
		args = append(args, now.Unix())
	}
	
	query += " WHERE job_id = ?"
	args = append(args, jobID)
	
	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update job status: %v", err)
	}
	
	return nil
}

// IncrementAttempts increments the job's attempt counter
func (r *JobQueueRepository) IncrementAttempts(jobID string) error {
	query := `
		UPDATE job_queue 
		SET attempts = attempts + 1, updated_at = ?
		WHERE job_id = ?`
	
	_, err := r.db.Exec(query, time.Now().Unix(), jobID)
	if err != nil {
		return fmt.Errorf("failed to increment attempts: %v", err)
	}
	
	return nil
}

// Delete removes a job from the queue
func (r *JobQueueRepository) Delete(jobID string) error {
	query := "DELETE FROM job_queue WHERE job_id = ?"
	
	result, err := r.db.Exec(query, jobID)
	if err != nil {
		return fmt.Errorf("failed to delete job: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %v", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("job not found")
	}
	
	return nil
}

// GetStatistics returns job queue statistics
func (r *JobQueueRepository) GetStatistics() (*models.JobStatistics, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending,
			SUM(CASE WHEN status = 'running' THEN 1 ELSE 0 END) as running,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed,
			SUM(CASE WHEN status = 'scheduled' THEN 1 ELSE 0 END) as scheduled
		FROM job_queue`
	
	var stats models.JobStatistics
	err := r.db.QueryRow(query).Scan(
		&stats.TotalJobs,
		&stats.PendingJobs,
		&stats.RunningJobs,
		&stats.CompletedJobs,
		&stats.FailedJobs,
		&stats.ScheduledJobs,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %v", err)
	}
	
	return &stats, nil
}

// CleanupOldJobs removes completed/failed jobs older than the specified duration
func (r *JobQueueRepository) CleanupOldJobs(olderThan time.Duration) (int, error) {
	cutoff := time.Now().Add(-olderThan)
	
	query := `
		DELETE FROM job_queue 
		WHERE (status = 'completed' OR status = 'failed') 
		  AND completed_at < ?`
	
	result, err := r.db.Exec(query, cutoff.Unix())
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old jobs: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %v", err)
	}
	
	return int(rowsAffected), nil
}

// scanJob scans a database row into a JobQueue struct
func (r *JobQueueRepository) scanJob(scanner interface{}) (*models.JobQueue, error) {
	var job models.JobQueue
	var payloadStr string
	var resultStr sql.NullString
	var errorStr sql.NullString
	var scheduledAt, startedAt, completedAt, updatedAt sql.NullInt64
	var createdAt int64
	
	var err error
	switch s := scanner.(type) {
	case *sql.Row:
		err = s.Scan(
			&job.ID, &job.JobID, &job.Type, &job.Status, &job.Priority,
			&payloadStr, &resultStr, &errorStr, &job.Attempts, &job.MaxAttempts,
			&scheduledAt, &startedAt, &completedAt, &createdAt, &updatedAt,
		)
	case *sql.Rows:
		err = s.Scan(
			&job.ID, &job.JobID, &job.Type, &job.Status, &job.Priority,
			&payloadStr, &resultStr, &errorStr, &job.Attempts, &job.MaxAttempts,
			&scheduledAt, &startedAt, &completedAt, &createdAt, &updatedAt,
		)
	default:
		return nil, fmt.Errorf("unsupported scanner type")
	}
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan job: %v", err)
	}
	
	// Parse payload JSON
	if err := json.Unmarshal([]byte(payloadStr), &job.Payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %v", err)
	}
	
	// Parse result JSON if present
	if resultStr.Valid {
		if err := json.Unmarshal([]byte(resultStr.String), &job.Result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %v", err)
		}
	}
	
	// Set nullable fields
	if errorStr.Valid {
		job.Error = errorStr.String
	}
	
	// Convert Unix timestamps to time.Time
	if scheduledAt.Valid {
		t := time.Unix(scheduledAt.Int64, 0)
		job.ScheduledAt = &t
	}
	
	if startedAt.Valid {
		t := time.Unix(startedAt.Int64, 0)
		job.StartedAt = &t
	}
	
	if completedAt.Valid {
		t := time.Unix(completedAt.Int64, 0)
		job.CompletedAt = &t
	}
	
	// CreatedAt is always present
	job.CreatedAt = time.Unix(createdAt, 0)
	
	if updatedAt.Valid {
		t := time.Unix(updatedAt.Int64, 0)
		job.UpdatedAt = &t
	}
	
	return &job, nil
}