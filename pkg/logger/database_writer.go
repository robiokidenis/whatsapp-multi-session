package logger

import (
	"time"
	"whatsapp-multi-session/internal/repository"
)

// DatabaseWriter implements LogWriter interface for database logging
type DatabaseWriter struct {
	repo *repository.LogRepository
}

// NewDatabaseWriter creates a new database log writer
func NewDatabaseWriter(repo *repository.LogRepository) *DatabaseWriter {
	return &DatabaseWriter{
		repo: repo,
	}
}

// WriteLog writes a log entry to the database
func (w *DatabaseWriter) WriteLog(level, message, component, sessionID string, userID *int64, metadata map[string]any) error {
	entry := &repository.LogEntry{
		Level:     level,
		Message:   message,
		Component: component,
		SessionID: sessionID,
		UserID:    userID,
		Metadata:  metadata,
		CreatedAt: time.Now().Unix(),
	}

	return w.repo.Save(entry)
}