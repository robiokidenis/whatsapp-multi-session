package config

import (
	"os"
	"strconv"
	"time"
	
	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Server configuration
	Port string

	// Database configuration
	DatabasePath   string
	WhatsAppDBPath string
	DatabaseType   string // "sqlite" or "mysql"
	MySQLHost      string
	MySQLPort      string
	MySQLUser      string
	MySQLPassword  string
	MySQLDatabase  string

	// JWT configuration
	JWTSecret     string
	JWTExpiration time.Duration

	// Admin credentials
	AdminUsername string
	AdminPassword string

	// Application settings
	EnableLogging  bool
	LogLevel       string
	MaxSessions    int
	SessionTimeout time.Duration

	// WhatsApp settings
	AutoConnect bool
	QRTimeout   time.Duration

	// Security settings
	CORSAllowedOrigins []string
	RateLimit          int

	// Webhook settings
	WebhookTimeout    time.Duration
	WebhookMaxRetries int
}

// Load loads configuration from environment variables
func Load() *Config {
	// Load .env.local first (for local development), then .env (fallback)
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load(".env")
	
	return &Config{
		// Server
		Port: getEnv("PORT", "8080"),

		// Database
		DatabasePath:   getEnv("DATABASE_PATH", "./database/session_metadata.db"),
		WhatsAppDBPath: getEnv("WHATSAPP_DB_PATH", "./database/sessions.db"),
		DatabaseType:   getEnv("DATABASE_TYPE", "sqlite"),
		MySQLHost:      getEnv("MYSQL_HOST", "localhost"),
		MySQLPort:      getEnv("MYSQL_PORT", "3306"),
		MySQLUser:      getEnv("MYSQL_USER", "root"),
		MySQLPassword:  getEnv("MYSQL_PASSWORD", ""),
		MySQLDatabase:  getEnv("MYSQL_DATABASE", "waGo"),

		// JWT
		JWTSecret:     getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		JWTExpiration: getDurationEnv("SESSION_TIMEOUT", 24*time.Hour),

		// Admin
		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),

		// Application
		EnableLogging:  getBoolEnv("ENABLE_LOGGING", true),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		MaxSessions:    getIntEnv("MAX_SESSIONS", 10),
		SessionTimeout: getDurationEnv("SESSION_TIMEOUT", 24*time.Hour),

		// WhatsApp
		AutoConnect: getBoolEnv("AUTO_CONNECT", true),
		QRTimeout:   getDurationEnv("QR_TIMEOUT", 30*time.Second),

		// Security
		CORSAllowedOrigins: getStringSliceEnv("CORS_ALLOWED_ORIGINS", []string{"*"}),
		RateLimit:          getIntEnv("RATE_LIMIT", 100),

		// Webhook
		WebhookTimeout:    getDurationEnv("WEBHOOK_TIMEOUT", 30*time.Second),
		WebhookMaxRetries: getIntEnv("WEBHOOK_MAX_RETRIES", 3),
	}
}

// Helper functions
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getBoolEnv(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		b, err := strconv.ParseBool(value)
		if err == nil {
			return b
		}
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		i, err := strconv.Atoi(value)
		if err == nil {
			return i
		}
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		d, err := time.ParseDuration(value)
		if err == nil {
			return d
		}
	}
	return fallback
}

func getStringSliceEnv(key string, fallback []string) []string {
	if value := os.Getenv(key); value != "" {
		if value == "*" {
			return []string{"*"}
		}
		// Simple comma-separated parsing
		return splitAndTrim(value, ",")
	}
	return fallback
}

func splitAndTrim(s, sep string) []string {
	parts := make([]string, 0)
	for _, part := range splitString(s, sep) {
		if trimmed := trimString(part); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	return stringSplit(s, sep)
}

func stringSplit(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i = start - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimString(s string) string {
	start := 0
	end := len(s)

	for start < end && isSpace(s[start]) {
		start++
	}

	for end > start && isSpace(s[end-1]) {
		end--
	}

	return s[start:end]
}

func isSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}
