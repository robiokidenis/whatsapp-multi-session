package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// LogEntry represents a log entry in the database
type LogEntry struct {
	ID        int64             `json:"id"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	Component string            `json:"component,omitempty"`
	SessionID string            `json:"session_id,omitempty"`
	UserID    *int64            `json:"user_id,omitempty"`
	Metadata  map[string]any    `json:"metadata,omitempty"`
	CreatedAt int64             `json:"created_at"`
}

// LogFilter represents filtering options for log queries
type LogFilter struct {
	Level     string `json:"level,omitempty"`
	Component string `json:"component,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	UserID    *int64 `json:"user_id,omitempty"`
	StartTime int64  `json:"start_time,omitempty"`
	EndTime   int64  `json:"end_time,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

// LogRepository handles log database operations
type LogRepository struct {
	db *Database
}

// NewLogRepository creates a new log repository
func NewLogRepository(db *Database) *LogRepository {
	return &LogRepository{db: db}
}

// Save saves a log entry to the database
func (r *LogRepository) Save(entry *LogEntry) error {
	if entry.CreatedAt == 0 {
		entry.CreatedAt = time.Now().Unix()
	}

	var metadataStr *string
	if entry.Metadata != nil && len(entry.Metadata) > 0 {
		metadataBytes, err := json.Marshal(entry.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %v", err)
		}
		metadataValue := string(metadataBytes)
		metadataStr = &metadataValue
	}

	query := `
		INSERT INTO logs (level, message, component, session_id, user_id, metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.DB().Exec(query,
		entry.Level,
		entry.Message,
		entry.Component,
		entry.SessionID,
		entry.UserID,
		metadataStr,
		entry.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save log entry: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %v", err)
	}

	entry.ID = id
	return nil
}

// GetLogs retrieves logs based on filter criteria
func (r *LogRepository) GetLogs(filter LogFilter) ([]LogEntry, error) {
	query := `SELECT id, level, message, component, session_id, user_id, metadata, created_at FROM logs WHERE 1=1`
	args := []any{}

	if filter.Level != "" {
		query += " AND level = ?"
		args = append(args, filter.Level)
	}

	if filter.Component != "" {
		query += " AND component = ?"
		args = append(args, filter.Component)
	}

	if filter.SessionID != "" {
		query += " AND session_id = ?"
		args = append(args, filter.SessionID)
	}

	if filter.UserID != nil {
		query += " AND user_id = ?"
		args = append(args, *filter.UserID)
	}

	if filter.StartTime > 0 {
		query += " AND created_at >= ?"
		args = append(args, filter.StartTime)
	}

	if filter.EndTime > 0 {
		query += " AND created_at <= ?"
		args = append(args, filter.EndTime)
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := r.db.DB().Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query logs: %v", err)
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var entry LogEntry
		var component, sessionID sql.NullString
		var userID sql.NullInt64
		var metadataStr sql.NullString

		err := rows.Scan(
			&entry.ID,
			&entry.Level,
			&entry.Message,
			&component,
			&sessionID,
			&userID,
			&metadataStr,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log entry: %v", err)
		}

		if component.Valid {
			entry.Component = component.String
		}
		if sessionID.Valid {
			entry.SessionID = sessionID.String
		}
		if userID.Valid {
			entry.UserID = &userID.Int64
		}

		if metadataStr.Valid && metadataStr.String != "" {
			var metadata map[string]any
			if err := json.Unmarshal([]byte(metadataStr.String), &metadata); err == nil {
				entry.Metadata = metadata
			}
		}

		logs = append(logs, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over log rows: %v", err)
	}

	return logs, nil
}

// GetLogCount returns the total count of logs matching the filter
func (r *LogRepository) GetLogCount(filter LogFilter) (int64, error) {
	query := `SELECT COUNT(*) FROM logs WHERE 1=1`
	args := []any{}

	if filter.Level != "" {
		query += " AND level = ?"
		args = append(args, filter.Level)
	}

	if filter.Component != "" {
		query += " AND component = ?"
		args = append(args, filter.Component)
	}

	if filter.SessionID != "" {
		query += " AND session_id = ?"
		args = append(args, filter.SessionID)
	}

	if filter.UserID != nil {
		query += " AND user_id = ?"
		args = append(args, *filter.UserID)
	}

	if filter.StartTime > 0 {
		query += " AND created_at >= ?"
		args = append(args, filter.StartTime)
	}

	if filter.EndTime > 0 {
		query += " AND created_at <= ?"
		args = append(args, filter.EndTime)
	}

	var count int64
	err := r.db.DB().QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count logs: %v", err)
	}

	return count, nil
}

// DeleteOldLogs deletes logs older than the specified timestamp
func (r *LogRepository) DeleteOldLogs(olderThan int64) (int64, error) {
	query := `DELETE FROM logs WHERE created_at < ?`
	result, err := r.db.DB().Exec(query, olderThan)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old logs: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %v", err)
	}

	return rowsAffected, nil
}