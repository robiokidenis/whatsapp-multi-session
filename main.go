package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/proto"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

// Session represents a WhatsApp session
type Session struct {
	ID            string              `json:"id"`
	Phone         string              `json:"phone"`          // Session identifier (can be auto-generated)
	ActualPhone   string              `json:"actual_phone"`   // Actual WhatsApp phone number after login
	Name          string              `json:"name"`
	Client        *whatsmeow.Client   `json:"-"`
	QRChan        <-chan whatsmeow.QRChannelItem `json:"-"`
	Connected     bool               `json:"connected"`
	LoggedIn      bool               `json:"logged_in"`
}

// SessionMetadata represents session data stored in database
type SessionMetadata struct {
	ID          string `json:"id"`
	Phone       string `json:"phone"`
	ActualPhone string `json:"actual_phone"`
	Name        string `json:"name"`
	CreatedAt   int64  `json:"created_at"`
}

// User represents a user account
type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"-"` // Don't include in JSON responses
	CreatedAt int64  `json:"created_at"`
}

// Config holds database and application configuration
type Config struct {
	DatabaseType     string // "mysql" or "sqlite"
	DatabaseHost     string
	DatabasePort     string
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
	JWTSecret        string
}

// Claims represents JWT claims
type Claims struct {
	Username string `json:"username"`
	UserID   int    `json:"user_id"`
	jwt.RegisteredClaims
}

// SessionManager manages multiple WhatsApp sessions
type SessionManager struct {
	sessions    map[string]*Session
	store       *sqlstore.Container
	metadataDB  *sql.DB
	mu          sync.RWMutex
}

// Global variables
var (
	sessionManager *SessionManager
	config         *Config
	upgrader       = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for development
		},
	}
)

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
		DatabaseType:     getEnv("DB_TYPE", "mysql"),
		DatabaseHost:     getEnv("DB_HOST", "localhost"),
		DatabasePort:     getEnv("DB_PORT", "3306"),
		DatabaseUser:     getEnv("DB_USER", "root"),
		DatabasePassword: getEnv("DB_PASSWORD", "robioki"),
		DatabaseName:     getEnv("DB_NAME", "whatsapGo"),
		JWTSecret:        getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
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
	var db *sql.DB
	var err error

	if cfg.DatabaseType == "mysql" {
		// MySQL connection string
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4",
			cfg.DatabaseUser,
			cfg.DatabasePassword,
			cfg.DatabaseHost,
			cfg.DatabasePort,
			cfg.DatabaseName,
		)
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to MySQL: %v", err)
		}
		
		// Test connection
		if err := db.Ping(); err != nil {
			return nil, fmt.Errorf("failed to ping MySQL database: %v", err)
		}
		
		log.Printf("Connected to MySQL database %s@%s:%s/%s", cfg.DatabaseUser, cfg.DatabaseHost, cfg.DatabasePort, cfg.DatabaseName)
	} else {
		// SQLite fallback
		db, err = sql.Open("sqlite3", "session_metadata.db")
		if err != nil {
			return nil, fmt.Errorf("failed to open SQLite database: %v", err)
		}
		log.Println("Using SQLite database: session_metadata.db")
	}

	return db, nil
}

// initSessionsTable creates the sessions metadata table if it doesn't exist
func initSessionsTable(db *sql.DB) error {
	var query string
	if config.DatabaseType == "mysql" {
		query = `
			CREATE TABLE IF NOT EXISTS session_metadata (
				id VARCHAR(255) PRIMARY KEY,
				phone VARCHAR(255) NOT NULL,
				actual_phone VARCHAR(255),
				name VARCHAR(255),
				created_at BIGINT NOT NULL,
				INDEX idx_phone (phone),
				INDEX idx_created_at (created_at)
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
		`
	} else {
		query = `
			CREATE TABLE IF NOT EXISTS session_metadata (
				id TEXT PRIMARY KEY,
				phone TEXT NOT NULL,
				actual_phone TEXT,
				name TEXT,
				created_at INTEGER NOT NULL
			)
		`
	}
	
	_, err := db.Exec(query)
	return err
}

// initUsersTable creates the users table if it doesn't exist
func initUsersTable(db *sql.DB) error {
	var query string
	if config.DatabaseType == "mysql" {
		query = `
			CREATE TABLE IF NOT EXISTS users (
				id INT AUTO_INCREMENT PRIMARY KEY,
				username VARCHAR(50) UNIQUE NOT NULL,
				password_hash VARCHAR(255) NOT NULL,
				created_at BIGINT NOT NULL,
				INDEX idx_username (username),
				INDEX idx_created_at (created_at)
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
		`
	} else {
		query = `
			CREATE TABLE IF NOT EXISTS users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				username TEXT UNIQUE NOT NULL,
				password_hash TEXT NOT NULL,
				created_at INTEGER NOT NULL
			)
		`
	}
	
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Failed to create users table: %v", err)
		log.Printf("Query was: %s", query)
		return err
	}

	// Create default admin user if no users exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Failed to hash password: %v", err)
			return err
		}

		_, err = db.Exec(`
			INSERT INTO users (username, password_hash, created_at) 
			VALUES (?, ?, ?)
		`, "admin", string(hashedPassword), time.Now().Unix())
		
		if err != nil {
			log.Printf("Failed to insert admin user: %v", err)
			return err
		}
		
		log.Println("Created default admin user (username: admin, password: admin123)")
	} else {
		log.Printf("Found %d existing users, skipping admin user creation", count)
	}

	return nil
}

// saveSessionMetadata saves session metadata to database
func (sm *SessionManager) saveSessionMetadata(session *Session) error {
	if config.DatabaseType == "mysql" {
		_, err := sm.metadataDB.Exec(`
			REPLACE INTO session_metadata (id, phone, actual_phone, name, created_at)
			VALUES (?, ?, ?, ?, ?)
		`, session.ID, session.Phone, session.ActualPhone, session.Name, time.Now().Unix())
		return err
	} else {
		_, err := sm.metadataDB.Exec(`
			INSERT OR REPLACE INTO session_metadata (id, phone, actual_phone, name, created_at)
			VALUES (?, ?, ?, ?, ?)
		`, session.ID, session.Phone, session.ActualPhone, session.Name, time.Now().Unix())
		return err
	}
}

// loadSessionMetadata loads session metadata from database
func (sm *SessionManager) loadSessionMetadata() ([]SessionMetadata, error) {
	rows, err := sm.metadataDB.Query(`
		SELECT id, phone, actual_phone, name, created_at
		FROM session_metadata
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []SessionMetadata
	for rows.Next() {
		var s SessionMetadata
		var actualPhone sql.NullString
		err := rows.Scan(&s.ID, &s.Phone, &actualPhone, &s.Name, &s.CreatedAt)
		if err != nil {
			continue
		}
		if actualPhone.Valid {
			s.ActualPhone = actualPhone.String
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
func generateJWT(userID int, username string) (string, error) {
	claims := &Claims{
		Username: username,
		UserID:   userID,
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

		// Add user info to request context
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
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
		SELECT id, username, password_hash, created_at 
		FROM users 
		WHERE username = ?
	`, username).Scan(&user.ID, &user.Username, &hashedPassword, &createdAtInterface)
	
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
	
	log.Printf("Initialized with latest whatsmeow version")
}

// restoreSession recreates a session from stored metadata
func restoreSession(metadata SessionMetadata, container *sqlstore.Container) *Session {
	log.Printf("Restoring session %s (%s)", metadata.ID, metadata.Name)
	
	// Find existing device in store
	var device *store.Device
	devices, err := container.GetAllDevices(context.Background())
	if err != nil {
		log.Printf("Error getting devices: %v", err)
		return nil
	}
	
	// Look for existing device
	for _, d := range devices {
		if d != nil && d.ID != nil && d.ID.User == strings.Replace(metadata.ActualPhone, "@s.whatsapp.net", "", 1) {
			device = d
			log.Printf("Found existing device for session %s", metadata.ID)
			break
		}
	}
	
	// If no existing device found, create new one (will need re-authentication)
	if device == nil {
		device = container.NewDevice()
		log.Printf("Created new device for session %s (will need re-authentication)", metadata.ID)
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
		Client:      client,
		Connected:   false,
		LoggedIn:    false,
	}
	
	// Register event handlers
	client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Connected:
			log.Printf("Session %s connected", metadata.ID)
			sessionManager.mu.Lock()
			if s, ok := sessionManager.sessions[metadata.ID]; ok {
				s.Connected = true
				s.LoggedIn = client.IsLoggedIn()
				
				// Update actual phone number if logged in
				if client.IsLoggedIn() && client.Store.ID != nil {
					s.ActualPhone = client.Store.ID.User + "@s.whatsapp.net"
					log.Printf("Session %s actual phone: %s", metadata.ID, s.ActualPhone)
					// Save updated metadata
					sessionManager.saveSessionMetadata(s)
				}
			}
			sessionManager.mu.Unlock()
		case *events.Disconnected:
			log.Printf("Session %s disconnected", metadata.ID)
			sessionManager.mu.Lock()
			if s, ok := sessionManager.sessions[metadata.ID]; ok {
				s.Connected = false
			}
			sessionManager.mu.Unlock()
		case *events.Message:
			log.Printf("Received message in session %s: %s", metadata.ID, v.Message.GetConversation())
		}
	})
	
	// Try to connect if device has stored credentials
	if client.Store.ID != nil {
		go func() {
			log.Printf("Auto-connecting restored session %s", metadata.ID)
			err := client.Connect()
			if err != nil {
				log.Printf("Failed to auto-connect session %s: %v", metadata.ID, err)
			}
		}()
	}
	
	return session
}

func main() {
	// Load configuration
	config = loadConfig()
	log.Printf("Starting WhatsApp Multi-Session Manager with %s database", config.DatabaseType)

	// Initialize WhatsApp database (always SQLite for whatsmeow)
	dbLog := waLog.Stdout("Database", "INFO", true)
	ctx := context.Background()
	container, err := sqlstore.New(ctx, "sqlite3", "file:sessions.db?_foreign_keys=on", dbLog)
	if err != nil {
		log.Fatal("Failed to initialize WhatsApp database:", err)
	}

	// Initialize metadata database (MySQL or SQLite based on config)
	metadataDB, err := setupDatabase(config)
	if err != nil {
		log.Fatal("Failed to setup metadata database:", err)
	}

	// Initialize database tables
	log.Println("Initializing database tables...")
	if err := initSessionsTable(metadataDB); err != nil {
		log.Fatal("Failed to initialize sessions table:", err)
	}
	log.Println("Sessions table initialized successfully")

	if err := initUsersTable(metadataDB); err != nil {
		log.Fatal("Failed to initialize users table:", err)
	}
	log.Println("Users table initialized successfully")

	// Initialize session manager
	sessionManager = &SessionManager{
		sessions:   make(map[string]*Session),
		store:      container,
		metadataDB: metadataDB,
	}

	// Restore existing sessions
	log.Println("Loading existing sessions...")
	sessionMetadata, err := sessionManager.loadSessionMetadata()
	if err != nil {
		log.Printf("Error loading session metadata: %v", err)
	} else {
		for _, metadata := range sessionMetadata {
			session := restoreSession(metadata, container)
			if session != nil {
				sessionManager.sessions[session.ID] = session
				log.Printf("Restored session %s (%s)", session.ID, session.Name)
			}
		}
		log.Printf("Restored %d sessions", len(sessionManager.sessions))
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
	
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// API Handlers

// loginHandler handles user authentication
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := authenticateUser(req.Username, req.Password)
	if err != nil {
		log.Printf("Authentication failed for user %s: %v", req.Username, err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := generateJWT(user.ID, user.Username)
	if err != nil {
		log.Printf("Failed to generate JWT for user %s: %v", req.Username, err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	log.Printf("User %s logged in successfully", req.Username)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"token":   token,
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
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
		log.Printf("Failed to create user %s: %v", req.Username, err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	userID, _ := result.LastInsertId()
	log.Printf("User %s registered successfully with ID %d", req.Username, userID)

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
		Phone string `json:"phone"`
		Name  string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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
		log.Printf("Auto-generated session ID: %s", sessionID)
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
		Client:      client,
		Connected:   false,
		LoggedIn:    false,
	}
	
	// Register event handlers
	client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Connected:
			log.Printf("Session %s connected", sessionID)
			sessionManager.mu.Lock()
			if s, ok := sessionManager.sessions[sessionID]; ok {
				s.Connected = true
				s.LoggedIn = client.IsLoggedIn()
				
				// Update actual phone number if logged in
				if client.IsLoggedIn() && client.Store.ID != nil {
					s.ActualPhone = client.Store.ID.User + "@s.whatsapp.net"
					log.Printf("Session %s actual phone: %s", sessionID, s.ActualPhone)
					// Save updated metadata
					sessionManager.saveSessionMetadata(s)
				}
			}
			sessionManager.mu.Unlock()
		case *events.Disconnected:
			log.Printf("Session %s disconnected", sessionID)
			sessionManager.mu.Lock()
			if s, ok := sessionManager.sessions[sessionID]; ok {
				s.Connected = false
			}
			sessionManager.mu.Unlock()
		case *events.Message:
			log.Printf("Received message in session %s: %s", sessionID, v.Message.GetConversation())
		}
	})

	sessionManager.sessions[sessionID] = session

	// Save session metadata to database
	if err := sessionManager.saveSessionMetadata(session); err != nil {
		log.Printf("Failed to save session metadata for %s: %v", sessionID, err)
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
		log.Printf("Failed to delete session metadata for %s: %v", id, err)
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

	log.Printf("Send message request for session %s", id)

	var req struct {
		To      string `json:"to"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Sending message to %s: %s", req.To, req.Message)

	sessionManager.mu.RLock()
	session, exists := sessionManager.sessions[id]
	sessionManager.mu.RUnlock()

	if !exists {
		log.Printf("Session %s not found", id)
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if !session.Client.IsLoggedIn() {
		log.Printf("Session %s not logged in", id)
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	log.Printf("Session %s is logged in, parsing JID: %s", id, req.To)

	// Ensure the recipient is in proper WhatsApp JID format
	var recipientJID string
	if strings.Contains(req.To, "@s.whatsapp.net") {
		recipientJID = req.To
	} else {
		// Add @s.whatsapp.net if not present
		recipientJID = req.To + "@s.whatsapp.net"
	}
	
	log.Printf("Formatted recipient JID: %s", recipientJID)

	// Parse JID
	jid, err := types.ParseJID(recipientJID)
	if err != nil {
		log.Printf("Failed to parse JID %s: %v", recipientJID, err)
		http.Error(w, "Invalid recipient: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Parsed JID: %s", jid)

	// Send message
	msg := &waProto.Message{
		Conversation: proto.String(req.Message),
	}

	log.Printf("Sending message via WhatsApp...")
	resp, err := session.Client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		http.Error(w, "Failed to send message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Message sent successfully with ID: %s", resp.ID)

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

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
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

	log.Printf("WebSocket connected for session %s", id)

	// Disconnect if already connected to reset the state
	if session.Client.IsConnected() {
		log.Printf("Disconnecting existing connection for %s", id)
		session.Client.Disconnect()
	}

	// Get QR channel BEFORE connecting
	ctx := context.Background()
	qrChan, err := session.Client.GetQRChannel(ctx)
	if err != nil {
		log.Printf("Failed to get QR channel for %s: %v", id, err)
		conn.WriteJSON(map[string]interface{}{
			"type":  "error",  
			"error": "Failed to get QR channel: " + err.Error(),
		})
		return
	}
	
	// Now connect after getting QR channel
	log.Printf("Connecting session %s...", id)
	err = session.Client.Connect()
	if err != nil {
		log.Printf("Connection error for %s: %v", id, err)
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
					log.Printf("QR channel closed for %s", id)
					return
				}
				
				log.Printf("QR event for %s: %s", id, evt.Event)
				
				if evt.Event == "code" {
					err := conn.WriteJSON(map[string]interface{}{
						"type": "qr",
						"data": map[string]interface{}{
							"qr":      evt.Code,
							"timeout": evt.Timeout,
						},
					})
					if err != nil {
						log.Printf("WebSocket write error: %v", err)
						return
					}
					log.Printf("QR code sent successfully for %s", id)
				} else if evt.Event == "success" {
					conn.WriteJSON(map[string]interface{}{
						"type":    "success",
						"message": "Login successful",
					})
					sessionManager.mu.Lock()
					session.LoggedIn = true
					session.Connected = true
					sessionManager.mu.Unlock()
					log.Printf("Login successful for %s", id)
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
			log.Printf("WebSocket read error for %s: %v", id, err)
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