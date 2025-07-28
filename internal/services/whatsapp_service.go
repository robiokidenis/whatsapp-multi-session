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
)

// WhatsAppService manages WhatsApp clients and sessions
type WhatsAppService struct {
	sessions       map[string]*models.Session
	store          *sqlstore.Container
	sessionRepo    *repository.SessionRepository
	logger         *logger.Logger
	mu             sync.RWMutex
	eventHandlers  map[string]func(*events.Message)
}

func init() {
	// Set up WhatsApp logging
	waLog.Stdout("Main", "INFO", true)

	// Set device properties to avoid client outdated error
	store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_CHROME.Enum()
	store.DeviceProps.Os = proto.String("Windows")
	store.DeviceProps.RequireFullSync = proto.Bool(false)
}

// NewWhatsAppService creates a new WhatsApp service
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

	// Create WhatsApp store with foreign keys enabled
	waLogger := waLog.Stdout("Store", "INFO", true)
	connectionString := fmt.Sprintf("file:%s?_foreign_keys=on", dbPath)
	container, err := sqlstore.New(context.Background(), "sqlite3", connectionString, waLogger)
	if err != nil {
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
		currentSessionCount := len(s.sessions)
		// This would need to be enhanced to get user's actual session limit from database
		// For now, we'll use a default limit
		sessionLimit := 5 // This should come from user's record
		
		if currentSessionCount >= sessionLimit {
			return nil, fmt.Errorf("session limit reached. You can create maximum %d sessions", sessionLimit)
		}
	}

	// Create device store
	deviceStore := s.store.NewDevice()
	
	// Create WhatsApp client
	clientLog := waLog.Stdout("Client", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	// Create session
	session := &models.Session{
		ID:         sessionID,
		Phone:      phoneForDisplay,
		Name:       req.Name,
		Position:   req.Position,
		WebhookURL: req.WebhookURL,
		Client:     client,
		Connected:  false,
		LoggedIn:   false,
		Connecting: false,
	}

	// Set up event handlers
	s.setupEventHandlers(session)

	// Store in memory
	s.sessions[sessionID] = session

	// Save to database
	metadata := &models.SessionMetadata{
		ID:         sessionID,
		Phone:      phoneForDisplay,
		Name:       req.Name,
		Position:   req.Position,
		WebhookURL: req.WebhookURL,
		CreatedAt:  time.Now(),
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

// ConnectSession initiates connection for a session
func (s *WhatsAppService) ConnectSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	if session.Connecting {
		return fmt.Errorf("session %s is already connecting", sessionID)
	}

	if session.Connected {
		return fmt.Errorf("session %s is already connected", sessionID)
	}

	session.Connecting = true

	go func() {
		defer func() {
			s.mu.Lock()
			session.Connecting = false
			s.mu.Unlock()
		}()

		if err := session.Client.Connect(); err != nil {
			s.logger.Error("Failed to connect session %s: %v", sessionID, err)
			return
		}

		s.mu.Lock()
		session.Connected = true
		s.mu.Unlock()

		s.logger.Info("Session %s connected successfully", sessionID)
	}()

	return nil
}

// DisconnectSession disconnects a session
func (s *WhatsAppService) DisconnectSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
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
		return fmt.Errorf("session %s not found", sessionID)
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
		return fmt.Errorf("session %s not found", sessionID)
	}

	if session.LoggedIn {
		return fmt.Errorf("session %s is already logged in", sessionID)
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
		return fmt.Errorf("session %s not found", sessionID)
	}

	if !session.LoggedIn {
		return fmt.Errorf("session %s is not logged in", sessionID)
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
		return "", fmt.Errorf("session %s not found", sessionID)
	}

	if session.LoggedIn {
		return "", fmt.Errorf("session %s is already logged in", sessionID)
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
		return fmt.Errorf("session %s not found", sessionID)
	}

	// Update in-memory session
	if req.Name != "" {
		session.Name = req.Name
	}
	if req.WebhookURL != "" {
		session.WebhookURL = req.WebhookURL
	}

	// Update in database
	metadata := &models.SessionMetadata{
		ID:         sessionID,
		Phone:      session.Phone,
		ActualPhone: session.ActualPhone,
		Name:       session.Name,
		Position:   req.Position,
		WebhookURL: session.WebhookURL,
	}

	return s.sessionRepo.Update(metadata)
}

// setupEventHandlers sets up event handlers for a session
func (s *WhatsAppService) setupEventHandlers(session *models.Session) {
	session.Client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Connected:
			s.mu.Lock()
			session.Connected = true
			session.LoggedIn = session.Client.Store.ID != nil
			s.mu.Unlock()
			
			s.logger.Info("Session %s connected", session.ID)

		case *events.Disconnected:
			s.mu.Lock()
			session.Connected = false
			s.mu.Unlock()
			
			s.logger.Info("Session %s disconnected", session.ID)

		case *events.LoggedOut:
			s.mu.Lock()
			session.LoggedIn = false
			s.mu.Unlock()
			
			s.logger.Info("Session %s logged out", session.ID)

		case *events.Message:
			// Handle incoming messages
			if handler, exists := s.eventHandlers[session.ID]; exists {
				handler(v)
			}
			
			// Send webhook if configured
			if session.WebhookURL != "" {
				go s.sendWebhook(session, v)
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
		// Create device store
		deviceStore := s.store.NewDevice()
		
		// Create WhatsApp client
		clientLog := waLog.Stdout("Client", "INFO", true)
		client := whatsmeow.NewClient(deviceStore, clientLog)

		// Create session
		session := &models.Session{
			ID:         metadata.ID,
			Phone:      metadata.Phone,
			ActualPhone: metadata.ActualPhone,
			Name:       metadata.Name,
			Position:   metadata.Position,
			WebhookURL: metadata.WebhookURL,
			Client:     client,
			Connected:  false,
			LoggedIn:   false,
			Connecting: false,
		}

		// Set up event handlers
		s.setupEventHandlers(session)

		// Store in memory
		s.sessions[metadata.ID] = session
	}

	s.logger.Info("Loaded %d existing sessions", len(metadatas))
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
		return "", fmt.Errorf("session %s not found", sessionID)
	}

	if !session.LoggedIn {
		return "", fmt.Errorf("session %s is not logged in", sessionID)
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

// SendAttachment sends a file attachment
func (s *WhatsAppService) SendAttachment(sessionID string, req *models.SendFileRequest) (string, error) {
	session, exists := s.GetSession(sessionID)
	if !exists {
		return "", fmt.Errorf("session %s not found", sessionID)
	}

	if !session.LoggedIn {
		return "", fmt.Errorf("session %s is not logged in", sessionID)
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
		return "", fmt.Errorf("session %s not found", sessionID)
	}

	if !session.LoggedIn {
		return "", fmt.Errorf("session %s is not logged in", sessionID)
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
		return "", fmt.Errorf("session %s not found", sessionID)
	}

	if !session.LoggedIn {
		return "", fmt.Errorf("session %s is not logged in", sessionID)
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
		return false, "", fmt.Errorf("session %s not found", sessionID)
	}

	if !session.LoggedIn {
		return false, "", fmt.Errorf("session %s is not logged in", sessionID)
	}

	// Parse number
	jid, err := types.ParseJID(number + "@s.whatsapp.net")
	if err != nil {
		return false, "", fmt.Errorf("invalid number format: %v", err)
	}

	// Check if number exists
	resp, err := session.Client.IsOnWhatsApp([]string{jid.User})
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
		return fmt.Errorf("session %s not found", sessionID)
	}

	if !session.LoggedIn {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	// Parse recipient JID
	jid, err := types.ParseJID(to)
	if err != nil {
		return fmt.Errorf("invalid recipient JID: %v", err)
	}

	// Send typing indicator using chat presence
	if typing {
		return session.Client.SendChatPresence(jid, types.ChatPresenceComposing, types.ChatPresenceMediaText)
	} else {
		return session.Client.SendChatPresence(jid, types.ChatPresencePaused, types.ChatPresenceMediaText)
	}
}

// GetGroups returns all groups for a session
func (s *WhatsAppService) GetGroups(sessionID string) ([]map[string]interface{}, error) {
	session, exists := s.GetSession(sessionID)
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	if !session.LoggedIn {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	// Get groups
	groups, err := session.Client.GetJoinedGroups()
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
	// Create webhook message
	webhookMsg := &models.WebhookMessage{
		SessionID:   session.ID,
		From:        evt.Info.Sender.String(),
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
		// TODO: Add media download URL
	} else if evt.Message.GetDocumentMessage() != nil {
		webhookMsg.Message = evt.Message.GetDocumentMessage().GetCaption()
		webhookMsg.MessageType = "document"
		// TODO: Add media download URL
	} else if evt.Message.GetAudioMessage() != nil {
		webhookMsg.MessageType = "audio"
		// TODO: Add media download URL
	} else if evt.Message.GetVideoMessage() != nil {
		webhookMsg.Message = evt.Message.GetVideoMessage().GetCaption()
		webhookMsg.MessageType = "video"
		// TODO: Add media download URL
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