package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"whatsapp-multi-session/internal/models"
)

type AutoReplyRepository struct {
	db *sql.DB
}

func NewAutoReplyRepository(db *sql.DB) *AutoReplyRepository {
	return &AutoReplyRepository{db: db}
}

// CreateAutoReply creates a new auto-reply rule
func (r *AutoReplyRepository) CreateAutoReply(autoReply *models.AutoReply) error {
	keywordsJSON, _ := json.Marshal(autoReply.Keywords)
	conditionsJSON, _ := json.Marshal(autoReply.Conditions)
	
	query := `
		INSERT INTO auto_replies (session_id, name, trigger_type, keywords, response, media_url, media_type, 
		                         is_active, priority, delay_min, delay_max, max_replies, time_start, time_end, 
		                         conditions, usage_count, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	result, err := r.db.Exec(query,
		autoReply.SessionID,
		autoReply.Name,
		autoReply.Trigger,
		string(keywordsJSON),
		autoReply.Response,
		autoReply.MediaURL,
		autoReply.MediaType,
		autoReply.IsActive,
		autoReply.Priority,
		autoReply.DelayMin,
		autoReply.DelayMax,
		autoReply.MaxReplies,
		autoReply.TimeStart,
		autoReply.TimeEnd,
		string(conditionsJSON),
		0, // initial usage count
		time.Now().Unix(),
	)
	
	if err != nil {
		return fmt.Errorf("failed to create auto-reply: %v", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get auto-reply ID: %v", err)
	}
	
	autoReply.ID = int(id)
	autoReply.CreatedAt = time.Now()
	autoReply.UsageCount = 0
	
	return nil
}

// GetAutoReply retrieves an auto-reply rule by ID
func (r *AutoReplyRepository) GetAutoReply(id int) (*models.AutoReply, error) {
	autoReply := &models.AutoReply{}
	var keywordsJSON, conditionsJSON sql.NullString
	var updatedAt sql.NullInt64
	var createdAt int64
	
	query := `
		SELECT id, session_id, name, trigger_type, keywords, response, media_url, media_type,
		       is_active, priority, delay_min, delay_max, max_replies, time_start, time_end,
		       conditions, usage_count, created_at, updated_at
		FROM auto_replies
		WHERE id = ?`
	
	err := r.db.QueryRow(query, id).Scan(
		&autoReply.ID,
		&autoReply.SessionID,
		&autoReply.Name,
		&autoReply.Trigger,
		&keywordsJSON,
		&autoReply.Response,
		&autoReply.MediaURL,
		&autoReply.MediaType,
		&autoReply.IsActive,
		&autoReply.Priority,
		&autoReply.DelayMin,
		&autoReply.DelayMax,
		&autoReply.MaxReplies,
		&autoReply.TimeStart,
		&autoReply.TimeEnd,
		&conditionsJSON,
		&autoReply.UsageCount,
		&createdAt,
		&updatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("auto-reply not found")
		}
		return nil, fmt.Errorf("failed to get auto-reply: %v", err)
	}
	
	// Parse timestamps
	autoReply.CreatedAt = time.Unix(createdAt, 0)
	if updatedAt.Valid {
		t := time.Unix(updatedAt.Int64, 0)
		autoReply.UpdatedAt = &t
	}
	
	// Parse JSON fields
	if keywordsJSON.Valid && keywordsJSON.String != "" {
		json.Unmarshal([]byte(keywordsJSON.String), &autoReply.Keywords)
	}
	
	if conditionsJSON.Valid && conditionsJSON.String != "" {
		json.Unmarshal([]byte(conditionsJSON.String), &autoReply.Conditions)
	}
	
	return autoReply, nil
}

// GetAutoRepliesBySession retrieves all auto-reply rules for a session
func (r *AutoReplyRepository) GetAutoRepliesBySession(sessionID string) ([]models.AutoReply, error) {
	return r.getAutoRepliesWithFilter("session_id = ?", []interface{}{sessionID}, "")
}

// GetActiveAutoRepliesBySession retrieves active auto-reply rules for a session, ordered by priority
func (r *AutoReplyRepository) GetActiveAutoRepliesBySession(sessionID string) ([]models.AutoReply, error) {
	return r.getAutoRepliesWithFilter("session_id = ? AND is_active = ?", []interface{}{sessionID, true}, "ORDER BY priority DESC")
}

// getAutoRepliesWithFilter helper function for querying auto-replies with filters
func (r *AutoReplyRepository) getAutoRepliesWithFilter(whereClause string, args []interface{}, orderBy string) ([]models.AutoReply, error) {
	query := `
		SELECT id, session_id, name, trigger_type, keywords, response, media_url, media_type,
		       is_active, priority, delay_min, delay_max, max_replies, time_start, time_end,
		       conditions, usage_count, created_at, updated_at
		FROM auto_replies
		WHERE ` + whereClause
	
	if orderBy != "" {
		query += " " + orderBy
	} else {
		query += " ORDER BY created_at DESC"
	}
	
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query auto-replies: %v", err)
	}
	defer rows.Close()
	
	var autoReplies []models.AutoReply
	
	for rows.Next() {
		autoReply := models.AutoReply{}
		var keywordsJSON, conditionsJSON sql.NullString
		var updatedAt sql.NullInt64
		var createdAt int64
		
		err := rows.Scan(
			&autoReply.ID,
			&autoReply.SessionID,
			&autoReply.Name,
			&autoReply.Trigger,
			&keywordsJSON,
			&autoReply.Response,
			&autoReply.MediaURL,
			&autoReply.MediaType,
			&autoReply.IsActive,
			&autoReply.Priority,
			&autoReply.DelayMin,
			&autoReply.DelayMax,
			&autoReply.MaxReplies,
			&autoReply.TimeStart,
			&autoReply.TimeEnd,
			&conditionsJSON,
			&autoReply.UsageCount,
			&createdAt,
			&updatedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan auto-reply: %v", err)
		}
		
		// Parse timestamps
		autoReply.CreatedAt = time.Unix(createdAt, 0)
		if updatedAt.Valid {
			t := time.Unix(updatedAt.Int64, 0)
			autoReply.UpdatedAt = &t
		}
		
		// Parse JSON fields
		if keywordsJSON.Valid && keywordsJSON.String != "" {
			json.Unmarshal([]byte(keywordsJSON.String), &autoReply.Keywords)
		}
		
		if conditionsJSON.Valid && conditionsJSON.String != "" {
			json.Unmarshal([]byte(conditionsJSON.String), &autoReply.Conditions)
		}
		
		autoReplies = append(autoReplies, autoReply)
	}
	
	return autoReplies, nil
}

// UpdateAutoReply updates an existing auto-reply rule
func (r *AutoReplyRepository) UpdateAutoReply(id int, req models.UpdateAutoReplyRequest) error {
	setParts := []string{}
	args := []interface{}{}
	
	if req.Name != "" {
		setParts = append(setParts, "name = ?")
		args = append(args, req.Name)
	}
	
	if req.Trigger != "" {
		setParts = append(setParts, "trigger_type = ?")
		args = append(args, req.Trigger)
	}
	
	if req.Keywords != nil {
		keywordsJSON, _ := json.Marshal(req.Keywords)
		setParts = append(setParts, "keywords = ?")
		args = append(args, string(keywordsJSON))
	}
	
	if req.Response != "" {
		setParts = append(setParts, "response = ?")
		args = append(args, req.Response)
	}
	
	if req.MediaURL != "" {
		setParts = append(setParts, "media_url = ?")
		args = append(args, req.MediaURL)
	}
	
	if req.MediaType != "" {
		setParts = append(setParts, "media_type = ?")
		args = append(args, req.MediaType)
	}
	
	if req.IsActive != nil {
		setParts = append(setParts, "is_active = ?")
		args = append(args, *req.IsActive)
	}
	
	if req.Priority > 0 {
		setParts = append(setParts, "priority = ?")
		args = append(args, req.Priority)
	}
	
	if req.DelayMin >= 0 {
		setParts = append(setParts, "delay_min = ?")
		args = append(args, req.DelayMin)
	}
	
	if req.DelayMax >= 0 {
		setParts = append(setParts, "delay_max = ?")
		args = append(args, req.DelayMax)
	}
	
	if req.MaxReplies >= 0 {
		setParts = append(setParts, "max_replies = ?")
		args = append(args, req.MaxReplies)
	}
	
	if req.TimeStart != "" {
		setParts = append(setParts, "time_start = ?")
		args = append(args, req.TimeStart)
	}
	
	if req.TimeEnd != "" {
		setParts = append(setParts, "time_end = ?")
		args = append(args, req.TimeEnd)
	}
	
	if req.Conditions != nil {
		conditionsJSON, _ := json.Marshal(req.Conditions)
		setParts = append(setParts, "conditions = ?")
		args = append(args, string(conditionsJSON))
	}
	
	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}
	
	setParts = append(setParts, "updated_at = ?")
	args = append(args, time.Now().Unix())
	args = append(args, id)
	
	query := fmt.Sprintf("UPDATE auto_replies SET %s WHERE id = ?", strings.Join(setParts, ", "))
	
	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update auto-reply: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("auto-reply not found")
	}
	
	return nil
}

// DeleteAutoReply deletes an auto-reply rule
func (r *AutoReplyRepository) DeleteAutoReply(id int) error {
	query := "DELETE FROM auto_replies WHERE id = ?"
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete auto-reply: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("auto-reply not found")
	}
	
	return nil
}

// IncrementUsageCount increments the usage count for an auto-reply rule
func (r *AutoReplyRepository) IncrementUsageCount(id int) error {
	query := "UPDATE auto_replies SET usage_count = usage_count + 1, updated_at = ? WHERE id = ?"
	
	result, err := r.db.Exec(query, time.Now().Unix(), id)
	if err != nil {
		return fmt.Errorf("failed to increment usage count: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("auto-reply not found")
	}
	
	return nil
}

// CreateAutoReplyLog creates a new auto-reply log entry
func (r *AutoReplyRepository) CreateAutoReplyLog(log *models.AutoReplyLog) error {
	query := `
		INSERT INTO auto_reply_logs (auto_reply_id, session_id, contact_phone, trigger_msg, response, success, error_msg, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	
	result, err := r.db.Exec(query,
		log.AutoReplyID,
		log.SessionID,
		log.ContactPhone,
		log.TriggerMsg,
		log.Response,
		log.Success,
		log.ErrorMsg,
		time.Now().Unix(),
	)
	
	if err != nil {
		return fmt.Errorf("failed to create auto-reply log: %v", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get auto-reply log ID: %v", err)
	}
	
	log.ID = int(id)
	log.CreatedAt = time.Now()
	
	return nil
}

// GetAutoReplyLogs retrieves auto-reply logs with optional filtering
func (r *AutoReplyRepository) GetAutoReplyLogs(autoReplyID *int, sessionID string, startDate, endDate *time.Time, limit int) ([]models.AutoReplyLog, error) {
	query := `
		SELECT arl.id, arl.auto_reply_id, arl.session_id, arl.contact_phone, arl.trigger_msg, 
		       arl.response, arl.success, arl.error_msg, arl.created_at
		FROM auto_reply_logs arl
		WHERE 1=1`
	
	args := []interface{}{}
	
	if autoReplyID != nil {
		query += " AND arl.auto_reply_id = ?"
		args = append(args, *autoReplyID)
	}
	
	if sessionID != "" {
		query += " AND arl.session_id = ?"
		args = append(args, sessionID)
	}
	
	if startDate != nil {
		query += " AND arl.created_at >= ?"
		args = append(args, startDate.Unix())
	}
	
	if endDate != nil {
		query += " AND arl.created_at <= ?"
		args = append(args, endDate.Unix())
	}
	
	query += " ORDER BY arl.created_at DESC"
	
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query auto-reply logs: %v", err)
	}
	defer rows.Close()
	
	var logs []models.AutoReplyLog
	
	for rows.Next() {
		log := models.AutoReplyLog{}
		var createdAt int64
		
		err := rows.Scan(
			&log.ID,
			&log.AutoReplyID,
			&log.SessionID,
			&log.ContactPhone,
			&log.TriggerMsg,
			&log.Response,
			&log.Success,
			&log.ErrorMsg,
			&createdAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan auto-reply log: %v", err)
		}
		
		log.CreatedAt = time.Unix(createdAt, 0)
		logs = append(logs, log)
	}
	
	return logs, nil
}

// GetAutoReplyLogsBySession retrieves auto-reply logs for a specific session
func (r *AutoReplyRepository) GetAutoReplyLogsBySession(sessionID string, startDate, endDate *time.Time) ([]models.AutoReplyLog, error) {
	return r.GetAutoReplyLogs(nil, sessionID, startDate, endDate, 0)
}

// DeleteOldAutoReplyLogs deletes auto-reply logs older than specified duration
func (r *AutoReplyRepository) DeleteOldAutoReplyLogs(olderThan time.Duration) (int, error) {
	cutoff := time.Now().Add(-olderThan)
	
	query := "DELETE FROM auto_reply_logs WHERE created_at < ?"
	
	result, err := r.db.Exec(query, cutoff.Unix())
	if err != nil {
		return 0, fmt.Errorf("failed to delete old auto-reply logs: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %v", err)
	}
	
	return int(rowsAffected), nil
}

// GetAutoReplyStatsBySession returns statistics for auto-replies in a session
func (r *AutoReplyRepository) GetAutoReplyStatsBySession(sessionID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Total rules
	var totalRules, activeRules int
	err := r.db.QueryRow("SELECT COUNT(*), SUM(CASE WHEN is_active THEN 1 ELSE 0 END) FROM auto_replies WHERE session_id = ?", sessionID).Scan(&totalRules, &activeRules)
	if err != nil {
		return nil, fmt.Errorf("failed to get rule counts: %v", err)
	}
	
	stats["total_rules"] = totalRules
	stats["active_rules"] = activeRules
	
	// Usage stats (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30).Unix()
	
	var totalTriggers, successfulTriggers int
	err = r.db.QueryRow(`
		SELECT COUNT(*), SUM(CASE WHEN success THEN 1 ELSE 0 END) 
		FROM auto_reply_logs 
		WHERE session_id = ? AND created_at >= ?`, sessionID, thirtyDaysAgo).Scan(&totalTriggers, &successfulTriggers)
	if err != nil {
		return nil, fmt.Errorf("failed to get trigger counts: %v", err)
	}
	
	stats["total_triggers"] = totalTriggers
	stats["successful_triggers"] = successfulTriggers
	
	if totalTriggers > 0 {
		stats["success_rate"] = float64(successfulTriggers) / float64(totalTriggers)
	} else {
		stats["success_rate"] = 0.0
	}
	
	// Most used rules
	rows, err := r.db.Query(`
		SELECT ar.id, ar.name, ar.usage_count
		FROM auto_replies ar
		WHERE ar.session_id = ?
		ORDER BY ar.usage_count DESC
		LIMIT 5`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query most used rules: %v", err)
	}
	defer rows.Close()
	
	type RuleUsage struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		UsageCount int    `json:"usage_count"`
	}
	
	var mostUsed []RuleUsage
	for rows.Next() {
		var usage RuleUsage
		err := rows.Scan(&usage.ID, &usage.Name, &usage.UsageCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan most used rule: %v", err)
		}
		mostUsed = append(mostUsed, usage)
	}
	stats["most_used_rules"] = mostUsed
	
	return stats, nil
}