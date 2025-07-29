package logger

import (
	"fmt"
	"log"
	"os"
)

// LogWriter interface for writing logs to different destinations
type LogWriter interface {
	WriteLog(level, message, component, sessionID string, userID *int64, metadata map[string]any) error
}

// Logger provides configurable logging
type Logger struct {
	enabled   bool
	level     string
	logger    *log.Logger
	writers   []LogWriter
	component string
	sessionID string
	userID    *int64
}

// LogLevel constants
const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

// New creates a new logger instance
func New(enabled bool, level string) *Logger {
	return &Logger{
		enabled: enabled,
		level:   level,
		logger:  log.New(os.Stdout, "", log.LstdFlags),
		writers: make([]LogWriter, 0),
	}
}

// AddWriter adds a log writer to the logger
func (l *Logger) AddWriter(writer LogWriter) {
	l.writers = append(l.writers, writer)
}

// WithContext creates a new logger with context
func (l *Logger) WithContext(component, sessionID string, userID *int64) *Logger {
	return &Logger{
		enabled:   l.enabled,
		level:     l.level,
		logger:    l.logger,
		writers:   l.writers,
		component: component,
		sessionID: sessionID,
		userID:    userID,
	}
}

// writeToWriters writes log entry to all registered writers
func (l *Logger) writeToWriters(level, message string, metadata map[string]any) {
	for _, writer := range l.writers {
		if err := writer.WriteLog(level, message, l.component, l.sessionID, l.userID, metadata); err != nil {
			// Log the error to stderr to avoid infinite loops
			fmt.Fprintf(os.Stderr, "Failed to write log to writer: %v\n", err)
		}
	}
}

// shouldLog checks if a message should be logged based on level
func (l *Logger) shouldLog(level string) bool {
	if !l.enabled {
		return false
	}

	levels := map[string]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
	}

	currentLevel, ok := levels[l.level]
	if !ok {
		currentLevel = 1 // default to info
	}

	msgLevel, ok := levels[level]
	if !ok {
		msgLevel = 1 // default to info
	}

	return msgLevel >= currentLevel
}

// Debug logs debug messages
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.shouldLog(LevelDebug) {
		message := fmt.Sprintf(format, v...)
		l.logger.Printf("[DEBUG] "+message)
		l.writeToWriters(LevelDebug, message, nil)
	}
}

// DebugWithMetadata logs debug messages with metadata
func (l *Logger) DebugWithMetadata(metadata map[string]any, format string, v ...interface{}) {
	if l.shouldLog(LevelDebug) {
		message := fmt.Sprintf(format, v...)
		l.logger.Printf("[DEBUG] "+message)
		l.writeToWriters(LevelDebug, message, metadata)
	}
}

// Info logs info messages
func (l *Logger) Info(format string, v ...interface{}) {
	if l.shouldLog(LevelInfo) {
		message := fmt.Sprintf(format, v...)
		l.logger.Printf("[INFO] "+message)
		l.writeToWriters(LevelInfo, message, nil)
	}
}

// InfoWithMetadata logs info messages with metadata
func (l *Logger) InfoWithMetadata(metadata map[string]any, format string, v ...interface{}) {
	if l.shouldLog(LevelInfo) {
		message := fmt.Sprintf(format, v...)
		l.logger.Printf("[INFO] "+message)
		l.writeToWriters(LevelInfo, message, metadata)
	}
}

// Warn logs warning messages
func (l *Logger) Warn(format string, v ...interface{}) {
	if l.shouldLog(LevelWarn) {
		message := fmt.Sprintf(format, v...)
		l.logger.Printf("[WARN] "+message)
		l.writeToWriters(LevelWarn, message, nil)
	}
}

// WarnWithMetadata logs warning messages with metadata
func (l *Logger) WarnWithMetadata(metadata map[string]any, format string, v ...interface{}) {
	if l.shouldLog(LevelWarn) {
		message := fmt.Sprintf(format, v...)
		l.logger.Printf("[WARN] "+message)
		l.writeToWriters(LevelWarn, message, metadata)
	}
}

// Error logs error messages
func (l *Logger) Error(format string, v ...interface{}) {
	if l.shouldLog(LevelError) {
		message := fmt.Sprintf(format, v...)
		l.logger.Printf("[ERROR] "+message)
		l.writeToWriters(LevelError, message, nil)
	}
}

// ErrorWithMetadata logs error messages with metadata
func (l *Logger) ErrorWithMetadata(metadata map[string]any, format string, v ...interface{}) {
	if l.shouldLog(LevelError) {
		message := fmt.Sprintf(format, v...)
		l.logger.Printf("[ERROR] "+message)
		l.writeToWriters(LevelError, message, metadata)
	}
}

// Printf logs formatted messages (for compatibility)
func (l *Logger) Printf(format string, v ...interface{}) {
	l.Info(format, v...)
}

// Println logs messages (for compatibility)
func (l *Logger) Println(v ...interface{}) {
	if l.shouldLog(LevelInfo) {
		l.logger.Println("[INFO]", fmt.Sprint(v...))
	}
}

// Print logs messages without newline (for compatibility)
func (l *Logger) Print(v ...interface{}) {
	if l.shouldLog(LevelInfo) {
		l.logger.Print("[INFO] ", fmt.Sprint(v...))
	}
}

// Fatal logs fatal messages and exits
func (l *Logger) Fatal(v ...interface{}) {
	l.logger.Fatal(v...)
}

// Fatalf logs formatted fatal messages and exits
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.logger.Fatalf(format, v...)
}