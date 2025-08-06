package services

import (
	"whatsapp-multi-session/internal/repository"
	"whatsapp-multi-session/pkg/logger"
)

type AnalyticsService struct {
	analyticsRepo *repository.AnalyticsRepository
	userRepo      *repository.UserRepository
	log           *logger.Logger
}

func NewAnalyticsService(analyticsRepo *repository.AnalyticsRepository, userRepo *repository.UserRepository, log *logger.Logger) *AnalyticsService {
	return &AnalyticsService{
		analyticsRepo: analyticsRepo,
		userRepo:      userRepo,
		log:           log,
	}
}

// AnalyticsData represents the complete analytics data
type AnalyticsData struct {
	MessageStats  *repository.MessageStats         `json:"message_stats"`
	SessionStats  *repository.SessionStats         `json:"session_stats"`
	UserStats     *repository.UserStats            `json:"user_stats,omitempty"`
	MessageTrend  []repository.TimeSeriesData      `json:"message_trend"`
	TopContacts   []map[string]interface{}         `json:"top_contacts"`
	SessionActivity []map[string]interface{}       `json:"session_activity"`
}

// GetAnalytics retrieves comprehensive analytics data
func (s *AnalyticsService) GetAnalytics(userId int64, isAdmin bool, timeRange string) (*AnalyticsData, error) {
	s.log.Info("Fetching analytics for user %d (admin: %v) with time range: %s", userId, isAdmin, timeRange)
	
	// If not admin, use the userId for filtering
	filterUserId := userId
	if isAdmin {
		filterUserId = 0 // 0 means no filtering (get all data)
	}
	
	// Get message statistics
	messageStats, err := s.analyticsRepo.GetMessageStats(filterUserId, timeRange)
	if err != nil {
		s.log.Error("Failed to get message stats: %v", err)
		return nil, err
	}
	
	// Get session statistics
	sessionStats, err := s.analyticsRepo.GetSessionStats(filterUserId)
	if err != nil {
		s.log.Error("Failed to get session stats: %v", err)
		return nil, err
	}
	
	// Get user statistics (admin only)
	var userStats *repository.UserStats
	if isAdmin {
		userStats, err = s.analyticsRepo.GetUserStats()
		if err != nil {
			s.log.Error("Failed to get user stats: %v", err)
			// Don't fail the whole request if user stats fail
		}
	}
	
	// Determine interval based on time range
	interval := "day"
	switch timeRange {
	case "today":
		interval = "hour"
	case "week":
		interval = "day"
	case "month":
		interval = "day"
	case "year":
		interval = "month"
	}
	
	// Get message time series
	messageTrend, err := s.analyticsRepo.GetMessageTimeSeries(filterUserId, timeRange, interval)
	if err != nil {
		s.log.Error("Failed to get message trend: %v", err)
		messageTrend = []repository.TimeSeriesData{} // Return empty array instead of failing
	}
	
	// Get top contacts
	topContacts, err := s.analyticsRepo.GetTopContacts(filterUserId, 10)
	if err != nil {
		s.log.Error("Failed to get top contacts: %v", err)
		topContacts = []map[string]interface{}{} // Return empty array instead of failing
	}
	
	// Get session activity
	sessionActivity, err := s.analyticsRepo.GetSessionActivity(filterUserId, 10)
	if err != nil {
		s.log.Error("Failed to get session activity: %v", err)
		sessionActivity = []map[string]interface{}{} // Return empty array instead of failing
	}
	
	return &AnalyticsData{
		MessageStats:    messageStats,
		SessionStats:    sessionStats,
		UserStats:       userStats,
		MessageTrend:    messageTrend,
		TopContacts:     topContacts,
		SessionActivity: sessionActivity,
	}, nil
}

// GetMessageStats retrieves only message statistics
func (s *AnalyticsService) GetMessageStats(userId int64, isAdmin bool, timeRange string) (*repository.MessageStats, error) {
	filterUserId := userId
	if isAdmin {
		filterUserId = 0
	}
	
	return s.analyticsRepo.GetMessageStats(filterUserId, timeRange)
}

// GetSessionStats retrieves only session statistics
func (s *AnalyticsService) GetSessionStats(userId int64, isAdmin bool) (*repository.SessionStats, error) {
	filterUserId := userId
	if isAdmin {
		filterUserId = 0
	}
	
	return s.analyticsRepo.GetSessionStats(filterUserId)
}