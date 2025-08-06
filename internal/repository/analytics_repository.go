package repository

import (
	"database/sql"
	"time"
)

type AnalyticsRepository struct {
	db *sql.DB
}

func NewAnalyticsRepository(db *sql.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

// MessageStats represents message statistics
type MessageStats struct {
	TotalMessages     int64 `json:"total_messages"`
	SentMessages      int64 `json:"sent_messages"`
	ReceivedMessages  int64 `json:"received_messages"`
	FailedMessages    int64 `json:"failed_messages"`
	MediaMessages     int64 `json:"media_messages"`
	TextMessages      int64 `json:"text_messages"`
}

// SessionStats represents session statistics
type SessionStats struct {
	TotalSessions    int64 `json:"total_sessions"`
	ActiveSessions   int64 `json:"active_sessions"`
	InactiveSessions int64 `json:"inactive_sessions"`
}

// UserStats represents user statistics
type UserStats struct {
	TotalUsers    int64 `json:"total_users"`
	ActiveUsers   int64 `json:"active_users"`
	AdminUsers    int64 `json:"admin_users"`
	RegularUsers  int64 `json:"regular_users"`
}

// TimeSeriesData represents time series data point
type TimeSeriesData struct {
	Time  time.Time `json:"time"`
	Value int64     `json:"value"`
}

// GetMessageStats retrieves message statistics
func (r *AnalyticsRepository) GetMessageStats(userId int64, timeRange string) (*MessageStats, error) {
	stats := &MessageStats{}
	
	// Check if messages table exists
	var tableExists int
	checkTableQuery := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='messages'`
	err := r.db.QueryRow(checkTableQuery).Scan(&tableExists)
	if err != nil {
		return stats, nil // Return empty stats if can't check table
	}
	
	if tableExists == 0 {
		// Messages table doesn't exist, return empty stats
		return stats, nil
	}
	
	// Base query parts
	baseQuery := `
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN direction = 'sent' THEN 1 ELSE 0 END) as sent,
			SUM(CASE WHEN direction = 'received' THEN 1 ELSE 0 END) as received,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed,
			SUM(CASE WHEN message_type IN ('image', 'video', 'audio', 'document') THEN 1 ELSE 0 END) as media,
			SUM(CASE WHEN message_type = 'text' THEN 1 ELSE 0 END) as text
		FROM messages m
		JOIN session_metadata s ON m.session_id = s.id
		WHERE 1=1
	`
	
	args := []interface{}{}
	
	// Add user filter if not admin
	if userId > 0 {
		baseQuery += " AND s.user_id = ?"
		args = append(args, userId)
	}
	
	// Add time range filter (SQLite syntax)
	switch timeRange {
	case "today":
		baseQuery += " AND m.created_at >= date('now')"
	case "week":
		baseQuery += " AND m.created_at >= date('now', '-7 days')"
	case "month":
		baseQuery += " AND m.created_at >= date('now', '-30 days')"
	case "year":
		baseQuery += " AND m.created_at >= date('now', '-1 year')"
	}
	
	row := r.db.QueryRow(baseQuery, args...)
	err = row.Scan(
		&stats.TotalMessages,
		&stats.SentMessages,
		&stats.ReceivedMessages,
		&stats.FailedMessages,
		&stats.MediaMessages,
		&stats.TextMessages,
	)
	
	if err != nil && err != sql.ErrNoRows {
		return stats, nil // Return empty stats on error
	}
	
	return stats, nil
}

// GetSessionStats retrieves session statistics
func (r *AnalyticsRepository) GetSessionStats(userId int64) (*SessionStats, error) {
	stats := &SessionStats{}
	
	// Since we don't have is_connected in session_metadata, we'll count all sessions as total
	// and assume they're active (since this is a simplified version)
	baseQuery := `
		SELECT 
			COUNT(*) as total
		FROM session_metadata
		WHERE 1=1
	`
	
	args := []interface{}{}
	
	if userId > 0 {
		baseQuery += " AND user_id = ?"
		args = append(args, userId)
	}
	
	row := r.db.QueryRow(baseQuery, args...)
	err := row.Scan(&stats.TotalSessions)
	
	if err != nil && err != sql.ErrNoRows {
		return stats, nil // Return empty stats on error
	}
	
	// For now, assume all sessions are active since we don't have connection status
	stats.ActiveSessions = stats.TotalSessions
	stats.InactiveSessions = 0
	
	return stats, nil
}

// GetUserStats retrieves user statistics (admin only)
func (r *AnalyticsRepository) GetUserStats() (*UserStats, error) {
	stats := &UserStats{}
	
	query := `
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN is_active = 1 THEN 1 ELSE 0 END) as active,
			SUM(CASE WHEN role = 'admin' THEN 1 ELSE 0 END) as admins,
			SUM(CASE WHEN role = 'user' THEN 1 ELSE 0 END) as regular
		FROM users
	`
	
	row := r.db.QueryRow(query)
	err := row.Scan(
		&stats.TotalUsers,
		&stats.ActiveUsers,
		&stats.AdminUsers,
		&stats.RegularUsers,
	)
	
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	
	return stats, nil
}

// GetMessageTimeSeries retrieves message count time series data
func (r *AnalyticsRepository) GetMessageTimeSeries(userId int64, timeRange string, interval string) ([]TimeSeriesData, error) {
	// Initialize with empty slice to avoid null response
	data := make([]TimeSeriesData, 0)
	
	// Check if messages table exists
	var tableExists int
	checkTableQuery := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='messages'`
	err := r.db.QueryRow(checkTableQuery).Scan(&tableExists)
	if err != nil || tableExists == 0 {
		// Return empty data if table doesn't exist
		return data, nil
	}
	
	// Determine groupBy based on interval (SQLite syntax)
	var groupBy string
	switch interval {
	case "hour":
		groupBy = "strftime('%Y-%m-%d %H:00:00', m.created_at)"
	case "day":
		groupBy = "date(m.created_at)"
	case "week":
		groupBy = "strftime('%Y-%W', m.created_at)"
	case "month":
		groupBy = "strftime('%Y-%m', m.created_at)"
	default:
		groupBy = "date(m.created_at)"
	}
	
	query := `
		SELECT 
			` + groupBy + ` as time_period,
			COUNT(*) as count
		FROM messages m
		JOIN session_metadata s ON m.session_id = s.id
		WHERE 1=1
	`
	
	args := []interface{}{}
	
	if userId > 0 {
		query += " AND s.user_id = ?"
		args = append(args, userId)
	}
	
	// Add time range filter (SQLite syntax)
	switch timeRange {
	case "today":
		query += " AND m.created_at >= date('now')"
	case "week":
		query += " AND m.created_at >= date('now', '-7 days')"
	case "month":
		query += " AND m.created_at >= date('now', '-30 days')"
	case "year":
		query += " AND m.created_at >= date('now', '-1 year')"
	}
	
	query += " GROUP BY time_period ORDER BY time_period"
	
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return data, nil // Return empty data on error
	}
	defer rows.Close()
	
	for rows.Next() {
		var timeStr string
		var count int64
		
		if err := rows.Scan(&timeStr, &count); err != nil {
			continue // Skip invalid rows
		}
		
		// Parse time based on format
		var t time.Time
		switch interval {
		case "hour":
			t, _ = time.Parse("2006-01-02 15:04:05", timeStr)
		case "day":
			t, _ = time.Parse("2006-01-02", timeStr)
		default:
			t, _ = time.Parse("2006-01-02", timeStr)
		}
		
		data = append(data, TimeSeriesData{
			Time:  t,
			Value: count,
		})
	}
	
	return data, nil
}

// GetTopContacts retrieves top contacts by message count
func (r *AnalyticsRepository) GetTopContacts(userId int64, limit int) ([]map[string]interface{}, error) {
	// Initialize with empty slice to avoid null response
	contacts := make([]map[string]interface{}, 0)
	
	// Check if messages table exists
	var tableExists int
	checkTableQuery := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='messages'`
	err := r.db.QueryRow(checkTableQuery).Scan(&tableExists)
	if err != nil || tableExists == 0 {
		// Return empty data if table doesn't exist
		return contacts, nil
	}
	
	query := `
		SELECT 
			m.recipient_jid as contact,
			COUNT(*) as message_count,
			MAX(m.created_at) as last_message
		FROM messages m
		JOIN session_metadata s ON m.session_id = s.id
		WHERE 1=1
	`
	
	args := []interface{}{}
	
	if userId > 0 {
		query += " AND s.user_id = ?"
		args = append(args, userId)
	}
	
	query += ` 
		GROUP BY m.recipient_jid 
		ORDER BY message_count DESC 
		LIMIT ?
	`
	args = append(args, limit)
	
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return contacts, nil // Return empty data on error
	}
	defer rows.Close()
	
	for rows.Next() {
		var contact string
		var messageCount int64
		var lastMessage time.Time
		
		if err := rows.Scan(&contact, &messageCount, &lastMessage); err != nil {
			continue // Skip invalid rows
		}
		
		contacts = append(contacts, map[string]interface{}{
			"contact":       contact,
			"message_count": messageCount,
			"last_message":  lastMessage,
		})
	}
	
	return contacts, nil
}

// GetSessionActivity retrieves session activity data
func (r *AnalyticsRepository) GetSessionActivity(userId int64, limit int) ([]map[string]interface{}, error) {
	// Initialize with empty slice to avoid null response
	sessions := make([]map[string]interface{}, 0)
	
	// First check if messages table exists to determine query strategy
	var tableExists int
	checkTableQuery := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='messages'`
	err := r.db.QueryRow(checkTableQuery).Scan(&tableExists)
	
	var query string
	args := []interface{}{}
	
	if err != nil || tableExists == 0 {
		// Messages table doesn't exist, just return session info without message counts
		query = `
			SELECT 
				s.id,
				s.phone,
				s.name,
				0 as message_count
			FROM session_metadata s
			WHERE 1=1
		`
		
		if userId > 0 {
			query += " AND s.user_id = ?"
			args = append(args, userId)
		}
		
		query += ` ORDER BY s.created_at DESC LIMIT ?`
		args = append(args, limit)
	} else {
		// Messages table exists, include message counts
		query = `
			SELECT 
				s.id,
				s.phone,
				s.name,
				COUNT(m.id) as message_count,
				MAX(m.created_at) as last_activity
			FROM session_metadata s
			LEFT JOIN messages m ON s.id = m.session_id
			WHERE 1=1
		`
		
		if userId > 0 {
			query += " AND s.user_id = ?"
			args = append(args, userId)
		}
		
		query += ` 
			GROUP BY s.id, s.phone, s.name
			ORDER BY last_activity DESC 
			LIMIT ?
		`
		args = append(args, limit)
	}
	
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return sessions, nil // Return empty data on error
	}
	defer rows.Close()
	for rows.Next() {
		if tableExists == 0 {
			// Simple scan without message data
			var id, messageCount int64
			var phoneNumber, name sql.NullString
			
			if err := rows.Scan(&id, &phoneNumber, &name, &messageCount); err != nil {
				continue // Skip invalid rows
			}
			
			session := map[string]interface{}{
				"id":            id,
				"phone_number":  phoneNumber.String,
				"name":          name.String,
				"is_connected":  true, // Assume connected since we don't have this field
				"message_count": messageCount,
			}
			
			sessions = append(sessions, session)
		} else {
			// Full scan with message data
			var id, messageCount int64
			var phoneNumber, name sql.NullString
			var lastActivity sql.NullTime
			
			if err := rows.Scan(&id, &phoneNumber, &name, &messageCount, &lastActivity); err != nil {
				continue // Skip invalid rows
			}
			
			session := map[string]interface{}{
				"id":            id,
				"phone_number":  phoneNumber.String,
				"name":          name.String,
				"is_connected":  true, // Assume connected since we don't have this field
				"message_count": messageCount,
			}
			
			if lastActivity.Valid {
				session["last_activity"] = lastActivity.Time
			}
			
			sessions = append(sessions, session)
		}
	}
	
	return sessions, nil
}