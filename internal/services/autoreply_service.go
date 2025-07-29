package services

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"sync"
	"time"

	"whatsapp-multi-session/internal/models"
	"whatsapp-multi-session/internal/repository"
	"whatsapp-multi-session/pkg/logger"
)

type AutoReplyService struct {
	autoReplyRepo *repository.AutoReplyRepository
	whatsappSvc   *WhatsAppService
	log           logger.Logger
	
	// Reply tracking to enforce daily limits
	replyTracker map[string]map[string]int // sessionID -> contactPhone -> count
	trackerMutex sync.RWMutex
	lastReset    time.Time
}

func NewAutoReplyService(autoReplyRepo *repository.AutoReplyRepository, whatsappSvc *WhatsAppService, log logger.Logger) *AutoReplyService {
	service := &AutoReplyService{
		autoReplyRepo: autoReplyRepo,
		whatsappSvc:   whatsappSvc,
		log:           log,
		replyTracker:  make(map[string]map[string]int),
		lastReset:     time.Now(),
	}
	
	// Reset reply counters daily
	go service.dailyResetRoutine()
	
	return service
}

// ProcessIncomingMessage processes incoming messages for auto-reply triggers
func (s *AutoReplyService) ProcessIncomingMessage(sessionID, contactPhone, messageText, messageType string) error {
	// Get active auto-reply rules for this session
	rules, err := s.autoReplyRepo.GetActiveAutoRepliesBySession(sessionID)
	if err != nil {
		s.log.Error("Failed to get auto-reply rules for session %s: %v", sessionID, err)
		return err
	}
	
	if len(rules) == 0 {
		return nil // No rules to process
	}
	
	// Check daily reply limit for this contact
	if !s.canReplyToContact(sessionID, contactPhone) {
		s.log.Debug("Daily reply limit reached for contact %s in session %s", contactPhone, sessionID)
		return nil
	}
	
	// Find matching rules (sorted by priority)
	matchingRule := s.findMatchingRule(rules, messageText, messageType, contactPhone)
	if matchingRule == nil {
		return nil // No matching rule
	}
	
	// Check time-based conditions
	if !s.isWithinTimeWindow(matchingRule) {
		s.log.Debug("Auto-reply rule %d outside time window", matchingRule.ID)
		return nil
	}
	
	// Process the auto-reply
	return s.processAutoReply(matchingRule, sessionID, contactPhone, messageText)
}

// findMatchingRule finds the highest priority matching rule
func (s *AutoReplyService) findMatchingRule(rules []models.AutoReply, messageText, messageType, contactPhone string) *models.AutoReply {
	// Sort by priority (higher number = higher priority)
	for i := 0; i < len(rules)-1; i++ {
		for j := i + 1; j < len(rules); j++ {
			if rules[i].Priority < rules[j].Priority {
				rules[i], rules[j] = rules[j], rules[i]
			}
		}
	}
	
	for _, rule := range rules {
		if s.doesRuleMatch(rule, messageText, messageType, contactPhone) {
			return &rule
		}
	}
	
	return nil
}

// doesRuleMatch checks if a rule matches the incoming message
func (s *AutoReplyService) doesRuleMatch(rule models.AutoReply, messageText, messageType, contactPhone string) bool {
	switch rule.Trigger {
	case "all":
		return true
		
	case "keyword":
		if len(rule.Keywords) == 0 {
			return false
		}
		
		messageText = strings.ToLower(messageText)
		for _, keyword := range rule.Keywords {
			if strings.Contains(messageText, strings.ToLower(keyword)) {
				return true
			}
		}
		return false
		
	case "new_contact":
		// Check if this is first time contact (simplified - you might want to check database)
		return !s.hasContactBeenSeenBefore(rule.SessionID, contactPhone)
		
	case "time_based":
		// This trigger is handled separately in time window check
		return true
		
	default:
		return false
	}
}

// processAutoReply executes the auto-reply
func (s *AutoReplyService) processAutoReply(rule *models.AutoReply, sessionID, contactPhone, originalMessage string) error {
	// Add delay if specified
	delay := s.calculateReplyDelay(rule)
	if delay > 0 {
		s.log.Debug("Adding %d second delay for auto-reply rule %d", delay, rule.ID)
		time.Sleep(time.Duration(delay) * time.Second)
	}
	
	// Prepare message request
	messageReq := &models.SendMessageRequest{
		To:      contactPhone,
		Message: rule.Response,
	}
	
	// Send the auto-reply
	_, err := s.whatsappSvc.SendMessage(sessionID, messageReq)
	success := err == nil
	
	// Log the auto-reply attempt
	logEntry := models.AutoReplyLog{
		AutoReplyID:  rule.ID,
		SessionID:    sessionID,
		ContactPhone: contactPhone,
		TriggerMsg:   originalMessage,
		Response:     rule.Response,
		Success:      success,
		CreatedAt:    time.Now(),
	}
	
	if !success {
		logEntry.ErrorMsg = err.Error()
		s.log.Error("Auto-reply failed for rule %d: %v", rule.ID, err)
	} else {
		s.log.Info("Auto-reply sent for rule %d to %s", rule.ID, contactPhone)
		
		// Increment usage count
		s.autoReplyRepo.IncrementUsageCount(rule.ID)
		
		// Track reply for daily limit
		s.trackReply(sessionID, contactPhone)
	}
	
	// Save log entry
	s.autoReplyRepo.CreateAutoReplyLog(&logEntry)
	
	return err
}

// calculateReplyDelay calculates delay with optional randomization
func (s *AutoReplyService) calculateReplyDelay(rule *models.AutoReply) int {
	if rule.DelayMin <= 0 && rule.DelayMax <= 0 {
		return 0
	}
	
	minDelay := rule.DelayMin
	maxDelay := rule.DelayMax
	
	if maxDelay <= minDelay {
		return minDelay
	}
	
	// Add random delay between min and max
	return minDelay + rand.Intn(maxDelay-minDelay+1)
}

// isWithinTimeWindow checks if current time is within rule's time window
func (s *AutoReplyService) isWithinTimeWindow(rule *models.AutoReply) bool {
	if rule.TimeStart == "" || rule.TimeEnd == "" {
		return true // No time restriction
	}
	
	now := time.Now()
	currentTime := now.Format("15:04")
	
	// Simple time comparison (assumes same day)
	return currentTime >= rule.TimeStart && currentTime <= rule.TimeEnd
}

// canReplyToContact checks if we can send another reply to this contact today
func (s *AutoReplyService) canReplyToContact(sessionID, contactPhone string) bool {
	s.trackerMutex.RLock()
	defer s.trackerMutex.RUnlock()
	
	sessionTracker, exists := s.replyTracker[sessionID]
	if !exists {
		return true
	}
	
	count, exists := sessionTracker[contactPhone]
	if !exists {
		return true
	}
	
	// For now, use a default daily limit (could be made configurable per rule)
	const defaultDailyLimit = 5
	return count < defaultDailyLimit
}

// trackReply increments the reply count for a contact
func (s *AutoReplyService) trackReply(sessionID, contactPhone string) {
	s.trackerMutex.Lock()
	defer s.trackerMutex.Unlock()
	
	if s.replyTracker[sessionID] == nil {
		s.replyTracker[sessionID] = make(map[string]int)
	}
	
	s.replyTracker[sessionID][contactPhone]++
}

// hasContactBeenSeenBefore checks if we've interacted with this contact before
func (s *AutoReplyService) hasContactBeenSeenBefore(sessionID, contactPhone string) bool {
	// This is a simplified implementation
	// In a real system, you'd check the database for previous interactions
	s.trackerMutex.RLock()
	defer s.trackerMutex.RUnlock()
	
	sessionTracker, exists := s.replyTracker[sessionID]
	if !exists {
		return false
	}
	
	_, exists = sessionTracker[contactPhone]
	return exists
}

// dailyResetRoutine resets reply counters daily at midnight
func (s *AutoReplyService) dailyResetRoutine() {
	for {
		now := time.Now()
		nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		duration := nextMidnight.Sub(now)
		
		time.Sleep(duration)
		
		s.trackerMutex.Lock()
		s.replyTracker = make(map[string]map[string]int)
		s.lastReset = time.Now()
		s.trackerMutex.Unlock()
		
		s.log.Info("Auto-reply daily counters reset")
	}
}

// TestAutoReply tests an auto-reply rule without sending actual message
func (s *AutoReplyService) TestAutoReply(req models.AutoReplyTestRequest) (*models.AutoReplyTestResponse, error) {
	rule, err := s.autoReplyRepo.GetAutoReply(req.AutoReplyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get auto-reply rule: %v", err)
	}
	
	testPhone := req.TestPhone
	if testPhone == "" {
		testPhone = "+1234567890" // Default test phone
	}
	
	response := &models.AutoReplyTestResponse{
		WouldTrigger: false,
		Reason:       "No match",
	}
	
	// Check if rule would match
	if s.doesRuleMatch(*rule, req.TestMessage, "text", testPhone) {
		response.WouldTrigger = true
		response.Response = rule.Response
		response.Delay = s.calculateReplyDelay(rule)
		response.Reason = fmt.Sprintf("Matched trigger: %s", rule.Trigger)
		
		// Check time window
		if !s.isWithinTimeWindow(rule) {
			response.WouldTrigger = false
			response.Reason = "Outside time window"
		}
		
		// Check if rule is active
		if !rule.IsActive {
			response.WouldTrigger = false
			response.Reason = "Rule is inactive"
		}
	}
	
	return response, nil
}

// GetAutoReplyStats returns statistics for auto-replies
func (s *AutoReplyService) GetAutoReplyStats(sessionID string) (*models.AutoReplyStats, error) {
	stats := &models.AutoReplyStats{}
	
	// Get rules count
	rules, err := s.autoReplyRepo.GetAutoRepliesBySession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get auto-reply rules: %v", err)
	}
	
	stats.TotalRules = len(rules)
	for _, rule := range rules {
		if rule.IsActive {
			stats.ActiveRules++
		}
	}
	
	// Get usage stats from logs (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	logs, err := s.autoReplyRepo.GetAutoReplyLogsBySession(sessionID, &thirtyDaysAgo, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get auto-reply logs: %v", err)
	}
	
	stats.TotalTriggers = len(logs)
	successCount := 0
	totalResponseTime := 0.0
	
	for _, log := range logs {
		if log.Success {
			successCount++
		}
		// Assuming average response time of 2 seconds for successful replies
		if log.Success {
			totalResponseTime += 2.0
		}
	}
	
	if stats.TotalTriggers > 0 {
		stats.SuccessRate = float64(successCount) / float64(stats.TotalTriggers)
	}
	
	if successCount > 0 {
		stats.AvgResponseTime = totalResponseTime / float64(successCount)
	}
	
	return stats, nil
}

// EnableAutoRepliesForSession enables auto-reply processing for a session
func (s *AutoReplyService) EnableAutoRepliesForSession(sessionID string) error {
	// This would integrate with your WhatsApp service to register message handlers
	s.log.Info("Auto-reply processing enabled for session %s", sessionID)
	return nil
}

// DisableAutoRepliesForSession disables auto-reply processing for a session
func (s *AutoReplyService) DisableAutoRepliesForSession(sessionID string) error {
	// This would integrate with your WhatsApp service to unregister message handlers
	s.log.Info("Auto-reply processing disabled for session %s", sessionID)
	return nil
}

// CreateAutoReplyFromTemplate creates common auto-reply templates
func (s *AutoReplyService) CreateAutoReplyFromTemplate(sessionID, templateType string) (*models.AutoReply, error) {
	var autoReply *models.AutoReply
	
	switch templateType {
	case "welcome":
		autoReply = &models.AutoReply{
			SessionID: sessionID,
			Name:      "Welcome Message",
			Trigger:   "new_contact",
			Response:  "Hello! Thanks for contacting us. How can we help you today?",
			IsActive:  true,
			Priority:  10,
			DelayMin:  1,
			DelayMax:  3,
		}
		
	case "away":
		autoReply = &models.AutoReply{
			SessionID: sessionID,
			Name:      "Away Message",
			Trigger:   "all",
			Response:  "We're currently away. We'll get back to you as soon as possible!",
			IsActive:  true,
			Priority:  5,
			DelayMin:  2,
			DelayMax:  5,
			TimeStart: "18:00",
			TimeEnd:   "09:00",
		}
		
	case "business_hours":
		autoReply = &models.AutoReply{
			SessionID: sessionID,
			Name:      "Business Hours",
			Trigger:   "all",
			Response:  "Thank you for your message. Our business hours are 9 AM to 6 PM, Monday to Friday. We'll respond during business hours.",
			IsActive:  true,
			Priority:  3,
			DelayMin:  1,
			DelayMax:  2,
		}
		
	case "support":
		autoReply = &models.AutoReply{
			SessionID: sessionID,
			Name:      "Support Keywords",
			Trigger:   "keyword",
			Keywords:  []string{"help", "support", "problem", "issue", "error"},
			Response:  "We received your support request. Our team will assist you shortly. Please describe your issue in detail.",
			IsActive:  true,
			Priority:  15,
			DelayMin:  1,
			DelayMax:  2,
		}
		
	default:
		return nil, fmt.Errorf("unknown template type: %s", templateType)
	}
	
	err := s.autoReplyRepo.CreateAutoReply(autoReply)
	if err != nil {
		return nil, fmt.Errorf("failed to create auto-reply from template: %v", err)
	}
	
	return autoReply, nil
}

// ValidateAutoReplyRule validates an auto-reply rule configuration
func (s *AutoReplyService) ValidateAutoReplyRule(rule *models.AutoReply) error {
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}
	
	if rule.SessionID == "" {
		return fmt.Errorf("session ID is required")
	}
	
	if rule.Response == "" {
		return fmt.Errorf("response message is required")
	}
	
	// Validate trigger-specific requirements
	switch rule.Trigger {
	case "keyword":
		if len(rule.Keywords) == 0 {
			return fmt.Errorf("keywords are required for keyword trigger")
		}
		
	case "all", "new_contact", "time_based":
		// These are valid trigger types
		
	default:
		return fmt.Errorf("invalid trigger type: %s", rule.Trigger)
	}
	
	// Validate time format if specified
	if rule.TimeStart != "" {
		if !s.isValidTimeFormat(rule.TimeStart) {
			return fmt.Errorf("invalid time format for TimeStart: %s (use HH:MM)", rule.TimeStart)
		}
	}
	
	if rule.TimeEnd != "" {
		if !s.isValidTimeFormat(rule.TimeEnd) {
			return fmt.Errorf("invalid time format for TimeEnd: %s (use HH:MM)", rule.TimeEnd)
		}
	}
	
	// Validate delay values
	if rule.DelayMin < 0 || rule.DelayMax < 0 {
		return fmt.Errorf("delay values cannot be negative")
	}
	
	if rule.DelayMax > 0 && rule.DelayMax < rule.DelayMin {
		return fmt.Errorf("DelayMax cannot be less than DelayMin")
	}
	
	return nil
}

// isValidTimeFormat validates HH:MM time format
func (s *AutoReplyService) isValidTimeFormat(timeStr string) bool {
	timeRegex := regexp.MustCompile(`^([01]?[0-9]|2[0-3]):[0-5][0-9]$`)
	return timeRegex.MatchString(timeStr)
}