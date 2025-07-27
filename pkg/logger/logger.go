package logger

import (
	"fmt"
	"log"
	"os"
)

// Logger provides configurable logging
type Logger struct {
	enabled bool
	level   string
	logger  *log.Logger
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
		l.logger.Printf("[DEBUG] "+format, v...)
	}
}

// Info logs info messages
func (l *Logger) Info(format string, v ...interface{}) {
	if l.shouldLog(LevelInfo) {
		l.logger.Printf("[INFO] "+format, v...)
	}
}

// Warn logs warning messages
func (l *Logger) Warn(format string, v ...interface{}) {
	if l.shouldLog(LevelWarn) {
		l.logger.Printf("[WARN] "+format, v...)
	}
}

// Error logs error messages
func (l *Logger) Error(format string, v ...interface{}) {
	if l.shouldLog(LevelError) {
		l.logger.Printf("[ERROR] "+format, v...)
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