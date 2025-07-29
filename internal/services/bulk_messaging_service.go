package services

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"whatsapp-multi-session/internal/models"
	"whatsapp-multi-session/pkg/logger"
)

type BulkMessageJob struct {
	ID           string                 `json:"id"`
	CampaignID   *int                   `json:"campaign_id,omitempty"`
	SessionID    string                 `json:"session_id"`
	Template     *models.MessageTemplate `json:"template"`
	Contacts     []models.Contact       `json:"contacts"`
	DelayBetween int                    `json:"delay_between"` // seconds
	RandomDelay  bool                   `json:"random_delay"`
	Variables    map[string]string      `json:"variables,omitempty"`
	Status       string                 `json:"status"` // "pending", "running", "paused", "completed", "failed"
	Progress     BulkMessageProgress    `json:"progress"`
	CreatedAt    time.Time              `json:"created_at"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	ctx          context.Context
	cancel       context.CancelFunc
}

type BulkMessageProgress struct {
	Total     int `json:"total"`
	Sent      int `json:"sent"`
	Failed    int `json:"failed"`
	Remaining int `json:"remaining"`
}

type BulkMessagingService struct {
	whatsappService *WhatsAppService
	jobs            map[string]*BulkMessageJob
	jobsMutex       sync.RWMutex
	log             logger.Logger
}

func NewBulkMessagingService(whatsappService *WhatsAppService, log logger.Logger) *BulkMessagingService {
	return &BulkMessagingService{
		whatsappService: whatsappService,
		jobs:            make(map[string]*BulkMessageJob),
		log:             log,
	}
}

// StartBulkMessage creates and starts a new bulk messaging job
func (s *BulkMessagingService) StartBulkMessage(req models.BulkMessageRequest, template *models.MessageTemplate, contacts []models.Contact) (*BulkMessageJob, error) {
	jobID := s.generateJobID()
	ctx, cancel := context.WithCancel(context.Background())
	
	job := &BulkMessageJob{
		ID:           jobID,
		SessionID:    req.SessionID,
		Template:     template,
		Contacts:     contacts,
		DelayBetween: req.DelayBetween,
		RandomDelay:  req.RandomDelay,
		Variables:    req.Variables,
		Status:       "pending",
		Progress: BulkMessageProgress{
			Total:     len(contacts),
			Sent:      0,
			Failed:    0,
			Remaining: len(contacts),
		},
		CreatedAt: time.Now(),
		ctx:       ctx,
		cancel:    cancel,
	}
	
	s.jobsMutex.Lock()
	s.jobs[jobID] = job
	s.jobsMutex.Unlock()
	
	// Start processing in background
	go s.processJob(job)
	
	s.log.Info("Started bulk messaging job %s for session %s with %d contacts", jobID, req.SessionID, len(contacts))
	
	return job, nil
}

// StartCampaignMessages creates and starts bulk messaging for a campaign
func (s *BulkMessagingService) StartCampaignMessages(campaign *models.Campaign, template *models.MessageTemplate, contacts []models.Contact) (*BulkMessageJob, error) {
	jobID := s.generateJobID()
	ctx, cancel := context.WithCancel(context.Background())
	
	job := &BulkMessageJob{
		ID:           jobID,
		CampaignID:   &campaign.ID,
		SessionID:    campaign.SessionID,
		Template:     template,
		Contacts:     contacts,
		DelayBetween: campaign.DelayBetween,
		RandomDelay:  campaign.RandomDelay,
		Variables:    campaign.Variables,
		Status:       "pending",
		Progress: BulkMessageProgress{
			Total:     len(contacts),
			Sent:      0,
			Failed:    0,
			Remaining: len(contacts),
		},
		CreatedAt: time.Now(),
		ctx:       ctx,
		cancel:    cancel,
	}
	
	s.jobsMutex.Lock()
	s.jobs[jobID] = job
	s.jobsMutex.Unlock()
	
	// Start processing in background
	go s.processJob(job)
	
	s.log.Info("Started campaign bulk messaging job %s for campaign %d with %d contacts", jobID, campaign.ID, len(contacts))
	
	return job, nil
}

// processJob processes messages for a bulk messaging job
func (s *BulkMessagingService) processJob(job *BulkMessageJob) {
	defer func() {
		if r := recover(); r != nil {
			s.log.Error("Bulk messaging job %s panicked: %v", job.ID, r)
			job.Status = "failed"
		}
	}()
	
	job.Status = "running"
	now := time.Now()
	job.StartedAt = &now
	
	s.log.Info("Processing bulk messaging job %s", job.ID)
	
	for i, contact := range job.Contacts {
		select {
		case <-job.ctx.Done():
			s.log.Info("Bulk messaging job %s was cancelled", job.ID)
			job.Status = "cancelled"
			return
		default:
		}
		
		// Process individual message
		success := s.processMessage(job, contact, i)
		
		// Update progress
		s.jobsMutex.Lock()
		if success {
			job.Progress.Sent++
		} else {
			job.Progress.Failed++
		}
		job.Progress.Remaining = job.Progress.Total - job.Progress.Sent - job.Progress.Failed
		s.jobsMutex.Unlock()
		
		// Add delay between messages (except for last message)
		if i < len(job.Contacts)-1 {
			delay := s.calculateDelay(job)
			s.log.Debug("Waiting %d seconds before next message", delay)
			
			select {
			case <-job.ctx.Done():
				job.Status = "cancelled"
				return
			case <-time.After(time.Duration(delay) * time.Second):
				// Continue to next message
			}
		}
	}
	
	// Mark job as completed
	job.Status = "completed"
	now = time.Now()
	job.CompletedAt = &now
	
	s.log.Info("Bulk messaging job %s completed. Sent: %d, Failed: %d", 
		job.ID, job.Progress.Sent, job.Progress.Failed)
}

// processMessage sends a single message
func (s *BulkMessagingService) processMessage(job *BulkMessageJob, contact models.Contact, index int) bool {
	// Generate personalized message content
	content, err := s.generateMessageContent(job.Template, contact, job.Variables)
	if err != nil {
		s.log.Error("Failed to generate message content for contact %d in job %s: %v", contact.ID, job.ID, err)
		return false
	}
	
	// Create message request
	messageReq := &models.SendMessageRequest{
		To:      contact.Phone,
		Message: content,
	}
	
	// Send message
	_, err = s.whatsappService.SendMessage(job.SessionID, messageReq)
	if err != nil {
		s.log.Error("Failed to send message to %s in job %s: %v", contact.Phone, job.ID, err)
		return false
	}
	
	s.log.Debug("Sent message %d/%d to %s (%s) in job %s", 
		index+1, len(job.Contacts), contact.Phone, contact.Name, job.ID)
	
	return true
}

// generateMessageContent creates personalized message content
func (s *BulkMessagingService) generateMessageContent(template *models.MessageTemplate, contact models.Contact, globalVars map[string]string) (string, error) {
	content := template.Content
	
	// Default contact variables
	contactVars := map[string]string{
		"{{name}}":     contact.Name,
		"{{phone}}":    contact.Phone,
		"{{email}}":    contact.Email,
		"{{company}}":  contact.Company,
		"{{position}}": contact.Position,
	}
	
	// Add global variables
	for key, value := range globalVars {
		if !strings.HasPrefix(key, "{{") {
			key = "{{" + key + "}}"
		}
		contactVars[key] = value
	}
	
	// Add template-specific variables
	if template.Variables != nil {
		for _, variable := range template.Variables {
			placeholder := variable.Placeholder
			if !strings.HasPrefix(placeholder, "{{") {
				placeholder = "{{" + placeholder + "}}"
			}
			
			// Use global var if exists, otherwise use default
			if value, exists := globalVars[variable.Name]; exists {
				contactVars[placeholder] = value
			} else if variable.DefaultValue != "" {
				contactVars[placeholder] = variable.DefaultValue
			}
		}
	}
	
	// Replace variables in content
	for placeholder, value := range contactVars {
		content = strings.ReplaceAll(content, placeholder, value)
	}
	
	return content, nil
}

// calculateDelay returns delay in seconds, with optional randomization
func (s *BulkMessagingService) calculateDelay(job *BulkMessageJob) int {
	baseDelay := job.DelayBetween
	if baseDelay <= 0 {
		baseDelay = 1 // Minimum 1 second delay
	}
	
	if !job.RandomDelay {
		return baseDelay
	}
	
	// Add random variation (±30% of base delay)
	variation := int(float64(baseDelay) * 0.3)
	if variation == 0 {
		variation = 1
	}
	
	randomVariation := rand.Intn(2*variation+1) - variation // Range: -variation to +variation
	finalDelay := baseDelay + randomVariation
	
	if finalDelay < 1 {
		finalDelay = 1
	}
	
	return finalDelay
}

// GetJob returns a job by ID
func (s *BulkMessagingService) GetJob(jobID string) (*BulkMessageJob, error) {
	s.jobsMutex.RLock()
	defer s.jobsMutex.RUnlock()
	
	job, exists := s.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found")
	}
	
	return job, nil
}

// GetJobs returns all jobs
func (s *BulkMessagingService) GetJobs() []*BulkMessageJob {
	s.jobsMutex.RLock()
	defer s.jobsMutex.RUnlock()
	
	jobs := make([]*BulkMessageJob, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}
	
	return jobs
}

// PauseJob pauses a running job
func (s *BulkMessagingService) PauseJob(jobID string) error {
	s.jobsMutex.Lock()
	defer s.jobsMutex.Unlock()
	
	job, exists := s.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found")
	}
	
	if job.Status != "running" {
		return fmt.Errorf("job is not running")
	}
	
	job.cancel()
	job.Status = "paused"
	
	s.log.Info("Paused bulk messaging job %s", jobID)
	return nil
}

// CancelJob cancels a job
func (s *BulkMessagingService) CancelJob(jobID string) error {
	s.jobsMutex.Lock()
	defer s.jobsMutex.Unlock()
	
	job, exists := s.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found")
	}
	
	if job.Status == "completed" || job.Status == "cancelled" {
		return fmt.Errorf("job is already %s", job.Status)
	}
	
	job.cancel()
	job.Status = "cancelled"
	
	s.log.Info("Cancelled bulk messaging job %s", jobID)
	return nil
}

// DeleteJob removes a completed job
func (s *BulkMessagingService) DeleteJob(jobID string) error {
	s.jobsMutex.Lock()
	defer s.jobsMutex.Unlock()
	
	job, exists := s.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found")
	}
	
	if job.Status == "running" {
		return fmt.Errorf("cannot delete running job")
	}
	
	if job.cancel != nil {
		job.cancel()
	}
	
	delete(s.jobs, jobID)
	
	s.log.Info("Deleted bulk messaging job %s", jobID)
	return nil
}

// CleanupOldJobs removes old completed jobs
func (s *BulkMessagingService) CleanupOldJobs(olderThan time.Duration) int {
	s.jobsMutex.Lock()
	defer s.jobsMutex.Unlock()
	
	cutoff := time.Now().Add(-olderThan)
	cleaned := 0
	
	for jobID, job := range s.jobs {
		if (job.Status == "completed" || job.Status == "cancelled" || job.Status == "failed") &&
			job.CreatedAt.Before(cutoff) {
			
			if job.cancel != nil {
				job.cancel()
			}
			delete(s.jobs, jobID)
			cleaned++
		}
	}
	
	if cleaned > 0 {
		s.log.Info("Cleaned up %d old bulk messaging jobs", cleaned)
	}
	
	return cleaned
}

// GetJobStats returns statistics for all jobs
func (s *BulkMessagingService) GetJobStats() map[string]interface{} {
	s.jobsMutex.RLock()
	defer s.jobsMutex.RUnlock()
	
	stats := map[string]interface{}{
		"total_jobs":     len(s.jobs),
		"running_jobs":   0,
		"completed_jobs": 0,
		"failed_jobs":    0,
		"cancelled_jobs": 0,
		"total_messages": 0,
		"sent_messages":  0,
		"failed_messages": 0,
	}
	
	for _, job := range s.jobs {
		switch job.Status {
		case "running":
			stats["running_jobs"] = stats["running_jobs"].(int) + 1
		case "completed":
			stats["completed_jobs"] = stats["completed_jobs"].(int) + 1
		case "failed":
			stats["failed_jobs"] = stats["failed_jobs"].(int) + 1
		case "cancelled":
			stats["cancelled_jobs"] = stats["cancelled_jobs"].(int) + 1
		}
		
		stats["total_messages"] = stats["total_messages"].(int) + job.Progress.Total
		stats["sent_messages"] = stats["sent_messages"].(int) + job.Progress.Sent
		stats["failed_messages"] = stats["failed_messages"].(int) + job.Progress.Failed
	}
	
	return stats
}

// generateJobID creates a unique job ID
func (s *BulkMessagingService) generateJobID() string {
	return fmt.Sprintf("job_%d_%d", time.Now().Unix(), rand.Intn(10000))
}

// EstimateJobDuration estimates how long a job will take
func (s *BulkMessagingService) EstimateJobDuration(contactCount, delayBetween int, randomDelay bool) time.Duration {
	avgDelay := delayBetween
	if randomDelay {
		// Add 15% for random variation average
		avgDelay = int(float64(delayBetween) * 1.15)
	}
	
	// Total time = (contacts - 1) * delay + estimated message sending time
	totalSeconds := (contactCount - 1) * avgDelay + contactCount * 2 // 2 seconds per message
	
	return time.Duration(totalSeconds) * time.Second
}

// JobSummary returns a simplified job summary for API responses
type JobSummary struct {
	ID          string              `json:"id"`
	CampaignID  *int                `json:"campaign_id,omitempty"`
	SessionID   string              `json:"session_id"`
	Status      string              `json:"status"`
	Progress    BulkMessageProgress `json:"progress"`
	CreatedAt   time.Time           `json:"created_at"`
	StartedAt   *time.Time          `json:"started_at,omitempty"`
	CompletedAt *time.Time          `json:"completed_at,omitempty"`
}

// GetJobSummary returns a simplified job summary
func (s *BulkMessagingService) GetJobSummary(jobID string) (*JobSummary, error) {
	job, err := s.GetJob(jobID)
	if err != nil {
		return nil, err
	}
	
	return &JobSummary{
		ID:          job.ID,
		CampaignID:  job.CampaignID,
		SessionID:   job.SessionID,
		Status:      job.Status,
		Progress:    job.Progress,
		CreatedAt:   job.CreatedAt,
		StartedAt:   job.StartedAt,
		CompletedAt: job.CompletedAt,
	}, nil
}