package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
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
	connectionString := fmt.Sprintf("%s?_foreign_keys=on", dbPath)
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
func (s *WhatsAppService) CreateSession(req *models.CreateSessionRequest) (*models.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate session ID if phone not provided
	sessionID := req.Phone
	if sessionID == "" {
		sessionID = s.generateSessionID()
	}

	// Check if session already exists
	if _, exists := s.sessions[sessionID]; exists {
		return nil, fmt.Errorf("session %s already exists", sessionID)
	}

	// Create device store
	deviceStore := s.store.NewDevice()
	
	// Create WhatsApp client
	clientLog := waLog.Stdout("Client", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	// Create session
	session := &models.Session{
		ID:         sessionID,
		Phone:      sessionID,
		Name:       req.Name,
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
		Phone:      sessionID,
		Name:       req.Name,
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
	return fmt.Sprintf("session_%d", time.Now().UnixNano()/1000000)
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

	// Parse recipient JID
	jid, err := types.ParseJID(req.To)
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

	// Parse recipient JID
	jid, err := types.ParseJID(req.To)
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