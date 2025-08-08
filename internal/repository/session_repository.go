package repository

import (
	"database/sql"
	"fmt"
	"time"

	"whatsapp-multi-session/internal/models"
)

// SessionRepository handles session data persistence
type SessionRepository struct {
	db *sql.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// proxyConfigToDBFields converts ProxyConfig to individual database fields
func proxyConfigToDBFields(config *models.ProxyConfig) (bool, string, string, int, string, string) {
	if config == nil {
		return false, "", "", 0, "", ""
	}
	return config.Enabled, config.Type, config.Host, config.Port, config.Username, config.Password
}

// dbFieldsToProxyConfig converts database fields to ProxyConfig
func dbFieldsToProxyConfig(enabled bool, proxyType, host string, port int, username, password string) *models.ProxyConfig {
	if !enabled && proxyType == "" && host == "" && port == 0 {
		return nil // No proxy configured
	}
	return &models.ProxyConfig{
		Enabled:  enabled,
		Type:     proxyType,
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
	}
}

// Create creates a new session
func (r *SessionRepository) Create(session *models.SessionMetadata) error {
	query := `
		INSERT INTO session_metadata (id, phone, actual_phone, name, position, webhook_url, auto_reply_text, 
		                             proxy_enabled, proxy_type, proxy_host, proxy_port, proxy_username, proxy_password,
		                             user_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	// Convert proxy config to database fields
	proxyEnabled, proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword := proxyConfigToDBFields(session.ProxyConfig)
	
	_, err := r.db.Exec(query,
		session.ID,
		session.Phone,
		session.ActualPhone,
		session.Name,
		session.Position,
		session.WebhookURL,
		session.AutoReplyText,
		proxyEnabled,
		proxyType,
		proxyHost,
		proxyPort,
		proxyUsername,
		proxyPassword,
		session.UserID,
		session.CreatedAt.Unix(),
	)
	
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	
	return nil
}

// GetByID retrieves a session by ID
func (r *SessionRepository) GetByID(id string) (*models.SessionMetadata, error) {
	session := &models.SessionMetadata{}
	query := `
		SELECT id, phone, actual_phone, name, position, webhook_url, auto_reply_text,
		       proxy_enabled, proxy_type, proxy_host, proxy_port, proxy_username, proxy_password,
		       user_id, created_at
		FROM session_metadata
		WHERE id = ?
	`
	
	var createdAtUnix int64
	var autoReplyText sql.NullString
	var proxyEnabled bool
	var proxyType, proxyHost, proxyUsername, proxyPassword string
	var proxyPort int
	
	err := r.db.QueryRow(query, id).Scan(
		&session.ID,
		&session.Phone,
		&session.ActualPhone,
		&session.Name,
		&session.Position,
		&session.WebhookURL,
		&autoReplyText,
		&proxyEnabled,
		&proxyType,
		&proxyHost,
		&proxyPort,
		&proxyUsername,
		&proxyPassword,
		&session.UserID,
		&createdAtUnix,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %v", err)
	}
	
	session.CreatedAt = time.Unix(createdAtUnix, 0)
	
	// Handle nullable auto_reply_text
	if autoReplyText.Valid {
		session.AutoReplyText = &autoReplyText.String
	}
	
	// Convert proxy fields to ProxyConfig
	session.ProxyConfig = dbFieldsToProxyConfig(proxyEnabled, proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword)
	
	return session, nil
}

// GetAll retrieves all sessions
func (r *SessionRepository) GetAll() ([]*models.SessionMetadata, error) {
	query := `
		SELECT id, phone, actual_phone, name, position, webhook_url, auto_reply_text, user_id, created_at
		FROM session_metadata
		ORDER BY position ASC, created_at DESC
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %v", err)
	}
	defer rows.Close()
	
	var sessions []*models.SessionMetadata
	for rows.Next() {
		session := &models.SessionMetadata{}
		
		var createdAtUnix int64
		var autoReplyText sql.NullString
		err := rows.Scan(
			&session.ID,
			&session.Phone,
			&session.ActualPhone,
			&session.Name,
			&session.Position,
			&session.WebhookURL,
			&autoReplyText,
			&session.UserID,
			&createdAtUnix,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %v", err)
		}
		
		session.CreatedAt = time.Unix(createdAtUnix, 0)
		
		// Handle nullable auto_reply_text
		if autoReplyText.Valid {
			session.AutoReplyText = &autoReplyText.String
		}
		
		sessions = append(sessions, session)
	}
	
	return sessions, nil
}

// Update updates a session
func (r *SessionRepository) Update(session *models.SessionMetadata) error {
	query := `
		UPDATE session_metadata
		SET phone = ?, actual_phone = ?, name = ?, position = ?, webhook_url = ?, auto_reply_text = ?
		WHERE id = ? AND user_id = ?
	`
	
	_, err := r.db.Exec(query,
		session.Phone,
		session.ActualPhone,
		session.Name,
		session.Position,
		session.WebhookURL,
		session.AutoReplyText,
		session.ID,
		session.UserID,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update session: %v", err)
	}
	
	return nil
}

// UpdateActualPhone updates the actual phone number after login
func (r *SessionRepository) UpdateActualPhone(id, actualPhone string) error {
	query := `UPDATE session_metadata SET actual_phone = ? WHERE id = ?`
	
	_, err := r.db.Exec(query, actualPhone, id)
	if err != nil {
		return fmt.Errorf("failed to update actual phone: %v", err)
	}
	
	return nil
}

// UpdateWebhook updates the webhook URL
func (r *SessionRepository) UpdateWebhook(id, webhookURL string) error {
	query := `UPDATE session_metadata SET webhook_url = ? WHERE id = ?`
	
	_, err := r.db.Exec(query, webhookURL, id)
	if err != nil {
		return fmt.Errorf("failed to update webhook: %v", err)
	}
	
	return nil
}

// UpdateAutoReplyText updates the auto reply text
func (r *SessionRepository) UpdateAutoReplyText(id string, autoReplyText *string) error {
	query := `UPDATE session_metadata SET auto_reply_text = ? WHERE id = ?`
	
	_, err := r.db.Exec(query, autoReplyText, id)
	if err != nil {
		return fmt.Errorf("failed to update auto reply text: %v", err)
	}
	
	return nil
}

// Delete deletes a session
func (r *SessionRepository) Delete(id string) error {
	query := `DELETE FROM session_metadata WHERE id = ?`
	
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %v", err)
	}
	
	return nil
}

// Count returns the total number of sessions
func (r *SessionRepository) Count() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM session_metadata`
	
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count sessions: %v", err)
	}
	
	return count, nil
}

// GetNextPosition returns the next available position
func (r *SessionRepository) GetNextPosition() (int, error) {
	var maxPosition sql.NullInt64
	query := `SELECT MAX(position) FROM session_metadata`
	
	err := r.db.QueryRow(query).Scan(&maxPosition)
	if err != nil {
		return 0, fmt.Errorf("failed to get max position: %v", err)
	}
	
	if maxPosition.Valid {
		return int(maxPosition.Int64) + 1, nil
	}
	
	return 0, nil
}

// GetByUserID retrieves all sessions for a specific user
func (r *SessionRepository) GetByUserID(userID int) ([]*models.SessionMetadata, error) {
	query := `
		SELECT id, phone, actual_phone, name, position, webhook_url, auto_reply_text, user_id, created_at
		FROM session_metadata
		WHERE user_id = ?
		ORDER BY position ASC, created_at DESC
	`
	
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions for user %d: %v", userID, err)
	}
	defer rows.Close()
	
	var sessions []*models.SessionMetadata
	for rows.Next() {
		session := &models.SessionMetadata{}
		
		var createdAtUnix int64
		var autoReplyText sql.NullString
		err := rows.Scan(
			&session.ID,
			&session.Phone,
			&session.ActualPhone,
			&session.Name,
			&session.Position,
			&session.WebhookURL,
			&autoReplyText,
			&session.UserID,
			&createdAtUnix,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %v", err)
		}
		
		session.CreatedAt = time.Unix(createdAtUnix, 0)
		
		// Handle nullable auto_reply_text
		if autoReplyText.Valid {
			session.AutoReplyText = &autoReplyText.String
		}
		
		sessions = append(sessions, session)
	}
	
	return sessions, nil
}

// GetByIDAndUserID retrieves a session by ID and user ID
func (r *SessionRepository) GetByIDAndUserID(id string, userID int) (*models.SessionMetadata, error) {
	session := &models.SessionMetadata{}
	query := `
		SELECT id, phone, actual_phone, name, position, webhook_url, auto_reply_text, user_id, created_at
		FROM session_metadata
		WHERE id = ? AND user_id = ?
	`
	
	var createdAtUnix int64
	var autoReplyText sql.NullString
	err := r.db.QueryRow(query, id, userID).Scan(
		&session.ID,
		&session.Phone,
		&session.ActualPhone,
		&session.Name,
		&session.Position,
		&session.WebhookURL,
		&autoReplyText,
		&session.UserID,
		&createdAtUnix,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %v", err)
	}
	
	session.CreatedAt = time.Unix(createdAtUnix, 0)
	
	// Handle nullable auto_reply_text
	if autoReplyText.Valid {
		session.AutoReplyText = &autoReplyText.String
	}
	
	return session, nil
}

// DeleteByIDAndUserID deletes a session by ID and user ID
func (r *SessionRepository) DeleteByIDAndUserID(id string, userID int) error {
	query := `DELETE FROM session_metadata WHERE id = ? AND user_id = ?`
	
	result, err := r.db.Exec(query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("session not found or not owned by user")
	}
	
	return nil
}

// CountByUserID returns the total number of sessions for a user
func (r *SessionRepository) CountByUserID(userID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM session_metadata WHERE user_id = ?`
	
	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count sessions for user %d: %v", userID, err)
	}
	
	return count, nil
}

// ReorderPositions updates positions for drag-and-drop reordering
func (r *SessionRepository) ReorderPositions(updates map[string]int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()
	
	query := `UPDATE session_metadata SET position = ? WHERE id = ?`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()
	
	for id, position := range updates {
		_, err := stmt.Exec(position, id)
		if err != nil {
			return fmt.Errorf("failed to update position for %s: %v", id, err)
		}
	}
	
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	
	return nil
}