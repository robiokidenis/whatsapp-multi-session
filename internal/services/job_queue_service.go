package services

import (
	"fmt"
	"sync"
	"time"

	"whatsapp-multi-session/internal/models"
	"whatsapp-multi-session/internal/repository"
	"whatsapp-multi-session/pkg/logger"
)

// JobQueueService manages the job queue and execution
type JobQueueService struct {
	repo                *repository.JobQueueRepository
	bulkMessagingService *BulkMessagingService
	logger              *logger.Logger
	isRunning           bool
	stopChan            chan struct{}
	wg                  sync.WaitGroup
	mu                  sync.RWMutex
	maxWorkers          int
	pollInterval        time.Duration
}

// NewJobQueueService creates a new job queue service
func NewJobQueueService(
	repo *repository.JobQueueRepository,
	bulkMessagingService *BulkMessagingService,
	logger *logger.Logger,
) *JobQueueService {
	return &JobQueueService{
		repo:                 repo,
		bulkMessagingService: bulkMessagingService,
		logger:               logger,
		maxWorkers:           5,
		pollInterval:         5 * time.Second,
		stopChan:             make(chan struct{}),
	}
}

// Start starts the job queue processor
func (s *JobQueueService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.isRunning {
		return fmt.Errorf("job queue service is already running")
	}
	
	s.isRunning = true
	s.logger.Info("Starting job queue service with %d workers", s.maxWorkers)
	
	// Start worker goroutines
	for i := 0; i < s.maxWorkers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}
	
	return nil
}

// Stop stops the job queue processor
func (s *JobQueueService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.isRunning {
		return nil
	}
	
	s.logger.Info("Stopping job queue service...")
	s.isRunning = false
	close(s.stopChan)
	
	// Wait for all workers to finish
	s.wg.Wait()
	s.logger.Info("Job queue service stopped")
	
	return nil
}

// EnqueueJob adds a new job to the queue
func (s *JobQueueService) EnqueueJob(req *models.JobQueueRequest) (*models.JobQueue, error) {
	job, err := s.repo.Create(req)
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue job: %v", err)
	}
	
	s.logger.Info("Job enqueued: %s (type: %s)", job.JobID, job.Type)
	return job, nil
}

// GetJob retrieves a job by ID
func (s *JobQueueService) GetJob(jobID string) (*models.JobQueue, error) {
	return s.repo.GetByJobID(jobID)
}

// GetJobs retrieves jobs with filtering and pagination
func (s *JobQueueService) GetJobs(status, jobType string, limit, offset int) ([]*models.JobQueue, int, error) {
	return s.repo.GetJobs(status, jobType, limit, offset)
}

// CancelJob cancels a pending or scheduled job
func (s *JobQueueService) CancelJob(jobID string) error {
	job, err := s.repo.GetByJobID(jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %v", err)
	}
	
	if job == nil {
		return fmt.Errorf("job not found")
	}
	
	if job.Status != "pending" && job.Status != "scheduled" {
		return fmt.Errorf("job cannot be cancelled (status: %s)", job.Status)
	}
	
	err = s.repo.UpdateStatus(jobID, "cancelled", nil, "Job cancelled by user")
	if err != nil {
		return fmt.Errorf("failed to cancel job: %v", err)
	}
	
	s.logger.Info("Job cancelled: %s", jobID)
	return nil
}

// RetryJob retries a failed job
func (s *JobQueueService) RetryJob(jobID string) error {
	job, err := s.repo.GetByJobID(jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %v", err)
	}
	
	if job == nil {
		return fmt.Errorf("job not found")
	}
	
	if job.Status != "failed" {
		return fmt.Errorf("job is not in failed state")
	}
	
	if !job.CanRetry() {
		return fmt.Errorf("job has exceeded maximum retry attempts")
	}
	
	err = s.repo.UpdateStatus(jobID, "pending", nil, "")
	if err != nil {
		return fmt.Errorf("failed to retry job: %v", err)
	}
	
	s.logger.Info("Job queued for retry: %s", jobID)
	return nil
}

// GetStatistics returns job queue statistics
func (s *JobQueueService) GetStatistics() (*models.JobStatistics, error) {
	return s.repo.GetStatistics()
}

// CleanupOldJobs removes old completed/failed jobs
func (s *JobQueueService) CleanupOldJobs(olderThan time.Duration) (int, error) {
	count, err := s.repo.CleanupOldJobs(olderThan)
	if err != nil {
		return 0, err
	}
	
	s.logger.Info("Cleaned up %d old jobs", count)
	return count, nil
}

// worker processes jobs from the queue
func (s *JobQueueService) worker(workerID int) {
	defer s.wg.Done()
	
	s.logger.Info("Worker %d started", workerID)
	
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.stopChan:
			s.logger.Info("Worker %d stopping", workerID)
			return
		case <-ticker.C:
			s.processJobs(workerID)
		}
	}
}

// processJobs processes available jobs
func (s *JobQueueService) processJobs(workerID int) {
	jobs, err := s.repo.GetPendingJobs(10) // Get up to 10 jobs at a time
	if err != nil {
		s.logger.Error("Worker %d failed to get pending jobs: %v", workerID, err)
		return
	}
	
	for _, job := range jobs {
		select {
		case <-s.stopChan:
			return
		default:
			s.processJob(workerID, job)
		}
	}
}

// processJob processes a single job
func (s *JobQueueService) processJob(workerID int, job *models.JobQueue) {
	s.logger.Info("Worker %d processing job %s (type: %s)", workerID, job.JobID, job.Type)
	
	// Update job status to running
	if err := s.repo.UpdateStatus(job.JobID, "running", nil, ""); err != nil {
		s.logger.Error("Failed to update job status to running: %v", err)
		return
	}
	
	// Increment attempts
	if err := s.repo.IncrementAttempts(job.JobID); err != nil {
		s.logger.Error("Failed to increment job attempts: %v", err)
	}
	
	// Process the job based on its type
	var result map[string]interface{}
	var jobError error
	
	switch job.Type {
	case "bulk_message":
		result, jobError = s.processBulkMessageJob(job)
	case "scheduled_message":
		result, jobError = s.processScheduledMessageJob(job)
	default:
		jobError = fmt.Errorf("unknown job type: %s", job.Type)
	}
	
	// Update job status based on result
	if jobError != nil {
		s.logger.Error("Worker %d job %s failed: %v", workerID, job.JobID, jobError)
		
		// Check if we should retry
		if job.Attempts+1 >= job.MaxAttempts {
			s.repo.UpdateStatus(job.JobID, "failed", result, jobError.Error())
		} else {
			s.repo.UpdateStatus(job.JobID, "pending", result, jobError.Error())
		}
	} else {
		s.logger.Info("Worker %d job %s completed successfully", workerID, job.JobID)
		s.repo.UpdateStatus(job.JobID, "completed", result, "")
	}
}

// processBulkMessageJob processes a bulk message job
func (s *JobQueueService) processBulkMessageJob(job *models.JobQueue) (map[string]interface{}, error) {
	payload, err := job.GetBulkMessagePayload()
	if err != nil {
		return nil, fmt.Errorf("failed to get bulk message payload: %v", err)
	}

	var templateID *int
	// TemplateID is already an int in the payload struct, not an interface{}
	if payload.TemplateID > 0 {
		templateID = &payload.TemplateID
	}
	
	// Message is already a string in the payload struct
	message := payload.Message
	
	req := models.BulkMessageRequest{
		SessionID:    payload.SessionID,
		TemplateID:   templateID,
		Message:      message,
		ContactIDs:   payload.ContactIDs,
		GroupID:      payload.GroupID,
		DelayBetween: payload.DelayBetween,
		RandomDelay:  payload.RandomDelay,
		Variables:    payload.Variables,
	}

	result, err := s.bulkMessagingService.ExecuteBulkMessage(&req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute bulk message: %v", err)
	}

	return map[string]interface{}{
		"total_contacts": result.TotalContacts,
		"sent_count":     result.SentCount,
		"failed_count":   result.FailedCount,
		"started_at":     result.StartedAt,
		"completed_at":   result.CompletedAt,
	}, nil
}

// processScheduledMessageJob processes a scheduled message job
func (s *JobQueueService) processScheduledMessageJob(job *models.JobQueue) (map[string]interface{}, error) {
	payload, err := job.GetScheduledMessagePayload()
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled message payload: %v", err)
	}
	
	// TODO: Integrate with WhatsApp service to send the message
	s.logger.Info("Sending scheduled message to %s via session %s", payload.Phone, payload.SessionID)
	
	// Simulate message sending
	time.Sleep(500 * time.Millisecond)
	
	result := map[string]interface{}{
		"phone":       payload.Phone,
		"session_id":  payload.SessionID,
		"message_id":  fmt.Sprintf("msg_%d", time.Now().Unix()),
		"sent_at":     time.Now().Unix(),
	}
	
	return result, nil
}

// EnqueueBulkMessage creates a bulk message job
func (s *JobQueueService) EnqueueBulkMessage(req *models.BulkMessageRequest, scheduledAt *time.Time) (*models.JobQueue, error) {
	payload := map[string]interface{}{
		"session_id":     req.SessionID,
		"template_id":    req.TemplateID,
		"message":        req.Message,
		"contact_ids":    req.ContactIDs,
		"group_id":       req.GroupID,
		"delay_between":  req.DelayBetween,
		"random_delay":   req.RandomDelay,
		"variables":      req.Variables,
	}
	
	jobReq := &models.JobQueueRequest{
		Type:        "bulk_message",
		Priority:    7, // Higher priority for bulk messages
		Payload:     payload,
		MaxAttempts: 3,
		ScheduledAt: scheduledAt,
	}
	
	return s.EnqueueJob(jobReq)
}

// EnqueueScheduledMessage creates a scheduled message job
func (s *JobQueueService) EnqueueScheduledMessage(sessionID, phone, message, messageType, mediaURL string, variables map[string]string, scheduledAt *time.Time) (*models.JobQueue, error) {
	payload := map[string]interface{}{
		"session_id":   sessionID,
		"phone":        phone,
		"message":      message,
		"message_type": messageType,
		"media_url":    mediaURL,
		"variables":    variables,
	}
	
	jobReq := &models.JobQueueRequest{
		Type:        "scheduled_message",
		Priority:    5, // Normal priority for individual messages
		Payload:     payload,
		MaxAttempts: 3,
		ScheduledAt: scheduledAt,
	}
	
	return s.EnqueueJob(jobReq)
}