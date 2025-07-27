package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/cors"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCompanionReg"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/proto"
)

// LoginAttempt tracks login attempts for rate limiting
type LoginAttempt struct {
	Count        int       `json:"count"`
	LastAttempt  time.Time `json:"last_attempt"`
	BlockedUntil time.Time `json:"blocked_until"`
}

// LoginRateLimiter manages login rate limiting
type LoginRateLimiter struct {
	attempts map[string]*LoginAttempt
	mu       sync.RWMutex

	// Configuration
	MaxAttempts     int           // Max attempts before blocking
	BlockDuration   time.Duration // How long to block after max attempts
	WindowDuration  time.Duration // Time window to count attempts
	CleanupInterval time.Duration // How often to clean up old attempts
}

// NewLoginRateLimiter creates a new rate limiter
func NewLoginRateLimiter() *LoginRateLimiter {
	limiter := &LoginRateLimiter{
		attempts:        make(map[string]*LoginAttempt),
		MaxAttempts:     5,                // 5 attempts
		BlockDuration:   15 * time.Minute, // Block for 15 minutes
		WindowDuration:  5 * time.Minute,  // 5 minute window
		CleanupInterval: 30 * time.Minute, // Cleanup every 30 minutes
	}

	// Start cleanup routine
	go limiter.cleanup()

	return limiter
}

// IsBlocked checks if an IP is currently blocked
func (l *LoginRateLimiter) IsBlocked(ip string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	attempt, exists := l.attempts[ip]
	if !exists {
		return false
	}

	return time.Now().Before(attempt.BlockedUntil)
}

// RecordAttempt records a login attempt for an IP
func (l *LoginRateLimiter) RecordAttempt(ip string, success bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	attempt, exists := l.attempts[ip]

	if !exists {
		attempt = &LoginAttempt{}
		l.attempts[ip] = attempt
	}

	if success {
		// Reset on successful login
		attempt.Count = 0
		attempt.BlockedUntil = time.Time{}
		return
	}

	// Check if we're in a new window
	if now.Sub(attempt.LastAttempt) > l.WindowDuration {
		attempt.Count = 1
	} else {
		attempt.Count++
	}

	attempt.LastAttempt = now

	// Block if max attempts reached
	if attempt.Count >= l.MaxAttempts {
		attempt.BlockedUntil = now.Add(l.BlockDuration)
		if logger != nil {
			logger.Warn("IP %s blocked for %v after %d failed attempts", ip, l.BlockDuration, attempt.Count)
		}
	}
}

// GetRemainingTime returns how long until the IP is unblocked
func (l *LoginRateLimiter) GetRemainingTime(ip string) time.Duration {
	l.mu.RLock()
	defer l.mu.RUnlock()

	attempt, exists := l.attempts[ip]
	if !exists {
		return 0
	}

	remaining := attempt.BlockedUntil.Sub(time.Now())
	if remaining < 0 {
		return 0
	}

	return remaining
}

// cleanup removes old entries periodically
func (l *LoginRateLimiter) cleanup() {
	ticker := time.NewTicker(l.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		now := time.Now()

		for ip, attempt := range l.attempts {
			// Remove if not blocked and last attempt was long ago
			if now.After(attempt.BlockedUntil) && now.Sub(attempt.LastAttempt) > l.WindowDuration*2 {
				delete(l.attempts, ip)
			}
		}

		l.mu.Unlock()
	}
}

// Global rate limiter instance
var loginLimiter *LoginRateLimiter

// Session represents a WhatsApp session
type Session struct {
	ID          string                         `json:"id"`
	Phone       string                         `json:"phone"`        // Session identifier (can be auto-generated)
	ActualPhone string                         `json:"actual_phone"` // Actual WhatsApp phone number after login
	Name        string                         `json:"name"`
	Position    int                            `json:"position"`    // Display position/order
	WebhookURL  string                         `json:"webhook_url"` // Webhook URL for receiving messages
	Client      *whatsmeow.Client              `json:"-"`
	QRChan      <-chan whatsmeow.QRChannelItem `json:"-"`
	Connected   bool                           `json:"connected"`
	LoggedIn    bool                           `json:"logged_in"`
	Connecting  bool                           `json:"-"` // Flag to prevent multiple WebSocket connections
}

// SessionMetadata represents session data stored in database
type SessionMetadata struct {
	ID          string `json:"id"`
	Phone       string `json:"phone"`
	ActualPhone string `json:"actual_phone"`
	Name        string `json:"name"`
	Position    int    `json:"position"`
	WebhookURL  string `json:"webhook_url"`
	CreatedAt   int64  `json:"created_at"`
}

// User represents a user account
type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	Password     string `json:"-"` // Don't include in JSON responses
	Role         string `json:"role"`
	SessionLimit int    `json:"session_limit"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    *int64 `json:"updated_at,omitempty"`
}

// Config holds database and application configuration
type Config struct {
	DatabasePath   string // Path to SQLite database file
	WhatsAppDBPath string // Path to WhatsApp SQLite database file
	JWTSecret      string
	EnableLogging  bool   // Whether to enable logging
	LogLevel       string // Log level: "debug", "info", "warn", "error"
	AdminUsername  string // Default admin username
	AdminPassword  string // Default admin password
}

// Claims represents JWT claims
type Claims struct {
	Username string `json:"username"`
	UserID   int    `json:"user_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// SessionManager manages multiple WhatsApp sessions
type SessionManager struct {
	sessions   map[string]*Session
	store      *sqlstore.Container
	metadataDB *sql.DB
	mu         sync.RWMutex
}

// Global variables
var (
	sessionManager *SessionManager
	config         *Config
	logger         *Logger
	upgrader       = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// Get allowed origins from environment variable
			corsOrigins := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000,http://localhost:8080,http://127.0.0.1:8080")
			
			// Allow all origins if set to "*"
			if corsOrigins == "*" {
				return true
			}
			
			origin := r.Header.Get("Origin")
			
			// Split comma-separated origins
			allowed := strings.Split(corsOrigins, ",")
			for _, allow := range allowed {
				allow = strings.TrimSpace(allow)
				if origin == allow {
					return true
				}
			}
			
			// Also allow if no origin (direct connection)
			return origin == ""
		},
	}
)

// Logger provides configurable logging
type Logger struct {
	enabled bool
	level   string
	logger  *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger(enabled bool, level string) *Logger {
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
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
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
	if l.shouldLog("debug") {
		l.logger.Printf("[DEBUG] "+format, v...)
	}
}

// Info logs info messages
func (l *Logger) Info(format string, v ...interface{}) {
	if l.shouldLog("info") {
		l.logger.Printf("[INFO] "+format, v...)
	}
}

// Warn logs warning messages
func (l *Logger) Warn(format string, v ...interface{}) {
	if l.shouldLog("warn") {
		l.logger.Printf("[WARN] "+format, v...)
	}
}

// Error logs error messages
func (l *Logger) Error(format string, v ...interface{}) {
	if l.shouldLog("error") {
		l.logger.Printf("[ERROR] "+format, v...)
	}
}

// Printf logs formatted messages (for compatibility)
func (l *Logger) Printf(format string, v ...interface{}) {
	l.Info(format, v...)
}

// Println logs messages (for compatibility)
func (l *Logger) Println(v ...interface{}) {
	if l.shouldLog("info") {
		l.logger.Println("[INFO]", fmt.Sprint(v...))
	}
}

// Print logs messages without newline (for compatibility)
func (l *Logger) Print(v ...interface{}) {
	if l.shouldLog("info") {
		l.logger.Print("[INFO] ", fmt.Sprint(v...))
	}
}

// Fatal logs fatal messages and exits
func (l *Logger) Fatal(v ...interface{}) {
	l.logger.Fatal(v...)
}

// generateSessionID generates a unique session identifier
func generateSessionID() string {
	// Generate a random number between 1000000000 and 9999999999 (10 digits)
	min := int64(1000000000)
	max := int64(9999999999)

	n, err := rand.Int(rand.Reader, big.NewInt(max-min))
	if err != nil {
		// Fallback to timestamp-based ID if random fails
		return fmt.Sprintf("session_%d", time.Now().UnixNano()/1000000)
	}

	return fmt.Sprintf("%d", n.Int64()+min)
}

// generatePhoneJID generates a WhatsApp JID format from session ID
func generatePhoneJID(sessionID string) string {
	return sessionID + "@s.whatsapp.net"
}

// loadConfig loads configuration from environment variables or uses defaults
func loadConfig() *Config {
	return &Config{
		DatabasePath:   getEnv("DATABASE_PATH", "./database/session_metadata.db"),
		WhatsAppDBPath: getEnv("WHATSAPP_DB_PATH", "./database/sessions.db"),
		JWTSecret:      getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		EnableLogging:  getEnv("ENABLE_LOGGING", "true") == "true",
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		AdminUsername:  getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:  getEnv("ADMIN_PASSWORD", "admin123"),
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// setupDatabase creates database connection and initializes tables
func setupDatabase(cfg *Config) (*sql.DB, error) {
	// Ensure database directory exists
	dir := filepath.Dir(cfg.DatabasePath)
	if logger != nil {
		logger.Info("Creating metadata database directory: %s", dir)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %v", err)
	}

	// Create empty database file if it doesn't exist
	if logger != nil {
		logger.Info("Initializing metadata database at: %s", cfg.DatabasePath)
	}
	if _, err := os.Stat(cfg.DatabasePath); os.IsNotExist(err) {
		file, err := os.Create(cfg.DatabasePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create database file: %v", err)
		}
		file.Close()
		if logger != nil {
			logger.Info("Created empty metadata database file")
		}
	}

	// SQLite connection only
	db, err := sql.Open("sqlite3", cfg.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping SQLite: %v", err)
	}

	if logger != nil {
		logger.Info("Using SQLite database: %s", cfg.DatabasePath)
	}
	return db, nil
}

// initSessionsTable creates the sessions metadata table if it doesn't exist
func initSessionsTable(db *sql.DB) error {
	// SQLite table creation
	query := `
		CREATE TABLE IF NOT EXISTS session_metadata (
			id TEXT PRIMARY KEY,
			phone TEXT NOT NULL,
			actual_phone TEXT,
			name TEXT,
			position INTEGER DEFAULT 0,
			webhook_url TEXT,
			created_at INTEGER NOT NULL
		)
	`

	_, err := db.Exec(query)
	return err
}

// initUsersTable creates the users table if it doesn't exist
func initUsersTable(db *sql.DB, cfg *Config) error {
	// SQLite table creation
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			session_limit INTEGER NOT NULL DEFAULT 5,
			is_active BOOLEAN NOT NULL DEFAULT 1,
			created_at INTEGER NOT NULL,
			updated_at INTEGER
		)
	`

	_, err := db.Exec(query)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to create users table: %v", err)
			logger.Debug("Query was: %s", query)
		}
		return err
	}

	// Migration: Add new columns if they don't exist
	migrations := []string{
		"ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user'",
		"ALTER TABLE users ADD COLUMN session_limit INTEGER NOT NULL DEFAULT 5",
		"ALTER TABLE users ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT 1",
		"ALTER TABLE users ADD COLUMN updated_at INTEGER",
	}

	for _, migration := range migrations {
		_, err := db.Exec(migration)
		if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
			if logger != nil {
				logger.Debug("Migration failed (non-critical): %v", err)
			}
		}
	}

	// Create default admin user if no users exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		// Get admin credentials from config
		adminUsername := cfg.AdminUsername
		adminPassword := cfg.AdminPassword
		
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		if err != nil {
			if logger != nil {
				logger.Error("Failed to hash password: %v", err)
			}
			return err
		}

		_, err = db.Exec(`
			INSERT INTO users (username, password_hash, role, session_limit, is_active, created_at) 
			VALUES (?, ?, ?, ?, ?, ?)
		`, adminUsername, string(hashedPassword), "admin", -1, 1, time.Now().Unix())

		if err != nil {
			if logger != nil {
				logger.Error("Failed to insert admin user: %v", err)
			}
			return err
		}

		if logger != nil {
			logger.Info("Created default admin user (username: %s, password: %s)", adminUsername, adminPassword)
		}
	} else {
		if logger != nil {
			logger.Info("Found %d existing users, skipping admin user creation", count)
		}
	}

	return nil
}

// saveSessionMetadata saves session metadata to database
func (sm *SessionManager) saveSessionMetadata(session *Session) error {
	// Get next position if not set
	if session.Position == 0 {
		var maxPosition int
		err := sm.metadataDB.QueryRow(`SELECT COALESCE(MAX(position), 0) FROM session_metadata`).Scan(&maxPosition)
		if err != nil {
			maxPosition = 0
		}
		session.Position = maxPosition + 1
	}

	// SQLite INSERT OR REPLACE
	_, err := sm.metadataDB.Exec(`
		INSERT OR REPLACE INTO session_metadata (id, phone, actual_phone, name, position, webhook_url, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, session.ID, session.Phone, session.ActualPhone, session.Name, session.Position, session.WebhookURL, time.Now().Unix())
	return err
}

// loadSessionMetadata loads session metadata from database
func (sm *SessionManager) loadSessionMetadata() ([]SessionMetadata, error) {
	rows, err := sm.metadataDB.Query(`
		SELECT id, phone, actual_phone, name, position, webhook_url, created_at
		FROM session_metadata 
		ORDER BY position ASC, created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []SessionMetadata
	for rows.Next() {
		var s SessionMetadata
		var actualPhone, webhookURL sql.NullString
		var position sql.NullInt64
		err := rows.Scan(&s.ID, &s.Phone, &actualPhone, &s.Name, &position, &webhookURL, &s.CreatedAt)
		if err != nil {
			continue
		}
		if actualPhone.Valid {
			s.ActualPhone = actualPhone.String
		}
		if position.Valid {
			s.Position = int(position.Int64)
		}
		if webhookURL.Valid {
			s.WebhookURL = webhookURL.String
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

// deleteSessionMetadata removes session metadata from database
func (sm *SessionManager) deleteSessionMetadata(sessionID string) error {
	_, err := sm.metadataDB.Exec(`DELETE FROM session_metadata WHERE id = ?`, sessionID)
	return err
}

// generateJWT generates a JWT token for the user
func generateJWT(userID int, username string, role string) (string, error) {
	claims := &Claims{
		Username: username,
		UserID:   userID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.JWTSecret))
}

// validateJWT validates a JWT token and returns claims
func validateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// authMiddleware validates JWT tokens for protected routes
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		claims, err := validateJWT(tokenString)
		if err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Check admin privileges for admin endpoints
		if strings.HasPrefix(r.URL.Path, "/api/admin/") {
			// Role-based access control: only admin role can access admin endpoints
			if claims.Role != "admin" {
				http.Error(w, "Admin privileges required", http.StatusForbidden)
				return
			}
		}

		// Add user info to request context
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
		ctx = context.WithValue(ctx, "role", claims.Role)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	}
}

// authenticateUser validates username and password
func authenticateUser(username, password string) (*User, error) {
	var user User
	var hashedPassword string
	var createdAtInterface interface{}

	err := sessionManager.metadataDB.QueryRow(`
		SELECT id, username, password_hash, role, session_limit, is_active, created_at 
		FROM users 
		WHERE username = ? AND is_active = 1
	`, username).Scan(&user.ID, &user.Username, &hashedPassword, &user.Role, &user.SessionLimit, &user.IsActive, &createdAtInterface)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	// Handle different timestamp formats
	switch v := createdAtInterface.(type) {
	case int64:
		user.CreatedAt = v
	case time.Time:
		user.CreatedAt = v.Unix()
	case []uint8: // MySQL timestamp as bytes
		if t, err := time.Parse("2006-01-02 15:04:05", string(v)); err == nil {
			user.CreatedAt = t.Unix()
		} else {
			user.CreatedAt = time.Now().Unix()
		}
	default:
		user.CreatedAt = time.Now().Unix()
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return &user, nil
}

func init() {
	// Set up logging
	waLog.Stdout("Main", "INFO", true)

	// Set device properties to avoid client outdated error
	store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_CHROME.Enum()
	store.DeviceProps.Os = proto.String("Windows")
	store.DeviceProps.RequireFullSync = proto.Bool(false)

	if logger != nil {
		logger.Info("Initialized with latest whatsmeow version")
	}
}

// restoreSession recreates a session from stored metadata
func restoreSession(metadata SessionMetadata, container *sqlstore.Container) *Session {
	if logger != nil {
		logger.Info("Restoring session %s (%s)", metadata.ID, metadata.Name)
	}

	// Find existing device in store
	var device *store.Device
	devices, err := container.GetAllDevices(context.Background())
	if err != nil {
		if logger != nil {
			logger.Error("Error getting devices: %v", err)
		}
		return nil
	}

	// Look for existing device
	for _, d := range devices {
		if d != nil && d.ID != nil && d.ID.User == strings.Replace(metadata.ActualPhone, "@s.whatsapp.net", "", 1) {
			device = d
			if logger != nil {
				logger.Debug("Found existing device for session %s", metadata.ID)
			}
			break
		}
	}

	// If no existing device found, create new one (will need re-authentication)
	if device == nil {
		device = container.NewDevice()
		if logger != nil {
			logger.Info("Created new device for session %s (will need re-authentication)", metadata.ID)
		}
	}

	// Create client
	clientLog := waLog.Stdout("Client:"+metadata.ID, "INFO", true)
	client := whatsmeow.NewClient(device, clientLog)
	client.EnableAutoReconnect = true
	client.AutoTrustIdentity = true

	// Create session
	session := &Session{
		ID:          metadata.ID,
		Phone:       metadata.Phone,
		ActualPhone: metadata.ActualPhone,
		Name:        metadata.Name,
		Position:    metadata.Position,
		WebhookURL:  metadata.WebhookURL,
		Client:      client,
		Connected:   false,
		LoggedIn:    false,
	}

	// Register event handlers
	client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Connected:
			if logger != nil {
				logger.Info("Session %s connected", metadata.ID)
			}
			sessionManager.mu.Lock()
			if s, ok := sessionManager.sessions[metadata.ID]; ok {
				s.Connected = true
				s.LoggedIn = client.IsLoggedIn()

				// Update actual phone number if logged in
				if client.IsLoggedIn() && client.Store.ID != nil {
					s.ActualPhone = client.Store.ID.User + "@s.whatsapp.net"
					if logger != nil {
						logger.Info("Session %s actual phone: %s", metadata.ID, s.ActualPhone)
					}
					// Save updated metadata
					sessionManager.saveSessionMetadata(s)
				}
			}
			sessionManager.mu.Unlock()
		case *events.Disconnected:
			if logger != nil {
				logger.Info("Session %s disconnected", metadata.ID)
			}
			sessionManager.mu.Lock()
			if s, ok := sessionManager.sessions[metadata.ID]; ok {
				s.Connected = false
			}
			sessionManager.mu.Unlock()
		case *events.Message:
			if logger != nil {
				logger.Info("Received message in session %s: %s", metadata.ID, v.Message.GetConversation())
			}
			// Send webhook if configured
			sessionManager.mu.RLock()
			if s, ok := sessionManager.sessions[metadata.ID]; ok {
				if logger != nil {
					logger.Info("Session %s webhook URL: '%s'", metadata.ID, s.WebhookURL)
				}
				if s.WebhookURL != "" {
					go sendWebhook(s.WebhookURL, metadata.ID, v)
				} else {
					if logger != nil {
						logger.Info("No webhook URL configured for session %s", metadata.ID)
					}
				}
			} else {
				if logger != nil {
					logger.Warn("Session %s not found in sessionManager.sessions", metadata.ID)
				}
			}
			sessionManager.mu.RUnlock()
		}
	})

	// Try to connect if device has stored credentials
	if client.Store.ID != nil {
		go func() {
			if logger != nil {
				logger.Info("Auto-connecting restored session %s", metadata.ID)
			}
			err := client.Connect()
			if err != nil {
				if logger != nil {
					logger.Error("Failed to auto-connect session %s: %v", metadata.ID, err)
				}
			}
		}()
	}

	return session
}

// Admin User Management Handlers

// listUsers returns all users (admin only)
func listUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := sessionManager.metadataDB.Query(`
		SELECT id, username, role, session_limit, is_active, created_at 
		FROM users 
		ORDER BY created_at DESC
	`)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to fetch users: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to fetch users",
		})
		return
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id int
		var username string
		var role string
		var sessionLimit int
		var isActive bool
		var createdAtInterface interface{}

		err := rows.Scan(&id, &username, &role, &sessionLimit, &isActive, &createdAtInterface)
		if err != nil {
			continue
		}

		// Handle created_at timestamp
		var createdAt time.Time
		switch v := createdAtInterface.(type) {
		case int64:
			createdAt = time.Unix(v, 0)
		case string:
			if parsed, err := time.Parse(time.RFC3339, v); err == nil {
				createdAt = parsed
			} else {
				createdAt = time.Now()
			}
		default:
			createdAt = time.Now()
		}

		users = append(users, map[string]interface{}{
			"id":            id,
			"username":      username,
			"role":          role,
			"session_limit": sessionLimit,
			"is_active":     isActive,
			"created_at":    createdAt.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    users,
	})
}

// createUser creates a new user (admin only)
func createUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username     string `json:"username"`
		Password     string `json:"password"`
		Role         string `json:"role"`
		SessionLimit int    `json:"session_limit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// Validate input
	if strings.TrimSpace(req.Username) == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Username is required",
		})
		return
	}

	if len(req.Username) < 3 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Username must be at least 3 characters",
		})
		return
	}

	if len(req.Password) < 6 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Password must be at least 6 characters",
		})
		return
	}

	// Set default values if not provided
	if req.Role == "" {
		req.Role = "user"
	}

	// Validate role
	if req.Role != "admin" && req.Role != "user" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Role must be either 'admin' or 'user'",
		})
		return
	}

	// Set default session limit if not provided
	if req.SessionLimit == 0 {
		req.SessionLimit = 5 // Default limit for regular users
	}

	// Admin users can have unlimited sessions (-1)
	if req.Role == "admin" && req.SessionLimit > 0 {
		req.SessionLimit = -1 // Unlimited for admin
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to hash password: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to hash password",
		})
		return
	}

	// Check if user already exists
	var existingID int
	err = sessionManager.metadataDB.QueryRow("SELECT id FROM users WHERE username = ?", req.Username).Scan(&existingID)
	if err != sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Username already exists",
		})
		return
	}

	// Create user
	result, err := sessionManager.metadataDB.Exec(`
		INSERT INTO users (username, password_hash, role, session_limit, is_active, created_at) 
		VALUES (?, ?, ?, ?, ?, ?)
	`, req.Username, string(hashedPassword), req.Role, req.SessionLimit, 1, time.Now().Unix())

	if err != nil {
		if logger != nil {
			logger.Error("Failed to create user %s: %v", req.Username, err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to create user",
		})
		return
	}

	userID, _ := result.LastInsertId()
	adminUsername := r.Context().Value("username").(string)
	if logger != nil {
		logger.Info("User %s created successfully by admin %s with ID %d", req.Username, adminUsername, userID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User created successfully",
		"data": map[string]interface{}{
			"id":         userID,
			"username":   req.Username,
			"created_at": time.Now().Format(time.RFC3339),
		},
	})
}

// getUserById gets a user by ID (admin only)
func getUserById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	var id int
	var username string
	var createdAtInterface interface{}

	err := sessionManager.metadataDB.QueryRow(`
		SELECT id, username, created_at 
		FROM users 
		WHERE id = ?
	`, userID).Scan(&id, &username, &createdAtInterface)

	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	if err != nil {
		if logger != nil {
			logger.Error("Failed to fetch user %s: %v", userID, err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to fetch user",
		})
		return
	}

	// Handle created_at timestamp
	var createdAt time.Time
	switch v := createdAtInterface.(type) {
	case int64:
		createdAt = time.Unix(v, 0)
	case string:
		if parsed, err := time.Parse(time.RFC3339, v); err == nil {
			createdAt = parsed
		} else {
			createdAt = time.Now()
		}
	default:
		createdAt = time.Now()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"id":         id,
			"username":   username,
			"created_at": createdAt.Format(time.RFC3339),
		},
	})
}

// updateUser updates a user (admin only)
func updateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	var req struct {
		Username     string `json:"username,omitempty"`
		Password     string `json:"password,omitempty"`
		Role         string `json:"role,omitempty"`
		SessionLimit *int   `json:"session_limit,omitempty"`
		IsActive     *bool  `json:"is_active,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// Check if user exists
	var existingUsername string
	err := sessionManager.metadataDB.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&existingUsername)
	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	// Prepare update fields
	updates := []string{}
	args := []interface{}{}

	if req.Username != "" && req.Username != existingUsername {
		if len(req.Username) < 3 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Username must be at least 3 characters",
			})
			return
		}

		// Check if new username already exists
		var duplicateID int
		err = sessionManager.metadataDB.QueryRow("SELECT id FROM users WHERE username = ? AND id != ?", req.Username, userID).Scan(&duplicateID)
		if err != sql.ErrNoRows {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Username already exists",
			})
			return
		}

		updates = append(updates, "username = ?")
		args = append(args, req.Username)
	}

	if req.Password != "" {
		if len(req.Password) < 6 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Password must be at least 6 characters",
			})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			if logger != nil {
				logger.Error("Failed to hash password: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Failed to hash password",
			})
			return
		}

		updates = append(updates, "password_hash = ?")
		args = append(args, string(hashedPassword))
	}

	if req.Role != "" {
		// Validate role
		if req.Role != "admin" && req.Role != "user" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Role must be either 'admin' or 'user'",
			})
			return
		}

		updates = append(updates, "role = ?")
		args = append(args, req.Role)
	}

	if req.SessionLimit != nil {
		// Validate session limit
		if *req.SessionLimit < -1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Session limit must be -1 (unlimited) or greater than 0",
			})
			return
		}

		updates = append(updates, "session_limit = ?")
		args = append(args, *req.SessionLimit)
	}

	if req.IsActive != nil {
		updates = append(updates, "is_active = ?")
		args = append(args, *req.IsActive)
	}

	// Always update the updated_at timestamp
	updates = append(updates, "updated_at = ?")
	args = append(args, time.Now().Unix())

	if len(updates) == 1 { // Only updated_at was added
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "No fields to update",
		})
		return
	}

	// Perform update
	args = append(args, userID)
	query := fmt.Sprintf("UPDATE users SET %s WHERE id = ?", strings.Join(updates, ", "))

	_, err = sessionManager.metadataDB.Exec(query, args...)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to update user %s: %v", userID, err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to update user",
		})
		return
	}

	adminUsername := r.Context().Value("username").(string)
	if logger != nil {
		logger.Info("User %s updated successfully by admin %s", userID, adminUsername)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User updated successfully",
	})
}

// deleteUser deletes a user (admin only)
func deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	// Check if user exists and get username
	var username string
	err := sessionManager.metadataDB.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	// Prevent deleting the admin user
	if username == "admin" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Cannot delete admin user",
		})
		return
	}

	// Delete user
	_, err = sessionManager.metadataDB.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to delete user %s: %v", userID, err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to delete user",
		})
		return
	}

	adminUsername := r.Context().Value("username").(string)
	if logger != nil {
		logger.Info("User %s (%s) deleted successfully by admin %s", userID, username, adminUsername)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User deleted successfully",
	})
}

func main() {
	// Load configuration
	config = loadConfig()

	// Initialize logger
	logger = NewLogger(config.EnableLogging, config.LogLevel)
	logger.Info("Starting WhatsApp Multi-Session Manager with sqlite database")

	// Initialize login rate limiter
	loginLimiter = NewLoginRateLimiter()
	logger.Info("Login rate limiter initialized")

	// Initialize WhatsApp database (always SQLite for whatsmeow)
	dbLog := waLog.Stdout("Database", "INFO", true)
	ctx := context.Background()
	// Ensure WhatsApp database directory exists
	whatsappDBDir := filepath.Dir(config.WhatsAppDBPath)
	logger.Info("Creating WhatsApp database directory: %s", whatsappDBDir)
	if err := os.MkdirAll(whatsappDBDir, 0755); err != nil {
		logger.Fatal("Failed to create WhatsApp database directory:", err)
	}

	// Create empty database file if it doesn't exist
	logger.Info("Initializing WhatsApp database at: %s", config.WhatsAppDBPath)
	if _, err := os.Stat(config.WhatsAppDBPath); os.IsNotExist(err) {
		file, err := os.Create(config.WhatsAppDBPath)
		if err != nil {
			logger.Fatal("Failed to create WhatsApp database file:", err)
		}
		file.Close()
		logger.Info("Created empty WhatsApp database file")
	}

	container, err := sqlstore.New(ctx, "sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on", config.WhatsAppDBPath), dbLog)
	if err != nil {
		logger.Fatal("Failed to initialize WhatsApp database:", err)
	}

	// Initialize metadata database (SQLite)
	metadataDB, err := setupDatabase(config)
	if err != nil {
		if logger != nil {
			logger.Fatal("Failed to setup metadata database:", err)
		} else {
			log.Fatal("Failed to setup metadata database:", err)
		}
	}

	// Initialize database tables
	if logger != nil {
		logger.Info("Initializing database tables...")
	}
	if err := initSessionsTable(metadataDB); err != nil {
		if logger != nil {
			logger.Fatal("Failed to initialize sessions table:", err)
		} else {
			log.Fatal("Failed to initialize sessions table:", err)
		}
	}
	if logger != nil {
		logger.Info("Sessions table initialized successfully")
	}

	if err := initUsersTable(metadataDB, config); err != nil {
		if logger != nil {
			logger.Fatal("Failed to initialize users table:", err)
		} else {
			log.Fatal("Failed to initialize users table:", err)
		}
	}
	if logger != nil {
		logger.Info("Users table initialized successfully")
	}

	// Initialize session manager
	sessionManager = &SessionManager{
		sessions:   make(map[string]*Session),
		store:      container,
		metadataDB: metadataDB,
	}

	// Restore existing sessions
	if logger != nil {
		logger.Info("Loading existing sessions...")
	}
	sessionMetadata, err := sessionManager.loadSessionMetadata()
	if err != nil {
		if logger != nil {
			logger.Error("Error loading session metadata: %v", err)
		}
	} else {
		for _, metadata := range sessionMetadata {
			session := restoreSession(metadata, container)
			if session != nil {
				sessionManager.sessions[session.ID] = session
				if logger != nil {
					logger.Info("Restored session %s (%s)", session.ID, session.Name)
				}
			}
		}
		if logger != nil {
			logger.Info("Restored %d sessions", len(sessionManager.sessions))
		}
	}

	// Set up routes
	router := mux.NewRouter()

	// Public API routes (no authentication required)
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/login", loginHandler).Methods("POST")
	api.HandleFunc("/register", registerHandler).Methods("POST")

	// Protected API routes (authentication required)
	api.HandleFunc("/sessions", authMiddleware(listSessions)).Methods("GET")
	api.HandleFunc("/sessions", authMiddleware(createSession)).Methods("POST")
	api.HandleFunc("/sessions/{id}", authMiddleware(getSession)).Methods("GET")
	api.HandleFunc("/sessions/{id}", authMiddleware(deleteSession)).Methods("DELETE")
	api.HandleFunc("/sessions/{id}/login", authMiddleware(loginSession)).Methods("POST")
	api.HandleFunc("/sessions/{id}/logout", authMiddleware(logoutSession)).Methods("POST")
	api.HandleFunc("/sessions/{id}/send", authMiddleware(sendMessage)).Methods("POST")
	api.HandleFunc("/sessions/{id}/qr", authMiddleware(getQR)).Methods("GET")

	// WebSocket for real-time QR updates (with token query auth)
	api.HandleFunc("/ws/{id}", handleWebSocketWithAuth)

	// API endpoint for sending messages with token and phone selection
	api.HandleFunc("/send", authMiddleware(sendMessageViaAPI)).Methods("POST")
	api.HandleFunc("/sessions/{id}/webhook", authMiddleware(updateSessionWebhook)).Methods("PUT")
	api.HandleFunc("/sessions/{id}/name", authMiddleware(updateSessionName)).Methods("PUT")
	api.HandleFunc("/sessions/{id}/check-number", authMiddleware(checkNumberOnWhatsApp)).Methods("POST")

	// Admin-only endpoints (requires admin privileges)
	api.HandleFunc("/admin/users", authMiddleware(listUsers)).Methods("GET")
	api.HandleFunc("/admin/users", authMiddleware(createUser)).Methods("POST")
	api.HandleFunc("/admin/users/{id}", authMiddleware(getUserById)).Methods("GET")
	api.HandleFunc("/admin/users/{id}", authMiddleware(updateUser)).Methods("PUT")
	api.HandleFunc("/admin/users/{id}", authMiddleware(deleteUser)).Methods("DELETE")
	api.HandleFunc("/sessions/{id}/groups", authMiddleware(listGroups)).Methods("GET")

	// Message attachments and interactions
	api.HandleFunc("/sessions/{id}/send-attachment", authMiddleware(sendAttachment)).Methods("POST")
	api.HandleFunc("/sessions/{id}/typing", authMiddleware(sendTyping)).Methods("POST")
	api.HandleFunc("/sessions/{id}/stop-typing", authMiddleware(stopTyping)).Methods("POST")

	// Serve React frontend
	router.PathPrefix("/").Handler(http.StripPrefix("/", SPAHandler("./frontend/dist/")))

	// Fallback for React Router (SPA)
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If it's an API route, return 404
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		// Otherwise serve index.html for SPA routing
		http.ServeFile(w, r, "./frontend/dist/index.html")
	})

	// Enable CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("API available at http://localhost:8080/api")
	fmt.Println("Frontend available at http://localhost:8080")

	if logger != nil {
		logger.Fatal(http.ListenAndServe(":8080", handler))
	} else {
		log.Fatal(http.ListenAndServe(":8080", handler))
	}
}

// API Handlers

// getClientIP extracts the real client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP in case of multiple proxies
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}

// loginHandler handles user authentication with rate limiting
func loginHandler(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)

	// Check if IP is currently blocked
	if loginLimiter.IsBlocked(clientIP) {
		remaining := loginLimiter.GetRemainingTime(clientIP)
		if logger != nil {
			logger.Warn("Login attempt from blocked IP %s, %v remaining", clientIP, remaining)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", fmt.Sprintf("%.0f", remaining.Seconds()))
		w.WriteHeader(http.StatusTooManyRequests)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":             false,
			"error":               "Too many failed login attempts. Please try again later.",
			"retry_after_seconds": int(remaining.Seconds()),
		})
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if logger != nil {
			logger.Warn("Invalid request body from IP %s: %v", clientIP, err)
		}
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" {
		if logger != nil {
			logger.Warn("Empty credentials from IP %s", clientIP)
		}
		loginLimiter.RecordAttempt(clientIP, false)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Username and password are required",
		})
		return
	}

	// Add a small delay to slow down brute force attempts
	time.Sleep(500 * time.Millisecond)

	user, err := authenticateUser(req.Username, req.Password)
	if err != nil {
		if logger != nil {
			logger.Warn("Authentication failed for user %s from IP %s: %v", req.Username, clientIP, err)
		}
		loginLimiter.RecordAttempt(clientIP, false)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid credentials",
		})
		return
	}

	token, err := generateJWT(user.ID, user.Username, user.Role)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to generate JWT for user %s from IP %s: %v", req.Username, clientIP, err)
		}
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Record successful login (this resets the attempt counter)
	loginLimiter.RecordAttempt(clientIP, true)
	if logger != nil {
		logger.Info("User %s logged in successfully from IP %s", req.Username, clientIP)
	}

	// Add security headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"token":   token,
		"user": map[string]interface{}{
			"id":            user.ID,
			"username":      user.Username,
			"role":          user.Role,
			"session_limit": user.SessionLimit,
			"is_active":     user.IsActive,
		},
	})
}

// registerHandler handles user registration
func registerHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Username) < 3 {
		http.Error(w, "Username must be at least 3 characters", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 6 {
		http.Error(w, "Password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Check if user already exists
	var existingID int
	err = sessionManager.metadataDB.QueryRow("SELECT id FROM users WHERE username = ?", req.Username).Scan(&existingID)
	if err != sql.ErrNoRows {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	// Insert new user
	result, err := sessionManager.metadataDB.Exec(`
		INSERT INTO users (username, password_hash, created_at) 
		VALUES (?, ?, ?)
	`, req.Username, string(hashedPassword), time.Now().Unix())

	if err != nil {
		if logger != nil {
			logger.Error("Failed to create user %s: %v", req.Username, err)
		}
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	userID, _ := result.LastInsertId()
	if logger != nil {
		logger.Info("User %s registered successfully with ID %d", req.Username, userID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User registered successfully",
		"user": map[string]interface{}{
			"id":       userID,
			"username": req.Username,
		},
	})
}

func listSessions(w http.ResponseWriter, r *http.Request) {
	sessionManager.mu.RLock()
	defer sessionManager.mu.RUnlock()

	sessions := make([]*Session, 0, len(sessionManager.sessions))
	for _, session := range sessionManager.sessions {
		sessions = append(sessions, session)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    sessions,
	})
}

func createSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone      string `json:"phone"`
		Name       string `json:"name"`
		Position   int    `json:"position"`
		WebhookURL string `json:"webhook_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get user info from context
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "User authentication required",
		})
		return
	}

	role, ok := r.Context().Value("role").(string)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "User role required",
		})
		return
	}

	// Check session limit (only for non-admin users)
	if role != "admin" {
		// Get user's session limit
		var sessionLimit int
		err := sessionManager.metadataDB.QueryRow(`
			SELECT session_limit FROM users WHERE id = ? AND is_active = 1
		`, userID).Scan(&sessionLimit)

		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Failed to get user session limit",
			})
			return
		}

		// Count current sessions (session limit of -1 means unlimited)
		if sessionLimit != -1 {
			sessionManager.mu.RLock()
			currentSessionCount := len(sessionManager.sessions)
			sessionManager.mu.RUnlock()

			if currentSessionCount >= sessionLimit {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success":          false,
					"error":            fmt.Sprintf("Session limit reached. You can create maximum %d sessions.", sessionLimit),
					"current_sessions": currentSessionCount,
					"session_limit":    sessionLimit,
				})
				return
			}
		}
	}

	sessionManager.mu.Lock()
	defer sessionManager.mu.Unlock()

	// Generate session ID if phone is not provided or is empty
	var sessionID string
	var phoneForDisplay string

	if req.Phone == "" {
		// Auto-generate session ID
		sessionID = generateSessionID()
		phoneForDisplay = generatePhoneJID(sessionID)
		if logger != nil {
			logger.Info("Auto-generated session ID: %s", sessionID)
		}
	} else {
		// Use provided phone number as session ID
		sessionID = req.Phone
		phoneForDisplay = req.Phone
	}

	// Check if session already exists
	if _, exists := sessionManager.sessions[sessionID]; exists {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Session already exists",
		})
		return
	}

	// Create new device
	device := sessionManager.store.NewDevice()

	// Create client with proper logging
	clientLog := waLog.Stdout("Client:"+sessionID, "INFO", true)
	client := whatsmeow.NewClient(device, clientLog)

	// Set client properties for stability
	client.EnableAutoReconnect = true
	client.AutoTrustIdentity = true

	// Create session
	session := &Session{
		ID:          sessionID,
		Phone:       phoneForDisplay,
		ActualPhone: "", // Will be set after login
		Name:        req.Name,
		Position:    req.Position,
		WebhookURL:  req.WebhookURL,
		Client:      client,
		Connected:   false,
		LoggedIn:    false,
	}

	// Register event handlers
	client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Connected:
			if logger != nil {
				logger.Info("Session %s connected", sessionID)
			}
			sessionManager.mu.Lock()
			if s, ok := sessionManager.sessions[sessionID]; ok {
				s.Connected = true
				s.LoggedIn = client.IsLoggedIn()

				// Update actual phone number if logged in
				if client.IsLoggedIn() && client.Store.ID != nil {
					s.ActualPhone = client.Store.ID.User + "@s.whatsapp.net"
					if logger != nil {
						logger.Info("Session %s actual phone: %s", sessionID, s.ActualPhone)
					}
					// Save updated metadata
					sessionManager.saveSessionMetadata(s)
				}
			}
			sessionManager.mu.Unlock()
		case *events.Disconnected:
			if logger != nil {
				logger.Info("Session %s disconnected", sessionID)
			}
			sessionManager.mu.Lock()
			if s, ok := sessionManager.sessions[sessionID]; ok {
				s.Connected = false
			}
			sessionManager.mu.Unlock()
		case *events.Message:
			if logger != nil {
				logger.Info("Received message in session %s: %s", sessionID, v.Message.GetConversation())
			}
			// Send webhook if configured
			sessionManager.mu.RLock()
			if s, ok := sessionManager.sessions[sessionID]; ok && s.WebhookURL != "" {
				go sendWebhook(s.WebhookURL, sessionID, v)
			}
			sessionManager.mu.RUnlock()
		}
	})

	sessionManager.sessions[sessionID] = session

	// Save session metadata to database
	if err := sessionManager.saveSessionMetadata(session); err != nil {
		if logger != nil {
			logger.Error("Failed to save session metadata for %s: %v", sessionID, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    session,
	})
}

func getSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	sessionManager.mu.RLock()
	defer sessionManager.mu.RUnlock()

	session, exists := sessionManager.sessions[id]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    session,
	})
}

func deleteSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	sessionManager.mu.Lock()
	defer sessionManager.mu.Unlock()

	session, exists := sessionManager.sessions[id]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Logout if connected
	if session.Client.IsConnected() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		session.Client.Logout(ctx)
	}

	delete(sessionManager.sessions, id)

	// Delete session metadata from database
	if err := sessionManager.deleteSessionMetadata(id); err != nil {
		if logger != nil {
			logger.Error("Failed to delete session metadata for %s: %v", id, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Session deleted",
	})
}

func loginSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	sessionManager.mu.Lock()
	session, exists := sessionManager.sessions[id]
	sessionManager.mu.Unlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if session.Client.IsConnected() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Already connected",
		})
		return
	}

	// Connect
	err := session.Client.Connect()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Connecting...",
	})
}

func logoutSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	sessionManager.mu.RLock()
	session, exists := sessionManager.sessions[id]
	sessionManager.mu.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := session.Client.Logout(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Connected = false
	session.LoggedIn = false

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Logged out",
	})
}

func sendMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if logger != nil {
		logger.Info("Send message request for session %s", id)
	}

	var req struct {
		To      string `json:"to"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if logger != nil {
			logger.Error("Failed to decode request body: %v", err)
		}
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if strings.TrimSpace(req.To) == "" {
		if logger != nil {
			logger.Warn("Empty recipient field for session %s", id)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Recipient (to) field is required",
		})
		return
	}

	if strings.TrimSpace(req.Message) == "" {
		if logger != nil {
			logger.Warn("Empty message field for session %s", id)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Message field is required",
		})
		return
	}

	if logger != nil {
		logger.Info("Sending message to %s: %s", req.To, req.Message)
	}

	sessionManager.mu.RLock()
	session, exists := sessionManager.sessions[id]
	sessionManager.mu.RUnlock()

	if !exists {
		if logger != nil {
			logger.Error("Session %s not found", id)
		}
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if !session.Client.IsLoggedIn() {
		if logger != nil {
			logger.Error("Session %s not logged in", id)
		}
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	if logger != nil {
		logger.Debug("Session %s is logged in, parsing JID: %s", id, req.To)
	}

	// Clean and format the recipient
	recipient := strings.TrimSpace(req.To)

	// Ensure the recipient is in proper WhatsApp JID format
	var recipientJID string
	if strings.Contains(recipient, "@s.whatsapp.net") {
		recipientJID = recipient
	} else if strings.Contains(recipient, "@g.us") {
		// Group JID
		recipientJID = recipient
	} else {
		// Clean the phone number - remove any non-digit characters except +
		phoneNumber := recipient
		if strings.HasPrefix(phoneNumber, "+") {
			phoneNumber = strings.ReplaceAll(phoneNumber[1:], " ", "")
			phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")
			phoneNumber = strings.ReplaceAll(phoneNumber, "(", "")
			phoneNumber = strings.ReplaceAll(phoneNumber, ")", "")
		} else {
			phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")
			phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")
			phoneNumber = strings.ReplaceAll(phoneNumber, "(", "")
			phoneNumber = strings.ReplaceAll(phoneNumber, ")", "")
		}

		// Validate phone number format (should be digits only after cleaning)
		for _, char := range phoneNumber {
			if char < '0' || char > '9' {
				if logger != nil {
					logger.Warn("Invalid phone number format: %s (cleaned: %s)", recipient, phoneNumber)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "Invalid phone number format. Use digits only (e.g., 6281234567890)",
				})
				return
			}
		}

		if len(phoneNumber) < 8 || len(phoneNumber) > 15 {
			if logger != nil {
				logger.Warn("Invalid phone number length: %s (length: %d)", phoneNumber, len(phoneNumber))
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Invalid phone number length. Should be 8-15 digits",
			})
			return
		}

		// Add @s.whatsapp.net if not present
		recipientJID = phoneNumber + "@s.whatsapp.net"
	}

	if logger != nil {
		logger.Debug("Formatted recipient JID: %s", recipientJID)
	}

	// Parse JID
	jid, err := types.ParseJID(recipientJID)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to parse JID %s: %v", recipientJID, err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid recipient format: " + err.Error(),
		})
		return
	}

	if logger != nil {
		logger.Debug("Parsed JID: %s", jid)
	}

	// Send message
	msg := &waProto.Message{
		Conversation: proto.String(req.Message),
	}

	if logger != nil {
		logger.Info("Sending message via WhatsApp...")
	}
	resp, err := session.Client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to send message: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to send message: " + err.Error(),
		})
		return
	}

	if logger != nil {
		logger.Info("Message sent successfully with ID: %s", resp.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"id":        resp.ID,
			"timestamp": resp.Timestamp,
		},
	})
}

func getQR(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	sessionManager.mu.RLock()
	session, exists := sessionManager.sessions[id]
	sessionManager.mu.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if session.Client.IsLoggedIn() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Already logged in",
		})
		return
	}

	// Disconnect if already connected to reset the state
	if session.Client.IsConnected() {
		session.Client.Disconnect()
	}

	// Get QR channel BEFORE connecting
	qrChan, err := session.Client.GetQRChannel(context.Background())
	if err != nil {
		http.Error(w, "Failed to get QR channel: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Now connect after getting QR channel
	err = session.Client.Connect()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Wait for QR
	select {
	case evt := <-qrChan:
		if evt.Event == "code" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"qr":      evt.Code,
					"timeout": evt.Timeout,
				},
			})
		} else {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "QR event: " + evt.Event,
			})
		}
	}
}

func handleWebSocketWithAuth(w http.ResponseWriter, r *http.Request) {
	// Check for token in query parameters
	token := r.URL.Query().Get("token")
	if token == "" {
		// Try to get from Authorization header
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if token == "" {
		http.Error(w, "No authorization token provided", http.StatusUnauthorized)
		return
	}

	// Validate JWT token
	claims := &Claims{}
	jwtToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JWTSecret), nil
	})

	if err != nil || !jwtToken.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	handleWebSocket(w, r)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	sessionManager.mu.RLock()
	session, exists := sessionManager.sessions[id]
	sessionManager.mu.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Prevent multiple concurrent WebSocket connections for same session
	sessionManager.mu.Lock()
	if session.Connecting {
		sessionManager.mu.Unlock()
		http.Error(w, "QR generation already in progress", http.StatusConflict)
		return
	}
	session.Connecting = true
	sessionManager.mu.Unlock()

	// Ensure we reset the connecting flag when done
	defer func() {
		sessionManager.mu.Lock()
		session.Connecting = false
		sessionManager.mu.Unlock()
	}()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if logger != nil {
			logger.Error("WebSocket upgrade error: %v", err)
		}
		return
	}
	defer conn.Close()

	// Check if already logged in
	if session.Client.IsLoggedIn() {
		conn.WriteJSON(map[string]interface{}{
			"type":    "success",
			"message": "Already logged in",
		})
		return
	}

	if logger != nil {
		logger.Info("WebSocket connected for session %s", id)
	}

	// Only disconnect if we're connected but not logged in (stale connection)
	if session.Client.IsConnected() && !session.Client.IsLoggedIn() {
		if logger != nil {
			logger.Info("Disconnecting stale connection for %s", id)
		}
		session.Client.Disconnect()
		// Give it a moment to disconnect cleanly
		time.Sleep(1 * time.Second)
	}

	// Get QR channel BEFORE connecting
	ctx := context.Background()
	qrChan, err := session.Client.GetQRChannel(ctx)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to get QR channel for %s: %v", id, err)
		}
		conn.WriteJSON(map[string]interface{}{
			"type":  "error",
			"error": "Failed to get QR channel: " + err.Error(),
		})
		return
	}

	// Now connect after getting QR channel
	if logger != nil {
		logger.Info("Connecting session %s...", id)
	}
	err = session.Client.Connect()
	if err != nil {
		if logger != nil {
			logger.Error("Connection error for %s: %v", id, err)
		}
		conn.WriteJSON(map[string]interface{}{
			"type":  "error",
			"error": err.Error(),
		})
		return
	}

	// Channel to signal when to stop
	done := make(chan bool, 1)

	// Send QR codes as they come
	go func() {
		defer func() {
			done <- true
		}()

		for {
			select {
			case evt, ok := <-qrChan:
				if !ok {
					if logger != nil {
						logger.Info("QR channel closed for %s", id)
					}
					return
				}

				if logger != nil {
					logger.Debug("QR event for %s: %s", id, evt.Event)
				}

				if evt.Event == "code" {
					err := conn.WriteJSON(map[string]interface{}{
						"type": "qr",
						"data": map[string]interface{}{
							"qr":      evt.Code,
							"timeout": evt.Timeout,
						},
					})
					if err != nil {
						if logger != nil {
							logger.Error("WebSocket write error: %v", err)
						}
						return
					}
					if logger != nil {
						logger.Info("QR code sent successfully for %s", id)
					}
				} else if evt.Event == "success" {
					conn.WriteJSON(map[string]interface{}{
						"type":    "success",
						"message": "Login successful",
					})
					sessionManager.mu.Lock()
					session.LoggedIn = true
					session.Connected = true
					sessionManager.mu.Unlock()
					if logger != nil {
						logger.Info("Login successful for %s", id)
					}
					return
				} else {
					// Handle other events like timeout
					conn.WriteJSON(map[string]interface{}{
						"type":    "event",
						"event":   evt.Event,
						"message": "QR " + evt.Event,
					})
				}
			case <-done:
				return
			}
		}
	}()

	// Keep connection alive and wait for close
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if logger != nil {
				logger.Error("WebSocket read error for %s: %v", id, err)
			}
			break
		}
	}

	// Signal the goroutine to stop
	select {
	case done <- true:
	default:
	}
}

// SPAHandler serves a Single Page Application with fallback to index.html
func SPAHandler(staticPath string) http.Handler {
	fileServer := http.FileServer(http.Dir(staticPath))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set proper MIME types for assets
		if strings.HasSuffix(r.URL.Path, ".css") {
			w.Header().Set("Content-Type", "text/css")
		} else if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		} else if strings.HasSuffix(r.URL.Path, ".html") {
			w.Header().Set("Content-Type", "text/html")
		}

		// Check if file exists
		path := staticPath + r.URL.Path
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// File does not exist, serve index.html
			w.Header().Set("Content-Type", "text/html")
			http.ServeFile(w, r, staticPath+"index.html")
			return
		}

		// File exists, serve it
		fileServer.ServeHTTP(w, r)
	})
}

// sendMessageViaAPI sends a message via API with phone selection
func sendMessageViaAPI(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone   string `json:"phone"`   // Session phone to use (can be session ID or actual phone)
		To      string `json:"to"`      // Recipient
		Message string `json:"message"` // Message content
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Phone == "" || req.To == "" || req.Message == "" {
		http.Error(w, "Phone, to, and message are required", http.StatusBadRequest)
		return
	}

	sessionManager.mu.RLock()
	defer sessionManager.mu.RUnlock()

	// Find session by phone (session ID or actual phone)
	var selectedSession *Session
	for _, session := range sessionManager.sessions {
		if session.ID == req.Phone ||
			session.Phone == req.Phone ||
			session.ActualPhone == req.Phone ||
			strings.Replace(session.ActualPhone, "@s.whatsapp.net", "", 1) == req.Phone {
			selectedSession = session
			break
		}
	}

	if selectedSession == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Session not found for phone: " + req.Phone,
		})
		return
	}

	if !selectedSession.LoggedIn {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Session is not logged in",
		})
		return
	}

	if logger != nil {
		logger.Info("API send message request for session %s to %s", selectedSession.ID, req.To)
	}

	// Format recipient JID
	recipientJID := req.To
	if !strings.Contains(recipientJID, "@") {
		recipientJID = recipientJID + "@s.whatsapp.net"
	}

	jid, err := types.ParseJID(recipientJID)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to parse JID %s: %v", recipientJID, err)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid recipient format",
		})
		return
	}

	// Send message
	msg := &waProto.Message{
		Conversation: proto.String(req.Message),
	}

	resp, err := selectedSession.Client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to send message: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to send message: " + err.Error(),
		})
		return
	}

	if logger != nil {
		logger.Info("API message sent successfully with ID: %s", resp.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"message_id": resp.ID,
		"timestamp":  resp.Timestamp.Unix(),
		"session":    selectedSession.ID,
	})
}

// sendWebhook sends incoming message data to configured webhook URL
func sendWebhook(webhookURL, sessionID string, message *events.Message) {
	if logger != nil {
		logger.Info("Sending webhook for session %s to URL: %s", sessionID, webhookURL)
	}

	// Detect message type and content
	var messageType string
	var content interface{}
	var mediaInfo map[string]interface{}

	// Check for different message types
	if message.Message.GetConversation() != "" {
		messageType = "text"
		content = message.Message.GetConversation()
	} else if imageMessage := message.Message.GetImageMessage(); imageMessage != nil {
		messageType = "image"
		content = imageMessage.GetCaption()
		mediaInfo = map[string]interface{}{
			"url":         imageMessage.GetURL(),
			"mime_type":   imageMessage.GetMimetype(),
			"file_length": imageMessage.GetFileLength(),
			"width":       imageMessage.GetWidth(),
			"height":      imageMessage.GetHeight(),
			"caption":     imageMessage.GetCaption(),
		}
	} else if videoMessage := message.Message.GetVideoMessage(); videoMessage != nil {
		messageType = "video"
		content = videoMessage.GetCaption()
		mediaInfo = map[string]interface{}{
			"url":         videoMessage.GetURL(),
			"mime_type":   videoMessage.GetMimetype(),
			"file_length": videoMessage.GetFileLength(),
			"duration":    videoMessage.GetSeconds(),
			"width":       videoMessage.GetWidth(),
			"height":      videoMessage.GetHeight(),
			"caption":     videoMessage.GetCaption(),
		}
	} else if audioMessage := message.Message.GetAudioMessage(); audioMessage != nil {
		messageType = "audio"
		content = ""
		mediaInfo = map[string]interface{}{
			"url":         audioMessage.GetURL(),
			"mime_type":   audioMessage.GetMimetype(),
			"file_length": audioMessage.GetFileLength(),
			"duration":    audioMessage.GetSeconds(),
			"voice_note":  audioMessage.GetPTT(), // Push-to-talk (voice note)
		}
	} else if documentMessage := message.Message.GetDocumentMessage(); documentMessage != nil {
		messageType = "document"
		content = documentMessage.GetTitle()
		mediaInfo = map[string]interface{}{
			"url":         documentMessage.GetURL(),
			"mime_type":   documentMessage.GetMimetype(),
			"file_length": documentMessage.GetFileLength(),
			"filename":    documentMessage.GetFileName(),
			"title":       documentMessage.GetTitle(),
		}
	} else if stickerMessage := message.Message.GetStickerMessage(); stickerMessage != nil {
		messageType = "sticker"
		content = ""
		mediaInfo = map[string]interface{}{
			"url":         stickerMessage.GetURL(),
			"mime_type":   stickerMessage.GetMimetype(),
			"file_length": stickerMessage.GetFileLength(),
			"width":       stickerMessage.GetWidth(),
			"height":      stickerMessage.GetHeight(),
		}
	} else if locationMessage := message.Message.GetLocationMessage(); locationMessage != nil {
		messageType = "location"
		content = locationMessage.GetName()
		mediaInfo = map[string]interface{}{
			"latitude":  locationMessage.GetDegreesLatitude(),
			"longitude": locationMessage.GetDegreesLongitude(),
			"name":      locationMessage.GetName(),
			"address":   locationMessage.GetAddress(),
		}
	} else if contactMessage := message.Message.GetContactMessage(); contactMessage != nil {
		messageType = "contact"
		content = contactMessage.GetDisplayName()
		mediaInfo = map[string]interface{}{
			"display_name": contactMessage.GetDisplayName(),
			"vcard":        contactMessage.GetVcard(),
		}
	} else {
		messageType = "unknown"
		content = "Unsupported message type"
	}

	// Prepare webhook payload
	webhookData := map[string]interface{}{
		"session_id": sessionID,
		"timestamp":  message.Info.Timestamp.Unix(),
		"message_id": message.Info.ID,
		"from": map[string]interface{}{
			"jid":       message.Info.Sender.String(),
			"phone":     message.Info.Sender.User,
			"push_name": message.Info.PushName,
		},
		"message_type": messageType,
		"content":      content,
		"is_from_me":   message.Info.IsFromMe,
		"is_group":     message.Info.Sender.Server == "g.us",
	}

	// Add media info if present
	if mediaInfo != nil {
		webhookData["media"] = mediaInfo
	}

	// Add group info if it's a group message
	if message.Info.Sender.Server == "g.us" {
		webhookData["group"] = map[string]interface{}{
			"jid":  message.Info.Chat.String(),
			"name": "", // Group name would need to be fetched separately
		}
	}

	// Convert to JSON
	jsonData, err := json.Marshal(webhookData)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to marshal webhook data for session %s: %v", sessionID, err)
		}
		return
	}

	// Send HTTP POST request to webhook URL
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Post(webhookURL, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		if logger != nil {
			logger.Error("Failed to send webhook for session %s to %s: %v", sessionID, webhookURL, err)
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if logger != nil {
			logger.Info("Webhook sent successfully for session %s to %s", sessionID, webhookURL)
		}
	} else {
		if logger != nil {
			logger.Warn("Webhook failed for session %s to %s: HTTP %d", sessionID, webhookURL, resp.StatusCode)
		}
	}
}

// updateSessionWebhook updates the webhook URL for a specific session
func updateSessionWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req struct {
		WebhookURL string `json:"webhook_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sessionManager.mu.Lock()
	session, exists := sessionManager.sessions[id]
	if !exists {
		sessionManager.mu.Unlock()
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Update session webhook URL
	session.WebhookURL = req.WebhookURL
	sessionManager.mu.Unlock()

	// Update in database
	err := sessionManager.saveSessionMetadata(session)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to save session metadata for %s: %v", id, err)
		}
		http.Error(w, "Failed to update webhook URL", http.StatusInternalServerError)
		return
	}

	if logger != nil {
		logger.Info("Updated webhook URL for session %s: %s", id, req.WebhookURL)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"message":     "Webhook URL updated successfully",
		"webhook_url": req.WebhookURL,
	})
}

// updateSessionName updates the name for a specific session
func updateSessionName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sessionManager.mu.Lock()
	session, exists := sessionManager.sessions[id]
	if !exists {
		sessionManager.mu.Unlock()
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Update session name
	session.Name = req.Name
	sessionManager.mu.Unlock()

	// Update in database
	err := sessionManager.saveSessionMetadata(session)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to save session metadata for %s: %v", id, err)
		}
		http.Error(w, "Failed to update session name", http.StatusInternalServerError)
		return
	}

	if logger != nil {
		logger.Info("Updated name for session %s: %s", id, req.Name)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Session name updated successfully",
		"name":    req.Name,
	})
}

// checkNumberOnWhatsApp checks if a phone number is registered on WhatsApp
func checkNumberOnWhatsApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if logger != nil {
		logger.Info("Check number request for session %s", id)
	}

	var req struct {
		Number string `json:"number"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if logger != nil {
			logger.Error("Failed to decode request body: %v", err)
		}
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Number == "" {
		http.Error(w, "Number is required", http.StatusBadRequest)
		return
	}

	if logger != nil {
		logger.Info("Checking if number %s is on WhatsApp", req.Number)
	}

	sessionManager.mu.RLock()
	session, exists := sessionManager.sessions[id]
	sessionManager.mu.RUnlock()

	if !exists {
		if logger != nil {
			logger.Error("Session %s not found", id)
		}
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if !session.LoggedIn {
		if logger != nil {
			logger.Error("Session %s is not logged in", id)
		}
		http.Error(w, "Session is not logged in", http.StatusUnauthorized)
		return
	}

	// Check if the number is on WhatsApp
	isOnWhatsApp, err := session.Client.IsOnWhatsApp([]string{req.Number})
	if err != nil {
		if logger != nil {
			logger.Error("Failed to check if number %s is on WhatsApp: %v", req.Number, err)
		}
		http.Error(w, "Failed to check number: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var result bool
	var verifiedName string
	if len(isOnWhatsApp) > 0 {
		result = isOnWhatsApp[0].IsIn
		if isOnWhatsApp[0].VerifiedName != nil {
			verifiedName = isOnWhatsApp[0].VerifiedName.Details.GetVerifiedName()
		}
	}

	if logger != nil {
		logger.Info("Number %s is on WhatsApp: %v", req.Number, result)
	}

	response := map[string]interface{}{
		"success":     true,
		"number":      req.Number,
		"on_whatsapp": result,
		"verified":    verifiedName != "",
	}

	if verifiedName != "" {
		response["verified_name"] = verifiedName
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// listGroups retrieves all groups for a session
func listGroups(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if logger != nil {
		logger.Info("List groups request for session %s", id)
	}

	sessionManager.mu.RLock()
	session, exists := sessionManager.sessions[id]
	sessionManager.mu.RUnlock()

	if !exists {
		if logger != nil {
			logger.Error("Session %s not found", id)
		}
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if !session.LoggedIn {
		if logger != nil {
			logger.Error("Session %s is not logged in", id)
		}
		http.Error(w, "Session is not logged in", http.StatusUnauthorized)
		return
	}

	// Get all groups
	groups, err := session.Client.GetJoinedGroups()
	if err != nil {
		if logger != nil {
			logger.Error("Failed to get groups for session %s: %v", id, err)
		}
		http.Error(w, "Failed to retrieve groups: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if logger != nil {
		logger.Info("Found %d groups for session %s", len(groups), id)
	}

	// Format groups data
	var groupList []map[string]interface{}
	for _, group := range groups {
		groupData := map[string]interface{}{
			"jid":               group.JID.String(),
			"name":              group.Name,
			"topic":             group.Topic,
			"owner":             group.OwnerJID.String(),
			"participant_count": len(group.Participants),
			"is_admin":          false,
			"is_super_admin":    false,
		}

		// Check if current user is admin or super admin
		userJID := session.Client.Store.ID
		if userJID != nil {
			for _, participant := range group.Participants {
				if participant.JID.User == userJID.User {
					if participant.IsAdmin {
						groupData["is_admin"] = true
					}
					if participant.IsSuperAdmin {
						groupData["is_super_admin"] = true
					}
					break
				}
			}
		}

		// Add group description if available
		if group.Topic != "" {
			groupData["description"] = group.Topic
		}

		// Add invite link if user is admin (optional, might require additional permission)
		if groupData["is_admin"].(bool) || groupData["is_super_admin"].(bool) {
			// Note: Getting invite link requires additional API call and permissions
			// groupData["invite_link"] = "..." // Can be implemented if needed
		}

		groupList = append(groupList, groupData)
	}

	response := map[string]interface{}{
		"success": true,
		"count":   len(groupList),
		"groups":  groupList,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// sendAttachment sends a message with file attachment
func sendAttachment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if logger != nil {
		logger.Info("Send attachment request for session %s", id)
	}

	// Parse multipart form with size limit (16MB)
	err := r.ParseMultipartForm(16 << 20)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to parse multipart form: %v", err)
		}
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	recipient := r.FormValue("recipient")
	caption := r.FormValue("caption")

	if recipient == "" {
		http.Error(w, "Recipient is required", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		if logger != nil {
			logger.Error("Failed to get file from form: %v", err)
		}
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	sessionManager.mu.RLock()
	session, exists := sessionManager.sessions[id]
	sessionManager.mu.RUnlock()

	if !exists {
		if logger != nil {
			logger.Error("Session %s not found", id)
		}
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if !session.LoggedIn {
		if logger != nil {
			logger.Error("Session %s is not logged in", id)
		}
		http.Error(w, "Session is not logged in", http.StatusUnauthorized)
		return
	}

	// Ensure the recipient is in proper WhatsApp JID format
	var recipientJID string
	if strings.Contains(recipient, "@s.whatsapp.net") || strings.Contains(recipient, "@g.us") {
		recipientJID = recipient
	} else {
		recipientJID = recipient + "@s.whatsapp.net"
	}

	if logger != nil {
		logger.Debug("Formatted recipient JID: %s", recipientJID)
	}

	// Parse JID
	jid, err := types.ParseJID(recipientJID)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to parse JID %s: %v", recipientJID, err)
		}
		http.Error(w, "Invalid recipient format", http.StatusBadRequest)
		return
	}

	// Read file content
	fileData, err := io.ReadAll(file)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to read file: %v", err)
		}
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Detect MIME type
	mimeType := http.DetectContentType(fileData)
	if logger != nil {
		logger.Debug("Detected MIME type: %s for file: %s", mimeType, fileHeader.Filename)
	}

	// Upload media
	if logger != nil {
		logger.Info("Attempting to upload file: %s, size: %d bytes, MIME: %s", fileHeader.Filename, len(fileData), mimeType)
	}
	uploaded, err := session.Client.Upload(context.Background(), fileData, whatsmeow.MediaType(mimeType))
	if err != nil {
		if logger != nil {
			logger.Error("Failed to upload media - File: %s, Size: %d, MIME: %s, Error: %v", fileHeader.Filename, len(fileData), mimeType, err)
		}
		http.Error(w, fmt.Sprintf("Failed to upload file: %v", err), http.StatusInternalServerError)
		return
	}
	if logger != nil {
		logger.Info("Media uploaded successfully: URL=%s, DirectPath=%s", uploaded.URL, uploaded.DirectPath)
	}

	// Create message based on file type
	var message *waProto.Message

	if strings.HasPrefix(mimeType, "image/") {
		message = &waProto.Message{
			ImageMessage: &waProto.ImageMessage{
				URL:           &uploaded.URL,
				DirectPath:    &uploaded.DirectPath,
				MediaKey:      uploaded.MediaKey,
				Mimetype:      &mimeType,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    &uploaded.FileLength,
				Caption:       &caption,
			},
		}
	} else if strings.HasPrefix(mimeType, "video/") {
		message = &waProto.Message{
			VideoMessage: &waProto.VideoMessage{
				URL:           &uploaded.URL,
				DirectPath:    &uploaded.DirectPath,
				MediaKey:      uploaded.MediaKey,
				Mimetype:      &mimeType,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    &uploaded.FileLength,
				Caption:       &caption,
			},
		}
	} else if strings.HasPrefix(mimeType, "audio/") {
		duration := uint32(0) // Would need audio processing to get actual duration
		message = &waProto.Message{
			AudioMessage: &waProto.AudioMessage{
				URL:           &uploaded.URL,
				DirectPath:    &uploaded.DirectPath,
				MediaKey:      uploaded.MediaKey,
				Mimetype:      &mimeType,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    &uploaded.FileLength,
				Seconds:       &duration,
			},
		}
	} else {
		// Generic document
		filename := fileHeader.Filename
		message = &waProto.Message{
			DocumentMessage: &waProto.DocumentMessage{
				URL:           &uploaded.URL,
				DirectPath:    &uploaded.DirectPath,
				MediaKey:      uploaded.MediaKey,
				Mimetype:      &mimeType,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    &uploaded.FileLength,
				FileName:      &filename,
				Title:         &filename,
			},
		}
	}

	// Send the message
	resp, err := session.Client.SendMessage(context.Background(), jid, message)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to send attachment message: %v", err)
		}
		http.Error(w, "Failed to send attachment", http.StatusInternalServerError)
		return
	}

	if logger != nil {
		logger.Info("Attachment sent successfully to %s with message ID: %s", recipient, resp.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"message":    "Attachment sent successfully",
		"message_id": resp.ID,
		"recipient":  recipient,
		"session":    session.ID,
		"filename":   fileHeader.Filename,
		"mime_type":  mimeType,
		"file_size":  len(fileData),
	})
}

// sendTyping sends typing indicator
func sendTyping(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if logger != nil {
		logger.Info("Send typing request for session %s", id)
	}

	var req struct {
		Recipient string `json:"recipient"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if logger != nil {
			logger.Error("Failed to decode request body: %v", err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Recipient == "" {
		http.Error(w, "Recipient is required", http.StatusBadRequest)
		return
	}

	sessionManager.mu.RLock()
	session, exists := sessionManager.sessions[id]
	sessionManager.mu.RUnlock()

	if !exists {
		if logger != nil {
			logger.Error("Session %s not found", id)
		}
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if !session.LoggedIn {
		if logger != nil {
			logger.Error("Session %s is not logged in", id)
		}
		http.Error(w, "Session is not logged in", http.StatusUnauthorized)
		return
	}

	// Ensure the recipient is in proper WhatsApp JID format
	var recipientJID string
	if strings.Contains(req.Recipient, "@s.whatsapp.net") || strings.Contains(req.Recipient, "@g.us") {
		recipientJID = req.Recipient
	} else {
		recipientJID = req.Recipient + "@s.whatsapp.net"
	}

	// Parse JID
	jid, err := types.ParseJID(recipientJID)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to parse JID %s: %v", recipientJID, err)
		}
		http.Error(w, "Invalid recipient format", http.StatusBadRequest)
		return
	}

	// Send typing indicator
	err = session.Client.SendChatPresence(jid, types.ChatPresenceComposing, types.ChatPresenceMediaText)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to send typing indicator: %v", err)
		}
		http.Error(w, "Failed to send typing indicator", http.StatusInternalServerError)
		return
	}

	if logger != nil {
		logger.Info("Typing indicator sent successfully to %s", req.Recipient)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"message":   "Typing indicator sent",
		"recipient": req.Recipient,
		"session":   session.ID,
	})
}

// stopTyping stops typing indicator
func stopTyping(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if logger != nil {
		logger.Info("Stop typing request for session %s", id)
	}

	var req struct {
		Recipient string `json:"recipient"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if logger != nil {
			logger.Error("Failed to decode request body: %v", err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Recipient == "" {
		http.Error(w, "Recipient is required", http.StatusBadRequest)
		return
	}

	sessionManager.mu.RLock()
	session, exists := sessionManager.sessions[id]
	sessionManager.mu.RUnlock()

	if !exists {
		if logger != nil {
			logger.Error("Session %s not found", id)
		}
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if !session.LoggedIn {
		if logger != nil {
			logger.Error("Session %s is not logged in", id)
		}
		http.Error(w, "Session is not logged in", http.StatusUnauthorized)
		return
	}

	// Ensure the recipient is in proper WhatsApp JID format
	var recipientJID string
	if strings.Contains(req.Recipient, "@s.whatsapp.net") || strings.Contains(req.Recipient, "@g.us") {
		recipientJID = req.Recipient
	} else {
		recipientJID = req.Recipient + "@s.whatsapp.net"
	}

	// Parse JID
	jid, err := types.ParseJID(recipientJID)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to parse JID %s: %v", recipientJID, err)
		}
		http.Error(w, "Invalid recipient format", http.StatusBadRequest)
		return
	}

	// Stop typing indicator (send paused state)
	err = session.Client.SendChatPresence(jid, types.ChatPresencePaused, types.ChatPresenceMediaText)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to stop typing indicator: %v", err)
		}
		http.Error(w, "Failed to stop typing indicator", http.StatusInternalServerError)
		return
	}

	if logger != nil {
		logger.Info("Stopped typing indicator for %s", req.Recipient)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"message":   "Typing indicator stopped",
		"recipient": req.Recipient,
		"session":   session.ID,
	})
}
