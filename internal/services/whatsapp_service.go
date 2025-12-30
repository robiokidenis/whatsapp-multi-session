package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCompanionReg"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	"whatsapp-multi-session/internal/models"
	"whatsapp-multi-session/internal/repository"
	"whatsapp-multi-session/pkg/logger"

	// Import SQLite driver for whatsmeow store (library requirement)
	_ "github.com/mattn/go-sqlite3"
)

// WhatsAppService manages WhatsApp clients and sessions
type WhatsAppService struct {
	sessions      map[string]*models.Session
	store         *sqlstore.Container
	sessionRepo   *repository.SessionRepository
	logger        *logger.Logger
	mu            sync.RWMutex
	eventHandlers map[string]func(*events.Message)
}

// UserAgentData contains browser and OS information for randomization
type UserAgentData struct {
	Browser   string
	Version   string
	OS        string
	OSVersion string
}

// predefined user agents for randomization
var userAgents = []UserAgentData{
	// Existing entries
	{"Chrome", "120.0.6099.71", "Windows", "10.0"},
	{"Chrome", "119.0.6045.159", "Windows", "11.0"},
	{"Chrome", "120.0.6099.71", "macOS", "14.1"},
	{"Chrome", "119.0.6045.199", "macOS", "13.6"},
	{"Firefox", "121.0", "Windows", "10.0"},
	{"Firefox", "120.0.1", "Windows", "11.0"},
	{"Firefox", "121.0", "macOS", "14.1"},
	{"Edge", "120.0.2210.61", "Windows", "10.0"},
	{"Edge", "119.0.2151.97", "Windows", "11.0"},
	{"Safari", "17.1", "macOS", "14.1"},
	{"Safari", "16.6", "macOS", "13.6"},

	// New Chrome entries
	{"Chrome", "123.0.6312.86", "Windows", "11.0"},
	{"Chrome", "122.0.6261.129", "macOS", "14.0"},
	{"Chrome", "121.0.6167.139", "macOS", "12.7"},
	{"Chrome", "120.0.6099.109", "Windows", "10.0"},
	{"Chrome", "119.0.6045.200", "Linux", "Ubuntu 22.04"},

	// New Firefox entries
	{"Firefox", "124.0", "Windows", "11.0"},
	{"Firefox", "123.0.1", "macOS", "13.5"},
	{"Firefox", "122.0.1", "Linux", "Debian 12"},
	{"Firefox", "121.0.1", "Windows", "10.0"},
	{"Firefox", "120.0", "macOS", "11.6"},

	// New Edge entries
	{"Edge", "123.0.2420.65", "Windows", "11.0"},
	{"Edge", "122.0.2365.66", "Windows", "10.0"},
	{"Edge", "121.0.2277.89", "macOS", "13.6"},
	{"Edge", "120.0.2210.95", "Windows", "11.0"},

	// Safari updates (macOS only)
	{"Safari", "17.5", "macOS", "14.4"},
	{"Safari", "16.3", "macOS", "12.6"},
	{"Safari", "15.6", "macOS", "11.7"},
	{"Safari", "14.1.2", "macOS", "10.15"},

	// Brave (Chromium-based)
	{"Brave", "1.65.132", "Windows", "11.0"},
	{"Brave", "1.64.113", "macOS", "14.0"},

	// Vivaldi (Chromium-based)
	{"Vivaldi", "6.7.3329.19", "Windows", "10.0"},
	{"Vivaldi", "6.6.3271.55", "Linux", "Fedora 39"},

	// Opera (Chromium-based)
	{"Opera", "108.0.5067.29", "Windows", "11.0"},
	{"Opera", "106.0.4998.66", "macOS", "13.4"},
}

func init() {
	// Set up WhatsApp logging
	waLog.Stdout("Main", "INFO", true)

	// Set default device properties - will be randomized per session
	store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_CHROME.Enum()
	store.DeviceProps.Os = proto.String("Windows")
	store.DeviceProps.RequireFullSync = proto.Bool(false)
}

// getRandomUserAgent returns a random user agent configuration
func (s *WhatsAppService) getRandomUserAgent() UserAgentData {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(userAgents))))
	if err != nil {
		// Fallback to first user agent if random fails
		return userAgents[0]
	}
	return userAgents[n.Int64()]
}

// setRandomDeviceProps sets random device properties for a session
func (s *WhatsAppService) setRandomDeviceProps(deviceStore *store.Device) {
	ua := s.getRandomUserAgent()

	// Set browser type based on random selection
	switch ua.Browser {
	case "Chrome":
		store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_CHROME.Enum()
	case "Firefox":
		store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_FIREFOX.Enum()
	case "Edge":
		store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_EDGE.Enum()
	case "Safari":
		store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_SAFARI.Enum()
	default:
		store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_CHROME.Enum()
	}

	// Set OS
	switch ua.OS {
	case "Windows":
		store.DeviceProps.Os = proto.String("Windows")
	case "macOS":
		store.DeviceProps.Os = proto.String("Mac OS")
	default:
		store.DeviceProps.Os = proto.String("Windows")
	}

	// Log the selected user agent for debugging
	s.logger.Info("Using random user agent - Browser: %s %s, OS: %s %s",
		ua.Browser, ua.Version, ua.OS, ua.OSVersion)
}

// NewWhatsAppService creates a new WhatsApp service
//
// DATABASE ISSUES TROUBLESHOOTING:
//
// Symptom: "failed to create WhatsApp store: failed to upgrade database"
// Cause: SQLite database is corrupted or has permission issues
// Solutions:
// 1. Run: ./scripts/reset-whatsapp-db.sh (deletes and recreates database)
// 2. Manual fix: rm -f whatsapp/sessions.db && docker-compose restart
// 3. Permission fix: chmod 666 whatsapp/sessions.db (then restart)
// 4. Check file ownership: ls -la whatsapp/ (should be readable by container user)
//
// Symptom: "attempt to write a readonly database"
// Cause: Database file permissions don't allow container user (UID 1001) to write
// Solutions:
// 1. Fix permissions: chmod 666 whatsapp/sessions.db
// 2. Fix ownership: chown 1001:1001 whatsapp/sessions.db
// 3. Delete and let container recreate: rm whatsapp/sessions.db && docker-compose restart
//
// Symptom: "failed to check if foreign keys are enabled"
// Cause: Database file doesn't exist or is corrupted
// Solutions:
// 1. Delete database: rm whatsapp/sessions.db
// 2. Restart container: docker-compose restart
// 3. Container will create fresh database automatically
//
func NewWhatsAppService(
	dbPath string,
	sessionRepo *repository.SessionRepository,
	log *logger.Logger,
) (*WhatsAppService, error) {
	// Ensure directory exists for WhatsApp database
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create WhatsApp database directory: %v", err)
	}

	// Check if WhatsApp database exists and might be corrupted
	// If there are consistent foreign key errors, we might need to recreate it
	if _, err := os.Stat(dbPath); err == nil {
		log.Info("WhatsApp database exists, using existing: %s", dbPath)
	}

	// Create WhatsApp store - whatsmeow requires SQLite with foreign keys for database upgrades
	waLogger := waLog.Stdout("Store", "INFO", true)

	// Try with foreign keys enabled (required by whatsmeow)
	connectionString := fmt.Sprintf("file:%s?_foreign_keys=on", dbPath)
	container, err := sqlstore.New(context.Background(), "sqlite3", connectionString, waLogger)
	if err != nil {
		// If foreign key error occurs, it might be due to corrupted database
		if strings.Contains(err.Error(), "foreign key") || strings.Contains(err.Error(), "constraint") {
			log.Error("WhatsApp database has foreign key issues: %v", err)
			log.Error("TROUBLESHOOTING:")
			log.Error("  → Run: ./scripts/reset-whatsapp-db.sh")
			log.Error("  → Or manually: rm whatsapp/sessions.db && docker-compose restart")
			log.Error("  → Or fix permissions: chmod 666 whatsapp/sessions.db")
			log.Error("⚠️  This will require re-authentication of all WhatsApp sessions")
		}
		return nil, fmt.Errorf("failed to create WhatsApp store: %v", err)
	}

	service := &WhatsAppService{
		sessions:      make(map[string]*models.Session),
		store:         container,
		sessionRepo:   sessionRepo,
		logger:        log,
		eventHandlers: make(map[string]func(*events.Message)),
	}

	// Load existing sessions
	if err := service.loadExistingSessions(); err != nil {
		return nil, fmt.Errorf("failed to load existing sessions: %v", err)
	}

	return service, nil
}

// CreateSession creates a new WhatsApp session
func (s *WhatsAppService) CreateSession(req *models.CreateSessionRequest, userID int, userRole string) (*models.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate session ID if phone not provided
	sessionID := req.Phone
	phoneForDisplay := req.Phone
	if sessionID == "" {
		sessionID = s.generateSessionID()
		phoneForDisplay = s.generatePhoneJID(sessionID)
	}

	// Check if session already exists
	if _, exists := s.sessions[sessionID]; exists {
		return nil, fmt.Errorf("session %s already exists", sessionID)
	}

	// Check session limit (only for non-admin users)
	if userRole != "admin" {
		// Get current session count for this user
		currentSessionCount, err := s.sessionRepo.CountByUserID(userID)
		if err != nil {
			return nil, fmt.Errorf("failed to check session count: %v", err)
		}

		// Default limit (this should come from user's record in the future)
		sessionLimit := 5

		if currentSessionCount >= sessionLimit {
			return nil, fmt.Errorf("session limit reached. You can create maximum %d sessions", sessionLimit)
		}
	}

	// Create device store with error handling
	deviceStore := s.store.NewDevice()
	if deviceStore == nil {
		return nil, fmt.Errorf("failed to create device store for session %s", sessionID)
	}

	// Set random device properties for this session
	s.setRandomDeviceProps(deviceStore)
	
	// Pre-save the device to avoid foreign key constraints during pairing
	if err := deviceStore.Save(context.Background()); err != nil {
		s.logger.Warn("Failed to pre-save device for session %s: %v", sessionID, err)
		// Continue anyway as this might not be critical
	}

	// Create WhatsApp client with error handling
	clientLog := waLog.Stdout("Client", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	// Validate the client was created successfully
	if client == nil {
		return nil, fmt.Errorf("failed to create WhatsApp client for session %s", sessionID)
	}

	// Enable auto-reconnect like in original
	client.EnableAutoReconnect = true
	client.AutoTrustIdentity = true

	// Create session - default enabled to true unless specified otherwise
	enabled := true
	if req.Enabled {
		enabled = req.Enabled
	}

	session := &models.Session{
		ID:            sessionID,
		Phone:         phoneForDisplay,
		Name:          req.Name,
		Position:      req.Position,
		WebhookURL:    req.WebhookURL,
		AutoReplyText: req.AutoReplyText,
		ProxyConfig:   req.ProxyConfig,
		Enabled:       enabled,
		Client:        client,
		Connected:     false,
		LoggedIn:      false,
		Connecting:    false,
	}

	// Set up event handlers
	s.setupEventHandlers(session)

	// Store in memory
	s.sessions[sessionID] = session

	// Save to database
	metadata := &models.SessionMetadata{
		ID:            sessionID,
		Phone:         phoneForDisplay,
		Name:          req.Name,
		Position:      req.Position,
		WebhookURL:    req.WebhookURL,
		AutoReplyText: req.AutoReplyText,
		ProxyConfig:   req.ProxyConfig,
		Enabled:       enabled,
		UserID:        userID,
		CreatedAt:     time.Now(),
	}

	if err := s.sessionRepo.Create(metadata); err != nil {
		delete(s.sessions, sessionID)
		return nil, fmt.Errorf("failed to save session metadata: %v", err)
	}

	s.logger.Info("Created new session: %s", sessionID)
	return session, nil
}

// GetSession returns a session by ID
func (s *WhatsAppService) GetSession(sessionID string) (*models.Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	return session, exists
}

// GetAllSessions returns all sessions
func (s *WhatsAppService) GetAllSessions() []*models.Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*models.Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

// GetSessionsByUserID returns all sessions for a specific user
func (s *WhatsAppService) GetSessionsByUserID(userID int) ([]*models.Session, error) {
	// Get sessions from database first to get user ownership info
	sessionMetadata, err := s.sessionRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions for user %d: %v", userID, err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var userSessions []*models.Session
	for _, metadata := range sessionMetadata {
		if session, exists := s.sessions[metadata.ID]; exists {
			userSessions = append(userSessions, session)
		}
	}

	return userSessions, nil
}

// IsSessionOwnedByUser checks if a session belongs to a specific user
func (s *WhatsAppService) IsSessionOwnedByUser(sessionID string, userID int) (bool, error) {
	sessionMetadata, err := s.sessionRepo.GetByIDAndUserID(sessionID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check session ownership: %v", err)
	}

	return sessionMetadata != nil, nil
}

// FindSessionByPhone finds a session by phone identifier (session ID or actual phone)
func (s *WhatsAppService) FindSessionByPhone(phoneIdentifier string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Try to find session by various identifiers
	for _, session := range s.sessions {
		if session.ID == phoneIdentifier ||
			session.Phone == phoneIdentifier ||
			session.ActualPhone == phoneIdentifier ||
			strings.Replace(session.ActualPhone, "@s.whatsapp.net", "", 1) == phoneIdentifier {
			return session.ID
		}
	}

	return ""
}

// ConnectSession initiates connection for a session
func (s *WhatsAppService) ConnectSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return models.NewNotFoundError("session %s not found", sessionID)
	}

	// Check if session is enabled
	if !session.Enabled {
		return models.NewBadRequestError("session %s is disabled and cannot be connected", sessionID)
	}

	if session.Connecting {
		return models.NewBadRequestError("session %s is already connecting", sessionID)
	}

	if session.Connected {
		return models.NewBadRequestError("session %s is already connected", sessionID)
	}

	session.Connecting = true

	go func() {
		defer func() {
			// Only clear connecting flag if we didn't successfully connect
			if !session.Connected {
				s.mu.Lock()
				session.Connecting = false
				s.mu.Unlock()
			}
		}()

		s.logger.Info("Attempting to connect session %s...", sessionID)

		if err := session.Client.Connect(); err != nil {
			s.logger.Error("Failed to connect session %s: %v", sessionID, err)
			s.mu.Lock()
			session.Connecting = false
			s.mu.Unlock()
			return
		}

		s.logger.Info("Connect() called for session %s, waiting for connection event...", sessionID)

		// Wait up to 30 seconds for the connection to be established
		timeout := time.After(30 * time.Second)
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				s.logger.Warn("Session %s connection timed out after 30 seconds", sessionID)
				s.mu.Lock()
				session.Connecting = false
				s.mu.Unlock()
				return
			case <-ticker.C:
				s.mu.Lock()
				isConnected := session.Client.IsConnected() && session.Connected
				s.mu.Unlock()

				if isConnected {
					s.logger.Info("Session %s connection established successfully", sessionID)
					s.mu.Lock()
					session.Connecting = false
					s.mu.Unlock()
					return
				}
				s.logger.Debug("Session %s still connecting... (IsConnected: %v, Connected: %v)",
					sessionID, session.Client.IsConnected(), session.Connected)
			}
		}
	}()

	return nil
}

// DisconnectSession disconnects a session
func (s *WhatsAppService) DisconnectSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return models.NewNotFoundError("session %s not found", sessionID)
	}

	session.Client.Disconnect()
	session.Connected = false
	session.LoggedIn = false

	s.logger.Info("Session %s disconnected", sessionID)
	return nil
}

// DeleteSession removes a session completely
func (s *WhatsAppService) DeleteSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return models.NewNotFoundError("session %s not found", sessionID)
	}

	// Disconnect if connected
	if session.Connected {
		session.Client.Disconnect()
	}

	// Remove from memory
	delete(s.sessions, sessionID)

	// Remove from database
	if err := s.sessionRepo.Delete(sessionID); err != nil {
		s.logger.Error("Failed to delete session from database: %v", err)
	}

	s.logger.Info("Session %s deleted", sessionID)
	return nil
}

// LoginSession initiates the login process for a session
func (s *WhatsAppService) LoginSession(sessionID string) error {
	session, exists := s.GetSession(sessionID)
	if !exists {
		return models.NewNotFoundError("session %s not found", sessionID)
	}

	// Check if session is enabled
	if !session.Enabled {
		return models.NewBadRequestError("session %s is disabled and cannot be logged in", sessionID)
	}

	if session.LoggedIn {
		return models.NewBadRequestError("session %s is already logged in", sessionID)
	}

	// Connect if not connected
	if !session.Connected {
		if err := s.ConnectSession(sessionID); err != nil {
			return fmt.Errorf("failed to connect session for login: %v", err)
		}

		// Wait a bit for connection to establish
		time.Sleep(2 * time.Second)
	}

	// Start QR code generation for login
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	qrChan, err := session.Client.GetQRChannel(ctx)
	if err != nil {
		return fmt.Errorf("failed to get QR channel: %v", err)
	}

	session.QRChan = qrChan
	s.logger.Info("QR code channel created for session %s", sessionID)

	return nil
}

// LogoutSession logs out a session
func (s *WhatsAppService) LogoutSession(sessionID string) error {
	session, exists := s.GetSession(sessionID)
	if !exists {
		return models.NewNotFoundError("session %s not found", sessionID)
	}

	if !session.LoggedIn {
		return models.NewBadRequestError("session %s is not logged in", sessionID)
	}

	// Perform logout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := session.Client.Logout(ctx)
	if err != nil {
		return fmt.Errorf("failed to logout session: %v", err)
	}

	// Update session state
	s.mu.Lock()
	session.LoggedIn = false
	s.mu.Unlock()

	s.logger.Info("Session %s logged out successfully", sessionID)
	return nil
}

// GetQRCode returns QR code for session login
func (s *WhatsAppService) GetQRCode(sessionID string) (string, error) {
	session, exists := s.GetSession(sessionID)
	if !exists {
		return "", models.NewNotFoundError("session %s not found", sessionID)
	}

	if session.LoggedIn {
		return "", models.NewBadRequestError("session %s is already logged in", sessionID)
	}

	// Start QR code generation
	qrChan, err := session.Client.GetQRChannel(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get QR channel: %v", err)
	}

	session.QRChan = qrChan

	// Wait for QR code
	select {
	case qr := <-qrChan:
		if qr.Event == "code" {
			return qr.Code, nil
		}
	case <-time.After(30 * time.Second):
		return "", fmt.Errorf("timeout waiting for QR code")
	}

	return "", fmt.Errorf("failed to get QR code")
}

// UpdateSession updates session metadata
func (s *WhatsAppService) UpdateSession(sessionID string, req *models.UpdateSessionRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return models.NewNotFoundError("session %s not found", sessionID)
	}

	// Get existing metadata from database to get user_id
	existingMetadata, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get existing session metadata: %v", err)
	}
	if existingMetadata == nil {
		return models.NewNotFoundError("session metadata not found in database for session %s", sessionID)
	}

	// Update in-memory session
	if req.Name != "" {
		session.Name = req.Name
	}
	if req.WebhookURL != "" {
		session.WebhookURL = req.WebhookURL
	}
	if req.AutoReplyText != nil {
		session.AutoReplyText = req.AutoReplyText
	}
	if req.ProxyConfig != nil {
		session.ProxyConfig = req.ProxyConfig
	}
	if req.Enabled != nil {
		session.Enabled = *req.Enabled
	}

	// Update in database with correct user_id
	metadata := &models.SessionMetadata{
		ID:            sessionID,
		Phone:         session.Phone,
		ActualPhone:   session.ActualPhone,
		Name:          session.Name,
		Position:      req.Position,
		WebhookURL:    session.WebhookURL,
		AutoReplyText: session.AutoReplyText,
		ProxyConfig:   session.ProxyConfig,
		Enabled:       session.Enabled,
		UserID:        existingMetadata.UserID, // Use the existing user_id from database
	}

	return s.sessionRepo.Update(metadata)
}

// UpdateSessionAutoReply updates only the auto-reply text for a session
func (s *WhatsAppService) UpdateSessionAutoReply(sessionID string, autoReplyText *string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return models.NewNotFoundError("session %s not found", sessionID)
	}

	// Update in-memory session
	session.AutoReplyText = autoReplyText

	// Update only the auto-reply text in database using the dedicated method
	return s.sessionRepo.UpdateAutoReplyText(sessionID, autoReplyText)
}

// UpdateSessionWebhook updates only the webhook URL for a session
func (s *WhatsAppService) UpdateSessionWebhook(sessionID string, webhookURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return models.NewNotFoundError("session %s not found", sessionID)
	}

	// Update in-memory session
	session.WebhookURL = webhookURL

	// Update in database
	if err := s.sessionRepo.UpdateSessionWebhook(sessionID, webhookURL); err != nil {
		return fmt.Errorf("failed to update webhook URL in database: %v", err)
	}

	s.logger.Info("Updated webhook URL for session %s: %s", sessionID, webhookURL)
	return nil
}

// UpdateSessionEnabled updates only the enabled status for a session
func (s *WhatsAppService) UpdateSessionEnabled(sessionID string, enabled bool) error {
	s.mu.Lock()
	session, exists := s.sessions[sessionID]
	if !exists {
		s.mu.Unlock()
		return models.NewNotFoundError("session %s not found", sessionID)
	}

	// Store previous state
	wasDisabled := !session.Enabled

	// Update in-memory session
	session.Enabled = enabled
	s.mu.Unlock()

	// Update only the enabled status in database using the dedicated method
	if err := s.sessionRepo.UpdateSessionEnabled(sessionID, enabled); err != nil {
		return err
	}

	// If session was disabled and is now enabled, try to auto-connect
	if wasDisabled && enabled {
		s.logger.Info("Session %s has been enabled, attempting to auto-connect", sessionID)

		// Check if session has stored credentials and is not already connected
		if !session.Connected && session.Client != nil {
			go func() {
				// Wait a moment to ensure everything is settled
				time.Sleep(1 * time.Second)

				s.logger.Info("Auto-connecting newly enabled session %s", sessionID)
				if err := session.Client.Connect(); err != nil {
					s.logger.Error("Failed to auto-connect enabled session %s: %v", sessionID, err)
				} else {
					s.logger.Info("Successfully connected enabled session %s", sessionID)
				}
			}()
		} else if session.Connected {
			s.logger.Info("Session %s is already connected", sessionID)
		} else {
			s.logger.Info("Session %s needs re-authentication (no stored credentials)", sessionID)
		}
	} else if !wasDisabled && !enabled && session.Connected {
		// If session was enabled and is now disabled, disconnect it
		s.logger.Info("Session %s has been disabled, disconnecting", sessionID)
		go func() {
			if session.Client != nil {
				session.Client.Disconnect()
				s.logger.Info("Disconnected disabled session %s", sessionID)
			}
		}()
	}

	return nil
}

// setupEventHandlers sets up event handlers for a session
//
// TROUBLESHOOTING CONNECTION ISSUES:
//
// 1. Symptom: "received Connected event but client is not actually connected"
//    Cause: whatsmeow library fired Connected event but websocket failed to establish
//    Solutions:
//    a) Check network connectivity to WhatsApp servers
//    b) Verify no firewall/proxy is blocking websocket connections
//    c) Try updating whatsmeow: go get go.mau.fi/whatsmeow@latest
//    d) Delete session and recreate (device may be corrupted)
//
// 2. Symptom: "Stream error, disconnecting"
//    Cause: WhatsApp server closed the connection due to protocol error or session issue
//    Solutions:
//    a) Session may be logged out elsewhere - reconnect
//    b) Update whatsmeow library for latest protocol fixes
//    c) Clear device store: delete whatsapp/sessions.db and recreate session
//
// 3. Symptom: Connection succeeds but messages fail with "websocket not connected"
//    Cause: Race condition between connection event and actual socket ready state
//    Solutions:
//    a) Add delay before sending messages after connect
//    b) Check session.Client.IsConnected() before sending
//    c) This should be fixed in current code - verify you have latest version
//
// 4. Symptom: "Failed to set online presence: can't send presence without PushName set"
//    Cause: Device store missing PushName required by WhatsApp
//    Solutions:
//    a) This is auto-fixed in current code by generating random PushName
//    b) If issue persists, manually set: session.Client.Store.PushName = "SomeName"
//
// 5. General connection issues:
//    - Check whatsmeow version: go list -m go.mau.fi/whatsmeow
//    - Update if outdated: go get go.mau.fi/whatsmeow@latest
//    - Check for GitHub issues: https://github.com/mautic/whatsmeow/issues
//    - Verify container has internet access
//
func (s *WhatsAppService) setupEventHandlers(session *models.Session) {
	session.Client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Connected:
			s.mu.Lock()
			// Verify the client is actually connected before marking as connected
			if session.Client.IsConnected() {
				session.Connected = true
				session.LoggedIn = session.Client.IsLoggedIn()

				// Update actual phone number if logged in (like original)
				if session.Client.IsLoggedIn() && session.Client.Store.ID != nil {
					session.ActualPhone = session.Client.Store.ID.User + "@s.whatsapp.net"
					s.logger.Info("Session %s actual phone: %s", session.ID, session.ActualPhone)

					// Set online presence for better typing indicator support
					go func() {
						// Ensure PushName is set before sending presence
						if session.Client.Store.PushName == "" {
							session.Client.Store.PushName = s.generateRandomName()
							s.logger.Debug("Set random push name for session %s", session.ID)
						}

						if err := session.Client.SendPresence(context.Background(), types.PresenceAvailable); err != nil {
							s.logger.Warn("Failed to set online presence for session %s: %v", session.ID, err)
						} else {
							s.logger.Debug("Set online presence for session %s", session.ID)
						}
					}()

					// Save updated metadata
					go func() {
						metadata := &models.SessionMetadata{
							ID:            session.ID,
							Phone:         session.Phone,
							ActualPhone:   session.ActualPhone,
							Name:          session.Name,
							Position:      session.Position,
							WebhookURL:    session.WebhookURL,
							AutoReplyText: session.AutoReplyText,
						}
						if err := s.sessionRepo.Update(metadata); err != nil {
							s.logger.Error("Failed to update session metadata: %v", err)
						}
					}()
				}
				s.mu.Unlock()
				s.logger.Info("Session %s connected", session.ID)
			} else {
				// Client reported Connected event but isn't actually connected
				session.Connected = false
				s.mu.Unlock()

				// TROUBLESHOOTING: This means whatsmeow fired Connected event but socket isn't connected
				// Possible causes:
				// 1. whatsmeow library bug - check for updates: go get go.mau.fi/whatsmeow@latest
				// 2. Network/firewall blocking websocket connection
				// 3. WhatsApp server rejected connection
				// 4. Device store corrupted - try deleting and recreating session
				s.logger.Error("Session %s CONNECTION ISSUE: Received Connected event but websocket is not connected!", session.ID)
				s.logger.Error("  → Check whatsmeow version: go list -m go.mau.fi/whatsmeow")
				s.logger.Error("  → Update if needed: go get go.mau.fi/whatsmeow@latest && go mod tidy")
				s.logger.Error("  → Check network connectivity and firewall settings")
				s.logger.Error("  → Try recreating the session (delete and create new)")
			}

		case *events.Disconnected:
			s.mu.Lock()
			session.Connected = false
			session.LoggedIn = false
			s.mu.Unlock()

			s.logger.Info("Session %s disconnected", session.ID)

		case *events.StreamError:
			s.mu.Lock()
			session.Connected = false
			session.LoggedIn = false
			s.mu.Unlock()

			// TROUBLESHOOTING: WhatsApp closed connection due to protocol error
			// Possible causes:
			// 1. Protocol version mismatch - update whatsmeow
			// 2. Session logged out from another device
			// 3. WhatsApp server rejected the connection
			// 4. Device store corrupted
			s.logger.Error("Session %s STREAM ERROR: WhatsApp closed the connection", session.ID)
			s.logger.Error("  → Check whatsmeow for protocol updates: go get go.mau.fi/whatsmeow@latest")
			s.logger.Error("  → Session may need re-authentication - scan QR code again")
			s.logger.Error("  → If issue persists, delete and recreate the session")

		case *events.LoggedOut:
			s.mu.Lock()
			session.LoggedIn = false
			session.ActualPhone = ""
			s.mu.Unlock()

			s.logger.Info("Session %s logged out", session.ID)

			// Update database to clear actual phone
			go func() {
				metadata := &models.SessionMetadata{
					ID:            session.ID,
					Phone:         session.Phone,
					ActualPhone:   "",
					Name:          session.Name,
					Position:      session.Position,
					WebhookURL:    session.WebhookURL,
					AutoReplyText: session.AutoReplyText,
				}
				if err := s.sessionRepo.Update(metadata); err != nil {
					s.logger.Error("Failed to update session metadata after logout: %v", err)
				}
			}()

		case *events.Message:
			// Handle incoming messages
			if handler, exists := s.eventHandlers[session.ID]; exists {
				handler(v)
			}

			// Only process auto-reply and webhook if session is enabled
			if session.Enabled {
				// Send auto reply if configured and this is an incoming message
				if session.AutoReplyText != nil && *session.AutoReplyText != "" && !v.Info.IsFromMe {
					go s.sendAutoReply(session, v)
				}

				// Send webhook if configured
				if session.WebhookURL != "" {
					go s.sendWebhook(session, v)
				}
			} else {
				// Log that the session is disabled and won't process messages
				if !v.Info.IsFromMe {
					s.logger.Debug("Session %s is disabled, skipping auto-reply and webhook for message from %s", session.ID, v.Info.Sender.User)
				}
			}
		}
	})
}

// SetMessageHandler sets a message handler for a session
func (s *WhatsAppService) SetMessageHandler(sessionID string, handler func(*events.Message)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.eventHandlers[sessionID] = handler
}

// loadExistingSessions loads sessions from database and reconnects if needed
func (s *WhatsAppService) loadExistingSessions() error {
	metadatas, err := s.sessionRepo.GetAll()
	if err != nil {
		return fmt.Errorf("failed to get sessions from database: %v", err)
	}

	for _, metadata := range metadatas {
		s.logger.Info("Restoring session %s (%s)", metadata.ID, metadata.Name)

		// Debug log to verify webhook, auto-reply and proxy are loaded
		if metadata.WebhookURL != "" {
			s.logger.Info("Session %s has webhook URL: %s", metadata.ID, metadata.WebhookURL)
		}
		if metadata.AutoReplyText != nil && *metadata.AutoReplyText != "" {
			s.logger.Info("Session %s has auto-reply text: %s", metadata.ID, *metadata.AutoReplyText)
		}
		if metadata.ProxyConfig != nil && metadata.ProxyConfig.Enabled {
			s.logger.Info("Session %s has proxy enabled: %s://%s:%d", metadata.ID, metadata.ProxyConfig.Type, metadata.ProxyConfig.Host, metadata.ProxyConfig.Port)
		}

		// Find existing device in store (like original)
		var deviceStore *store.Device
		devices, err := s.store.GetAllDevices(context.Background())
		if err != nil {
			s.logger.Error("Error getting devices: %v", err)
		} else {
			// Look for existing device by comparing JID
			for _, d := range devices {
				if d != nil && d.ID != nil {
					// Compare by actual phone if available, otherwise try by session ID
					deviceUser := d.ID.User
					if metadata.ActualPhone != "" && deviceUser == strings.Replace(metadata.ActualPhone, "@s.whatsapp.net", "", 1) {
						deviceStore = d
						s.logger.Debug("Found existing device for session %s by actual phone", metadata.ID)
						break
					} else if deviceUser == metadata.ID {
						deviceStore = d
						s.logger.Debug("Found existing device for session %s by ID", metadata.ID)
						break
					}
				}
			}
		}

		// If no existing device found, create new one (will need re-authentication)
		if deviceStore == nil {
			deviceStore = s.store.NewDevice()
			s.logger.Info("Created new device for session %s (will need re-authentication)", metadata.ID)
		}

		// Create WhatsApp client
		clientLog := waLog.Stdout("Client:"+metadata.ID, "INFO", true)
		client := whatsmeow.NewClient(deviceStore, clientLog)

		// Enable auto-reconnect like in original
		client.EnableAutoReconnect = true
		client.AutoTrustIdentity = true

		// Create session
		session := &models.Session{
			ID:            metadata.ID,
			Phone:         metadata.Phone,
			ActualPhone:   metadata.ActualPhone,
			Name:          metadata.Name,
			Position:      metadata.Position,
			WebhookURL:    metadata.WebhookURL,
			AutoReplyText: metadata.AutoReplyText,
			ProxyConfig:   metadata.ProxyConfig,
			Enabled:       metadata.Enabled,
			Client:        client,
			Connected:     false,
			LoggedIn:      false,
			Connecting:    false,
		}

		// Set up event handlers
		s.setupEventHandlers(session)

		// Store in memory
		s.sessions[metadata.ID] = session

		// Try to connect if device has stored credentials and session is enabled
		if deviceStore != nil && deviceStore.ID != nil {
			if metadata.Enabled {
				go func(sessionID string, client *whatsmeow.Client) {
					// Wait a bit before connecting to ensure everything is initialized
					time.Sleep(2 * time.Second)

					s.logger.Info("Auto-connecting restored session %s with JID %s", sessionID, deviceStore.ID.String())
					err := client.Connect()
					if err != nil {
						s.logger.Error("Failed to auto-connect session %s: %v", sessionID, err)
					}
				}(metadata.ID, client)
			} else {
				s.logger.Info("Session %s is disabled, skipping auto-connect", metadata.ID)
			}
		} else {
			s.logger.Info("Session %s needs re-authentication (no stored credentials)", metadata.ID)
		}
	}

	s.logger.Info("Restored %d sessions", len(s.sessions))
	return nil
}

// generateSessionID generates a unique session ID
func (s *WhatsAppService) generateSessionID() string {
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
func (s *WhatsAppService) generatePhoneJID(sessionID string) string {
	return sessionID + "@s.whatsapp.net"
}

// SendMessage sends a text message
func (s *WhatsAppService) SendMessage(sessionID string, req *models.SendMessageRequest) (string, error) {
	session, exists := s.GetSession(sessionID)
	if !exists {
		return "", models.NewNotFoundError("session not found")
	}

	// Check if session is connected
	if !session.Connected {
		return "", models.NewServiceUnavailableError("session is not connected. Please connect the session first")
	}

	// Check if session is logged in
	if !session.LoggedIn {
		return "", models.NewUnauthorizedError("session is not authenticated. Please scan QR code to login")
	}

	// Format recipient JID like original implementation
	recipientJID := req.To

	// Check if it's a phone number (contains only digits and +)
	if !strings.Contains(req.To, "@") {
		phoneNumber := strings.ReplaceAll(req.To, "+", "")
		phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")
		phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")

		if len(phoneNumber) < 8 || len(phoneNumber) > 15 {
			return "", fmt.Errorf("invalid phone number length. Should be 8-15 digits")
		}

		// Add @s.whatsapp.net if not present
		recipientJID = phoneNumber + "@s.whatsapp.net"
	}

	s.logger.Debug("Formatted recipient JID: %s", recipientJID)

	// Parse recipient JID
	jid, err := types.ParseJID(recipientJID)
	if err != nil {
		return "", fmt.Errorf("invalid recipient JID: %v", err)
	}

	// Convert to non-device JID for message sending
	jid = jid.ToNonAD()

	// Send message
	msg := &waProto.Message{
		Conversation: proto.String(req.Message),
	}

	resp, err := session.Client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		return "", fmt.Errorf("failed to send message: %v", err)
	}

	return resp.ID, nil
}

// ForwardMessage forwards an existing message to another recipient
func (s *WhatsAppService) ForwardMessage(sessionID string, req *models.ForwardMessageRequest) (string, error) {
	session, exists := s.GetSession(sessionID)
	if !exists {
		return "", models.NewNotFoundError("session not found")
	}

	// Check if session is connected
	if !session.Connected {
		return "", models.NewServiceUnavailableError("session is not connected. Please connect the session first")
	}

	// Check if session is logged in
	if !session.LoggedIn {
		return "", models.NewUnauthorizedError("session is not authenticated. Please scan QR code to login")
	}

	// Validate message text
	if req.Text == "" {
		return "", fmt.Errorf("message text is required for forwarding")
	}

	// Format recipient JID
	recipientJID := req.To
	if !strings.Contains(req.To, "@") {
		phoneNumber := strings.ReplaceAll(req.To, "+", "")
		phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")
		phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")
		
		if len(phoneNumber) < 8 || len(phoneNumber) > 15 {
			return "", fmt.Errorf("invalid phone number length. Should be 8-15 digits")
		}
		
		recipientJID = phoneNumber + "@s.whatsapp.net"
	}

	s.logger.Debug("Forwarding message %s to JID: %s", req.MessageID, recipientJID)

	// Parse recipient JID
	jid, err := types.ParseJID(recipientJID)
	if err != nil {
		return "", fmt.Errorf("invalid recipient JID: %v", err)
	}

	// Convert to non-device JID for message sending
	jid = jid.ToNonAD()

	// Create forward message with context info
	msg := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: proto.String(req.Text),
			ContextInfo: &waProto.ContextInfo{
				StanzaID:       proto.String(req.MessageID),
				Participant:    proto.String(recipientJID),
				IsForwarded:    proto.Bool(true),
				ForwardingScore: proto.Uint32(1),
			},
		},
	}

	// Send the forward message
	resp, err := session.Client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		return "", fmt.Errorf("failed to forward message: %v", err)
	}

	s.logger.Info("Message forwarded to %s from session %s", recipientJID, sessionID)
	return resp.ID, nil
}

// ReplyMessage sends a reply to a specific message
func (s *WhatsAppService) ReplyMessage(sessionID string, req *models.ReplyMessageRequest) (string, error) {
	session, exists := s.GetSession(sessionID)
	if !exists {
		return "", models.NewNotFoundError("session not found")
	}

	// Check if session is connected
	if !session.Connected {
		return "", models.NewServiceUnavailableError("session is not connected. Please connect the session first")
	}

	// Check if session is logged in
	if !session.LoggedIn {
		return "", models.NewUnauthorizedError("session is not authenticated. Please scan QR code to login")
	}

	// Format recipient JID
	recipientJID := req.To
	if !strings.Contains(req.To, "@") {
		phoneNumber := strings.ReplaceAll(req.To, "+", "")
		phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")
		phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")
		
		if len(phoneNumber) < 8 || len(phoneNumber) > 15 {
			return "", fmt.Errorf("invalid phone number length. Should be 8-15 digits")
		}
		
		recipientJID = phoneNumber + "@s.whatsapp.net"
	}

	s.logger.Debug("Replying to message %s with JID: %s", req.QuotedMessageID, recipientJID)

	// Parse recipient JID
	jid, err := types.ParseJID(recipientJID)
	if err != nil {
		return "", fmt.Errorf("invalid recipient JID: %v", err)
	}

	// Convert to non-device JID for message sending
	jid = jid.ToNonAD()

	// Create reply message with context info
	msg := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: proto.String(req.Message),
			ContextInfo: &waProto.ContextInfo{
				StanzaID:      proto.String(req.QuotedMessageID),
				Participant:   proto.String(recipientJID),
				QuotedMessage: &waProto.Message{},
			},
		},
	}

	// Send the reply message
	resp, err := session.Client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		return "", fmt.Errorf("failed to send reply message: %v", err)
	}

	s.logger.Info("Reply sent to %s from session %s", recipientJID, sessionID)
	return resp.ID, nil
}

// SendLocation sends a location message
func (s *WhatsAppService) SendLocation(sessionID string, req *models.SendLocationRequest) (string, error) {
	session, exists := s.GetSession(sessionID)
	if !exists {
		return "", models.NewNotFoundError("session not found")
	}

	// Check if session is connected
	if !session.Connected {
		return "", models.NewServiceUnavailableError("session is not connected. Please connect the session first")
	}

	// Check if session is logged in
	if !session.LoggedIn {
		return "", models.NewUnauthorizedError("session is not authenticated. Please scan QR code to login")
	}

	// Format recipient JID like original implementation
	recipientJID := req.To

	// Check if it's a phone number (contains only digits and +)
	if !strings.Contains(req.To, "@") {
		phoneNumber := strings.ReplaceAll(req.To, "+", "")
		phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")
		phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")

		if len(phoneNumber) < 8 || len(phoneNumber) > 15 {
			return "", fmt.Errorf("invalid phone number length. Should be 8-15 digits")
		}

		// Add @s.whatsapp.net if not present
		recipientJID = phoneNumber + "@s.whatsapp.net"
	}

	s.logger.Debug("Formatted recipient JID for location: %s", recipientJID)

	// Parse recipient JID
	jid, err := types.ParseJID(recipientJID)
	if err != nil {
		return "", fmt.Errorf("invalid recipient JID: %v", err)
	}

	// Convert to non-device JID for message sending
	jid = jid.ToNonAD()

	// Create location message
	msg := &waProto.Message{
		LocationMessage: &waProto.LocationMessage{
			DegreesLatitude:  proto.Float64(req.Latitude),
			DegreesLongitude: proto.Float64(req.Longitude),
		},
	}

	// Send location message
	resp, err := session.Client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		return "", fmt.Errorf("failed to send location: %v", err)
	}

	s.logger.Info("Location sent to %s from session %s", recipientJID, sessionID)
	return resp.ID, nil
}

// SendAttachment sends a file attachment
func (s *WhatsAppService) SendAttachment(sessionID string, req *models.SendFileRequest) (string, error) {
	session, exists := s.GetSession(sessionID)
	if !exists {
		return "", models.NewNotFoundError("session not found")
	}

	if !session.Connected {
		return "", models.NewServiceUnavailableError("session is not connected")
	}

	if !session.LoggedIn {
		return "", models.NewUnauthorizedError("session is not authenticated")
	}

	// Format recipient JID like original implementation
	recipientJID := req.To

	// Check if it's a phone number (contains only digits and +)
	if !strings.Contains(req.To, "@") {
		phoneNumber := strings.ReplaceAll(req.To, "+", "")
		phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")
		phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")

		if len(phoneNumber) < 8 || len(phoneNumber) > 15 {
			return "", fmt.Errorf("invalid phone number length. Should be 8-15 digits")
		}

		// Add @s.whatsapp.net if not present
		recipientJID = phoneNumber + "@s.whatsapp.net"
	}

	// Parse recipient JID
	jid, err := types.ParseJID(recipientJID)
	if err != nil {
		return "", fmt.Errorf("invalid recipient JID: %v", err)
	}

	// Convert to non-device JID for message sending
	jid = jid.ToNonAD()

	// Decode base64 file
	fileData, err := base64.StdEncoding.DecodeString(req.File)
	if err != nil {
		return "", fmt.Errorf("invalid base64 file data: %v", err)
	}

	// Upload file
	uploaded, err := session.Client.Upload(context.Background(), fileData, whatsmeow.MediaDocument)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %v", err)
	}

	// Send file message
	msg := &waProto.Message{
		DocumentMessage: &waProto.DocumentMessage{
			URL:           proto.String(uploaded.URL),
			Mimetype:      proto.String(http.DetectContentType(fileData)),
			Title:         proto.String(req.FileName),
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			MediaKey:      uploaded.MediaKey,
			FileName:      proto.String(req.FileName),
			FileEncSHA256: uploaded.FileEncSHA256,
			DirectPath:    proto.String(uploaded.DirectPath),
		},
	}

	if req.Caption != "" {
		msg.DocumentMessage.Caption = proto.String(req.Caption)
	}

	resp, err := session.Client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		return "", fmt.Errorf("failed to send attachment: %v", err)
	}

	return resp.ID, nil
}

// SendFileFromURL downloads a file from URL and sends it
func (s *WhatsAppService) SendFileFromURL(sessionID string, req *models.SendFileURLRequest) (string, error) {
	session, exists := s.GetSession(sessionID)
	if !exists {
		return "", models.NewNotFoundError("session not found")
	}

	if !session.Connected {
		return "", models.NewServiceUnavailableError("session is not connected")
	}

	if !session.LoggedIn {
		return "", models.NewUnauthorizedError("session is not authenticated")
	}

	// Format recipient JID
	recipientJID := req.To
	if !strings.Contains(req.To, "@") {
		phoneNumber := strings.ReplaceAll(req.To, "+", "")
		phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")
		phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")

		if len(phoneNumber) < 8 || len(phoneNumber) > 15 {
			return "", fmt.Errorf("invalid phone number length. Should be 8-15 digits")
		}

		recipientJID = phoneNumber + "@s.whatsapp.net"
	}

	jid, err := types.ParseJID(recipientJID)
	if err != nil {
		return "", fmt.Errorf("invalid recipient JID: %v", err)
	}

	// Convert to non-device JID for message sending
	jid = jid.ToNonAD()

	// Download file from URL
	fileData, contentType, filename, err := s.downloadFile(req.URL)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %v", err)
	}

	// Use provided filename or extract from URL
	if req.FileName != "" {
		filename = req.FileName
	}

	// Determine media type
	mediaType := s.getMediaType(contentType, req.Type)

	// Send based on media type
	return s.sendMediaByType(session, jid, fileData, contentType, filename, req.Caption, mediaType)
}

// SendImage sends an image (enhanced version)
func (s *WhatsAppService) SendImage(sessionID string, req *models.SendImageRequest) (string, error) {
	session, exists := s.GetSession(sessionID)
	if !exists {
		return "", models.NewNotFoundError("session not found")
	}

	if !session.Connected {
		return "", models.NewServiceUnavailableError("session is not connected")
	}

	if !session.LoggedIn {
		return "", models.NewUnauthorizedError("session is not authenticated")
	}

	// Format recipient JID
	recipientJID := req.To
	if !strings.Contains(req.To, "@") {
		phoneNumber := strings.ReplaceAll(req.To, "+", "")
		phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")
		phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")

		if len(phoneNumber) < 8 || len(phoneNumber) > 15 {
			return "", fmt.Errorf("invalid phone number length. Should be 8-15 digits")
		}

		recipientJID = phoneNumber + "@s.whatsapp.net"
	}

	jid, err := types.ParseJID(recipientJID)
	if err != nil {
		return "", fmt.Errorf("invalid recipient JID: %v", err)
	}

	// Convert to non-device JID for message sending
	jid = jid.ToNonAD()

	// Decode base64 image
	imageData, err := base64.StdEncoding.DecodeString(req.Image)
	if err != nil {
		return "", fmt.Errorf("invalid base64 image data: %v", err)
	}

	// Upload image
	uploaded, err := session.Client.Upload(context.Background(), imageData, whatsmeow.MediaImage)
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %v", err)
	}

	// Create image message
	msg := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			URL:           proto.String(uploaded.URL),
			Mimetype:      proto.String(http.DetectContentType(imageData)),
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			DirectPath:    proto.String(uploaded.DirectPath),
		},
	}

	if req.Caption != "" {
		msg.ImageMessage.Caption = proto.String(req.Caption)
	}

	resp, err := session.Client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		return "", fmt.Errorf("failed to send image: %v", err)
	}

	return resp.ID, nil
}

// downloadFile downloads a file from URL and saves it locally
func (s *WhatsAppService) downloadFile(url string) ([]byte, string, string, error) {
	// Create downloads directory if it doesn't exist
	downloadsDir := "./downloads"
	if err := os.MkdirAll(downloadsDir, 0755); err != nil {
		return nil, "", "", fmt.Errorf("failed to create downloads directory: %v", err)
	}

	// Download file
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Read file data
	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to read file data: %v", err)
	}

	// Get content type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(fileData)
	}

	// Extract filename from URL or Content-Disposition
	filename := s.extractFilename(url, resp.Header.Get("Content-Disposition"))

	// Save file locally
	localPath := filepath.Join(downloadsDir, filename)
	if err := os.WriteFile(localPath, fileData, 0644); err != nil {
		s.logger.Error("Failed to save file locally: %v", err)
		// Continue even if local save fails
	} else {
		s.logger.Info("File saved locally: %s", localPath)
	}

	return fileData, contentType, filename, nil
}

// extractFilename extracts filename from URL or Content-Disposition header
func (s *WhatsAppService) extractFilename(url, contentDisposition string) string {
	// Try to get filename from Content-Disposition header
	if contentDisposition != "" {
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err == nil {
			if filename, ok := params["filename"]; ok {
				return filename
			}
		}
	}

	// Extract from URL
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		filename := parts[len(parts)-1]
		// Remove query parameters
		if idx := strings.Index(filename, "?"); idx != -1 {
			filename = filename[:idx]
		}
		if filename != "" {
			return filename
		}
	}

	// Generate timestamp-based filename
	return fmt.Sprintf("file_%d", time.Now().Unix())
}

// getMediaType determines the media type based on content type
func (s *WhatsAppService) getMediaType(contentType, requestedType string) string {
	if requestedType != "" {
		return requestedType
	}

	switch {
	case strings.HasPrefix(contentType, "image/"):
		return "image"
	case strings.HasPrefix(contentType, "video/"):
		return "video"
	case strings.HasPrefix(contentType, "audio/"):
		return "audio"
	default:
		return "document"
	}
}

// sendMediaByType sends media based on the determined type
func (s *WhatsAppService) sendMediaByType(session *models.Session, jid types.JID, fileData []byte, contentType, filename, caption, mediaType string) (string, error) {
	ctx := context.Background()

	switch mediaType {
	case "image":
		uploaded, err := session.Client.Upload(ctx, fileData, whatsmeow.MediaImage)
		if err != nil {
			return "", fmt.Errorf("failed to upload image: %v", err)
		}

		msg := &waProto.Message{
			ImageMessage: &waProto.ImageMessage{
				URL:           proto.String(uploaded.URL),
				Mimetype:      proto.String(contentType),
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    proto.Uint64(uploaded.FileLength),
				MediaKey:      uploaded.MediaKey,
				FileEncSHA256: uploaded.FileEncSHA256,
				DirectPath:    proto.String(uploaded.DirectPath),
			},
		}

		if caption != "" {
			msg.ImageMessage.Caption = proto.String(caption)
		}

		resp, err := session.Client.SendMessage(ctx, jid, msg)
		if err != nil {
			return "", fmt.Errorf("failed to send image: %v", err)
		}
		return resp.ID, nil

	case "video":
		uploaded, err := session.Client.Upload(ctx, fileData, whatsmeow.MediaVideo)
		if err != nil {
			return "", fmt.Errorf("failed to upload video: %v", err)
		}

		msg := &waProto.Message{
			VideoMessage: &waProto.VideoMessage{
				URL:           proto.String(uploaded.URL),
				Mimetype:      proto.String(contentType),
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    proto.Uint64(uploaded.FileLength),
				MediaKey:      uploaded.MediaKey,
				FileEncSHA256: uploaded.FileEncSHA256,
				DirectPath:    proto.String(uploaded.DirectPath),
			},
		}

		if caption != "" {
			msg.VideoMessage.Caption = proto.String(caption)
		}

		resp, err := session.Client.SendMessage(ctx, jid, msg)
		if err != nil {
			return "", fmt.Errorf("failed to send video: %v", err)
		}
		return resp.ID, nil

	case "audio":
		uploaded, err := session.Client.Upload(ctx, fileData, whatsmeow.MediaAudio)
		if err != nil {
			return "", fmt.Errorf("failed to upload audio: %v", err)
		}

		msg := &waProto.Message{
			AudioMessage: &waProto.AudioMessage{
				URL:           proto.String(uploaded.URL),
				Mimetype:      proto.String(contentType),
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    proto.Uint64(uploaded.FileLength),
				MediaKey:      uploaded.MediaKey,
				FileEncSHA256: uploaded.FileEncSHA256,
				DirectPath:    proto.String(uploaded.DirectPath),
			},
		}

		resp, err := session.Client.SendMessage(ctx, jid, msg)
		if err != nil {
			return "", fmt.Errorf("failed to send audio: %v", err)
		}
		return resp.ID, nil

	default: // document
		uploaded, err := session.Client.Upload(ctx, fileData, whatsmeow.MediaDocument)
		if err != nil {
			return "", fmt.Errorf("failed to upload document: %v", err)
		}

		msg := &waProto.Message{
			DocumentMessage: &waProto.DocumentMessage{
				URL:           proto.String(uploaded.URL),
				Mimetype:      proto.String(contentType),
				Title:         proto.String(filename),
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    proto.Uint64(uploaded.FileLength),
				MediaKey:      uploaded.MediaKey,
				FileName:      proto.String(filename),
				FileEncSHA256: uploaded.FileEncSHA256,
				DirectPath:    proto.String(uploaded.DirectPath),
			},
		}

		if caption != "" {
			msg.DocumentMessage.Caption = proto.String(caption)
		}

		resp, err := session.Client.SendMessage(ctx, jid, msg)
		if err != nil {
			return "", fmt.Errorf("failed to send document: %v", err)
		}
		return resp.ID, nil
	}
}

// CheckNumber checks if a number is registered on WhatsApp
func (s *WhatsAppService) CheckNumber(sessionID string, number string) (bool, string, error) {
	session, exists := s.GetSession(sessionID)
	if !exists {
		return false, "", models.NewNotFoundError("session not found")
	}

	if !session.Connected {
		return false, "", models.NewServiceUnavailableError("session is not connected")
	}

	if !session.LoggedIn {
		return false, "", models.NewUnauthorizedError("session is not authenticated")
	}

	// Parse number
	jid, err := types.ParseJID(number + "@s.whatsapp.net")
	if err != nil {
		return false, "", fmt.Errorf("invalid number format: %v", err)
	}

	// Check if number exists
	resp, err := session.Client.IsOnWhatsApp(context.Background(), []string{jid.User})
	if err != nil {
		return false, "", fmt.Errorf("failed to check number: %v", err)
	}

	if len(resp) > 0 && resp[0].IsIn {
		return true, resp[0].JID.String(), nil
	}

	return false, "", nil
}

// SendTyping sends typing indicator
func (s *WhatsAppService) SendTyping(sessionID string, to string, typing bool) error {
	session, exists := s.GetSession(sessionID)
	if !exists {
		return models.NewNotFoundError("session not found")
	}

	if !session.Connected {
		return models.NewServiceUnavailableError("session is not connected")
	}

	if !session.LoggedIn {
		return models.NewUnauthorizedError("session is not authenticated")
	}


	// Format recipient JID (same logic as SendMessage)
	recipientJID := to
	if !strings.Contains(to, "@") {
		// Clean phone number
		phoneNumber := strings.ReplaceAll(to, "+", "")
		phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")
		phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")

		if len(phoneNumber) < 8 || len(phoneNumber) > 15 {
			return fmt.Errorf("invalid phone number length. Should be 8-15 digits")
		}

		// Add @s.whatsapp.net if not present
		recipientJID = phoneNumber + "@s.whatsapp.net"
	}

	s.logger.Debug("Formatted recipient JID for typing: %s", recipientJID)

	// Parse recipient JID
	jid, err := types.ParseJID(recipientJID)
	if err != nil {
		return fmt.Errorf("invalid recipient JID: %v", err)
	}

	// Convert to non-device JID for presence/typing indicators
	jid = jid.ToNonAD()

	// Ensure we have a push name (required for presence/typing to work properly)
	if session.Client.Store.PushName == "" {
		pushName := s.generateRandomName()
		session.Client.Store.PushName = pushName
		s.logger.Debug("Set push name '%s' for typing indicator", pushName)
	}

	// CRITICAL: Set online presence first - this is mandatory for typing indicators
	s.logger.Debug("Setting online presence for typing indicator...")
	if err = session.Client.SendPresence(context.Background(), types.PresenceAvailable); err != nil {
		s.logger.Error("Failed to set online presence: %v", err)
		return fmt.Errorf("failed to set online presence: %v", err)
	}
	s.logger.Debug("✅ Online presence set successfully")

	// Subscribe to presence updates for the target contact (helps with reliability)
	if err = session.Client.SubscribePresence(context.Background(), jid); err != nil {
		s.logger.Debug("Failed to subscribe to presence for %s: %v", jid.String(), err)
		// Not critical, continue
	}

	// Longer delay to ensure presence is fully processed by WhatsApp
	time.Sleep(300 * time.Millisecond)

	// Send typing indicator using chat presence
	var presenceType types.ChatPresence
	if typing {
		presenceType = types.ChatPresenceComposing
	} else {
		presenceType = types.ChatPresencePaused
	}

	err = session.Client.SendChatPresence(context.Background(), jid, presenceType, types.ChatPresenceMediaText)
	if err != nil {
		return fmt.Errorf("failed to send typing indicator: %v", err)
	}

	if typing {
		s.logger.Info("Sent typing indicator to %s from session %s (push name: %s)", 
			jid.String(), sessionID, session.Client.Store.PushName)
	} else {
		s.logger.Info("Stopped typing indicator to %s from session %s", jid.String(), sessionID)
	}

	return nil
}

// SetPresence sets the presence status for a session
func (s *WhatsAppService) SetPresence(sessionID string, status string) error {
	session, exists := s.GetSession(sessionID)
	if !exists {
		return models.NewNotFoundError("session not found")
	}

	if !session.Connected {
		return models.NewServiceUnavailableError("session is not connected")
	}

	if !session.LoggedIn {
		return models.NewUnauthorizedError("session is not authenticated")
	}

	var presence types.Presence
	switch strings.ToLower(status) {
	case "available", "online":
		presence = types.PresenceAvailable
	case "unavailable", "offline":
		presence = types.PresenceUnavailable
	default:
		return fmt.Errorf("invalid presence status: %s (valid: available, unavailable)", status)
	}

	err := session.Client.SendPresence(context.Background(), presence)
	if err != nil {
		return fmt.Errorf("failed to set presence: %v", err)
	}

	s.logger.Info("Set presence to %s for session %s", status, sessionID)
	return nil
}

// GetGroups returns all groups for a session
func (s *WhatsAppService) GetGroups(sessionID string) ([]map[string]interface{}, error) {
	session, exists := s.GetSession(sessionID)
	if !exists {
		return nil, models.NewNotFoundError("session not found")
	}

	if !session.Connected {
		return nil, models.NewServiceUnavailableError("session is not connected")
	}

	if !session.LoggedIn {
		return nil, models.NewUnauthorizedError("session is not authenticated")
	}

	// Get groups
	groups, err := session.Client.GetJoinedGroups(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get groups: %v", err)
	}

	result := make([]map[string]interface{}, 0, len(groups))
	for _, groupInfo := range groups {
		result = append(result, map[string]interface{}{
			"jid":     groupInfo.JID.String(),
			"name":    groupInfo.Name,
			"topic":   groupInfo.Topic,
			"owner":   groupInfo.OwnerJID.String(),
			"created": groupInfo.GroupCreated.Unix(),
		})
	}

	return result, nil
}

// sendWebhook sends incoming message data to configured webhook URL
func (s *WhatsAppService) sendWebhook(session *models.Session, evt *events.Message) {
	// Don't send webhook if session is disabled
	if !session.Enabled {
		s.logger.Debug("Session %s is disabled, skipping webhook", session.ID)
		return
	}

	// Get sender name from push name (most reliable method)
	senderName := "Unknown"
	if evt.Info.PushName != "" {
		senderName = evt.Info.PushName
	} else if evt.Info.Sender.User != "" {
		// Fallback to phone number if no push name
		senderName = evt.Info.Sender.User
	}

	// Create webhook message
	webhookMsg := &models.WebhookMessage{
		SessionID:   session.ID,
		From:        evt.Info.Sender.String(),
		FromName:    senderName,
		To:          session.Client.Store.ID.String(),
		Timestamp:   evt.Info.Timestamp,
		ID:          evt.Info.ID,
		IsGroup:     evt.Info.IsGroup,
		MessageType: "text",
	}

	// Handle group messages
	if evt.Info.IsGroup {
		webhookMsg.GroupID = evt.Info.Chat.String()
	}

	// Extract message content based on type
	if evt.Message.GetConversation() != "" {
		webhookMsg.Message = evt.Message.GetConversation()
		webhookMsg.MessageType = "text"
	} else if evt.Message.GetExtendedTextMessage() != nil {
		webhookMsg.Message = evt.Message.GetExtendedTextMessage().GetText()
		webhookMsg.MessageType = "text"
	} else if evt.Message.GetImageMessage() != nil {
		webhookMsg.Message = evt.Message.GetImageMessage().GetCaption()
		webhookMsg.MessageType = "image"
		// Download and save media file
		if fileName, err := s.downloadIncomingMedia(session, evt); err == nil {
			// Create temporary access URL (valid for 1 hour)
			webhookMsg.MediaURL = fmt.Sprintf("/api/media/temp/%s?expires=%d", fileName, time.Now().Add(time.Hour).Unix())
		}
	} else if evt.Message.GetDocumentMessage() != nil {
		webhookMsg.Message = evt.Message.GetDocumentMessage().GetCaption()
		webhookMsg.MessageType = "document"
		// Download and save media file
		if fileName, err := s.downloadIncomingMedia(session, evt); err == nil {
			// Create temporary access URL (valid for 1 hour)
			webhookMsg.MediaURL = fmt.Sprintf("/api/media/temp/%s?expires=%d", fileName, time.Now().Add(time.Hour).Unix())
		}
	} else if evt.Message.GetAudioMessage() != nil {
		webhookMsg.MessageType = "audio"
		// Download and save media file
		if fileName, err := s.downloadIncomingMedia(session, evt); err == nil {
			// Create temporary access URL (valid for 1 hour)
			webhookMsg.MediaURL = fmt.Sprintf("/api/media/temp/%s?expires=%d", fileName, time.Now().Add(time.Hour).Unix())
		}
	} else if evt.Message.GetVideoMessage() != nil {
		webhookMsg.Message = evt.Message.GetVideoMessage().GetCaption()
		webhookMsg.MessageType = "video"
		// Download and save media file
		if fileName, err := s.downloadIncomingMedia(session, evt); err == nil {
			// Create temporary access URL (valid for 1 hour)
			webhookMsg.MediaURL = fmt.Sprintf("/api/media/temp/%s?expires=%d", fileName, time.Now().Add(time.Hour).Unix())
		}
	} else {
		webhookMsg.Message = "[Unsupported message type]"
		webhookMsg.MessageType = "unknown"
	}

	// Send webhook with retries
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := s.sendWebhookHTTP(session.WebhookURL, webhookMsg); err != nil {
			s.logger.Error("Webhook attempt %d failed for session %s: %v", attempt, session.ID, err)
			if attempt < maxRetries {
				// Exponential backoff
				time.Sleep(time.Duration(attempt*attempt) * time.Second)
			}
		} else {
			s.logger.Info("Webhook sent successfully for session %s", session.ID)
			break
		}
	}
}

// sendAutoReply sends an automatic reply to incoming messages
func (s *WhatsAppService) sendAutoReply(session *models.Session, evt *events.Message) {
	// Don't reply if session is disabled
	if !session.Enabled {
		s.logger.Debug("Session %s is disabled, skipping auto-reply", session.ID)
		return
	}

	// Don't reply to group messages or if no auto reply text is set
	if evt.Info.IsGroup || session.AutoReplyText == nil || *session.AutoReplyText == "" {
		return
	}

	// Don't reply to messages that are already replies or system messages
	if evt.Message.GetExtendedTextMessage() != nil && evt.Message.GetExtendedTextMessage().GetContextInfo() != nil {
		return // Don't reply to replies
	}

	// Create the reply message
	replyMsg := &waProto.Message{
		Conversation: proto.String(*session.AutoReplyText),
	}

	// Convert sender JID to user JID (remove device part)
	userJID := evt.Info.Sender.ToNonAD()

	// Send the auto reply
	_, err := session.Client.SendMessage(context.Background(), userJID, replyMsg)
	if err != nil {
		s.logger.Error("Failed to send auto reply for session %s: %v", session.ID, err)
		return
	}

	s.logger.Info("Auto reply sent to %s in session %s", userJID.User, session.ID)
}

// sendWebhookHTTP sends the webhook message via HTTP POST
func (s *WhatsAppService) sendWebhookHTTP(webhookURL string, msg *models.WebhookMessage) error {
	// Marshal message to JSON
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook message: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "WhatsApp-Multi-Session/1.0")

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// downloadIncomingMedia downloads media from incoming messages and saves locally
func (s *WhatsAppService) downloadIncomingMedia(session *models.Session, evt *events.Message) (string, error) {
	var mediaData []byte
	var fileName string
	var err error
	ctx := context.Background()

	// Create media directory
	mediaDir := "./media/received"
	if err := os.MkdirAll(mediaDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create media directory: %v", err)
	}

	// Generate unique filename
	timestamp := time.Now().Unix()
	sessionID := session.ID
	messageID := evt.Info.ID

	// Download based on message type
	if img := evt.Message.GetImageMessage(); img != nil {
		mediaData, err = session.Client.Download(ctx, img)
		if err != nil {
			return "", fmt.Errorf("failed to download image: %v", err)
		}
		fileName = fmt.Sprintf("%s_%d_%s.jpg", sessionID, timestamp, messageID)
	} else if doc := evt.Message.GetDocumentMessage(); doc != nil {
		mediaData, err = session.Client.Download(ctx, doc)
		if err != nil {
			return "", fmt.Errorf("failed to download document: %v", err)
		}
		// Use original filename if available, otherwise generate one
		if doc.GetFileName() != "" {
			ext := filepath.Ext(doc.GetFileName())
			name := strings.TrimSuffix(doc.GetFileName(), ext)
			fileName = fmt.Sprintf("%s_%d_%s_%s%s", sessionID, timestamp, messageID, name, ext)
		} else {
			fileName = fmt.Sprintf("%s_%d_%s.bin", sessionID, timestamp, messageID)
		}
	} else if video := evt.Message.GetVideoMessage(); video != nil {
		mediaData, err = session.Client.Download(ctx, video)
		if err != nil {
			return "", fmt.Errorf("failed to download video: %v", err)
		}
		fileName = fmt.Sprintf("%s_%d_%s.mp4", sessionID, timestamp, messageID)
	} else if audio := evt.Message.GetAudioMessage(); audio != nil {
		mediaData, err = session.Client.Download(ctx, audio)
		if err != nil {
			return "", fmt.Errorf("failed to download audio: %v", err)
		}
		fileName = fmt.Sprintf("%s_%d_%s.ogg", sessionID, timestamp, messageID)
	} else {
		return "", fmt.Errorf("unsupported media type")
	}

	// Save file locally
	filePath := filepath.Join(mediaDir, fileName)
	if err := os.WriteFile(filePath, mediaData, 0644); err != nil {
		return "", fmt.Errorf("failed to save media file: %v", err)
	}

	s.logger.Info("Downloaded media file: %s", filePath)
	return fileName, nil
}

// Close closes all sessions and the service
func (s *WhatsAppService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, session := range s.sessions {
		if session.Connected {
			session.Client.Disconnect()
		}
	}

	return nil
}

// GetConversations retrieves all conversations/chats for a session
func (s *WhatsAppService) GetConversations(sessionID string) ([]*models.Conversation, error) {
	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	s.mu.RUnlock()

	if !exists {
		return nil, models.ErrSessionNotFound
	}

	if !session.Connected || !session.LoggedIn {
		return nil, models.ErrSessionNotAuthenticated
	}

	// Get all contacts from the store
	contacts, err := session.Client.Store.Contacts.GetAllContacts(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts: %v", err)
	}

	// Get joined groups
	groups, err := session.Client.GetJoinedGroups(context.Background())
	if err != nil {
		s.logger.Warn("Failed to get groups: %v", err)
		// Continue without groups
		groups = []*types.GroupInfo{}
	}

	// Create conversations list
	conversations := make([]*models.Conversation, 0, len(contacts)+len(groups))

	// Add individual chats
	for jid, contact := range contacts {
		if jid.Server != types.DefaultUserServer {
			continue // Skip non-user contacts
		}

		conversation := &models.Conversation{
			JID:     jid.String(),
			Name:    s.getContactName(contact),
			IsGroup: false,
			// Note: WhatsApp doesn't provide unread count, last message, etc. via these APIs
			// These would need to be tracked by listening to events or using history sync
			UnreadCount: 0,
			IsPinned:    false,
			IsMuted:     false,
			IsArchived:  false,
		}
		conversations = append(conversations, conversation)
	}

	// Add group chats
	for _, group := range groups {
		conversation := &models.Conversation{
			JID:         group.JID.String(),
			Name:        group.Name,
			IsGroup:     true,
			UnreadCount: 0,
			IsPinned:    false,
			IsMuted:     false,
			IsArchived:  false,
		}
		conversations = append(conversations, conversation)
	}

	return conversations, nil
}

// getContactName returns the best available name for a contact
func (s *WhatsAppService) getContactName(contact types.ContactInfo) string {
	// Debug log the contact info we received
	s.logger.Debug("Contact info - FullName: '%s', BusinessName: '%s', PushName: '%s'",
		contact.FullName, contact.BusinessName, contact.PushName)

	// Priority: FullName > BusinessName > PushName > "Unknown"
	if contact.FullName != "" {
		return contact.FullName
	}
	if contact.BusinessName != "" {
		return contact.BusinessName
	}
	if contact.PushName != "" {
		return contact.PushName
	}
	return "Unknown"
}

// generateRandomName generates a realistic random name for push name
func (s *WhatsAppService) generateRandomName() string {
	firstNames := []string{
		"Alex", "Sam", "Jordan", "Taylor", "Casey", "Morgan", "Jamie", "Riley",
		"Avery", "Peyton", "Quinn", "Sage", "Rowan", "Emery", "Hayden", "Finley",
		"Cameron", "Drew", "Blake", "Reese", "Parker", "River", "Skylar", "Lane",
		"Kendall", "Harley", "Phoenix", "Dakota", "Charlie", "Frankie", "Kai",
		"Robin", "Eden", "Jules", "Ari", "Reign", "Remy", "Briar", "Sage",
	}

	// Use crypto/rand for better randomization
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(firstNames))))
	if err != nil {
		// Fallback to timestamp-based selection
		return firstNames[time.Now().Unix()%int64(len(firstNames))]
	}

	return firstNames[n.Int64()]
}
