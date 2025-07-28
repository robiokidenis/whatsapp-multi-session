package handlers

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.mau.fi/whatsmeow"

	"whatsapp-multi-session/internal/models"
	"whatsapp-multi-session/internal/services"
	"whatsapp-multi-session/pkg/logger"
)

// SessionHandler handles session-related endpoints
type SessionHandler struct {
	whatsappService *services.WhatsAppService
	logger          *logger.Logger
	jwtSecret       string
	upgrader        websocket.Upgrader
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(
	whatsappService *services.WhatsAppService,
	jwtSecret string,
	log *logger.Logger,
) *SessionHandler {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// Allow connections from localhost and development origins
			origin := r.Header.Get("Origin")
			allowed := []string{
				"http://localhost:3000",
				"http://127.0.0.1:3000",
				"http://localhost:8080",
				"http://127.0.0.1:8080",
			}
			for _, allow := range allowed {
				if origin == allow {
					return true
				}
			}
			// Also allow if no origin (direct connection)
			return origin == ""
		},
	}

	return &SessionHandler{
		whatsappService: whatsappService,
		logger:          log,
		jwtSecret:       jwtSecret,
		upgrader:        upgrader,
	}
}

// CreateSession handles session creation
func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Auto-generate name if empty
	if req.Name == "" {
		req.Name = "Session-" + generateSessionID()
	}

	// Get user info from context
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "User authentication required", http.StatusUnauthorized)
		return
	}

	role, ok := r.Context().Value("role").(string)
	if !ok {
		http.Error(w, "User role required", http.StatusUnauthorized)
		return
	}

	// Create session
	session, err := h.whatsappService.CreateSession(&req, userID, role)
	if err != nil {
		h.logger.Error("Failed to create session: %v", err)
		HandleErrorWithMessage(w, http.StatusInternalServerError, err.Error(), models.ErrCodeInternalServer)
		return
	}

	// Convert to response
	response := &models.SessionResponse{
		ID:          session.ID,
		Phone:       session.Phone,
		ActualPhone: session.ActualPhone,
		Name:        session.Name,
		Position:    session.Position,
		WebhookURL:  session.WebhookURL,
		Connected:   session.Connected,
		LoggedIn:    session.LoggedIn,
	}

	WriteSuccessResponse(w, "Session created successfully", response)
}

// GetSessions handles getting all sessions
func (h *SessionHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
	sessions := h.whatsappService.GetAllSessions()

	// Convert to response format
	responses := make([]*models.SessionResponse, len(sessions))
	for i, session := range sessions {
		responses[i] = &models.SessionResponse{
			ID:          session.ID,
			Phone:       session.Phone,
			ActualPhone: session.ActualPhone,
			Name:        session.Name,
			Position:    session.Position,
			WebhookURL:  session.WebhookURL,
			Connected:   session.Connected,
			LoggedIn:    session.LoggedIn,
		}
	}

	WriteSuccessResponse(w, "Sessions retrieved successfully", responses)
}

// GetSession handles getting a specific session
func (h *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	session, exists := h.whatsappService.GetSession(sessionID)
	if !exists {
		HandleErrorWithMessage(w, http.StatusNotFound, "Session not found", models.ErrCodeNotFound)
		return
	}

	response := &models.SessionResponse{
		ID:          session.ID,
		Phone:       session.Phone,
		ActualPhone: session.ActualPhone,
		Name:        session.Name,
		Position:    session.Position,
		WebhookURL:  session.WebhookURL,
		Connected:   session.Connected,
		LoggedIn:    session.LoggedIn,
	}

	WriteSuccessResponse(w, "Session retrieved successfully", response)
}

// ConnectSession handles session connection
func (h *SessionHandler) ConnectSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	if err := h.whatsappService.ConnectSession(sessionID); err != nil {
		h.logger.Error("Failed to connect session %s: %v", sessionID, err)
		HandleError(w, err)
		return
	}

	WriteSuccessResponse(w, "Connection initiated", nil)
}

// DisconnectSession handles session disconnection
func (h *SessionHandler) DisconnectSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	if err := h.whatsappService.DisconnectSession(sessionID); err != nil {
		h.logger.Error("Failed to disconnect session %s: %v", sessionID, err)
		HandleError(w, err)
		return
	}

	WriteSuccessResponse(w, "Session disconnected", nil)
}

// DeleteSession handles session deletion
func (h *SessionHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	if err := h.whatsappService.DeleteSession(sessionID); err != nil {
		h.logger.Error("Failed to delete session %s: %v", sessionID, err)
		HandleError(w, err)
		return
	}

	WriteSuccessResponse(w, "Session deleted successfully", nil)
}

// GetQRCode handles QR code generation
func (h *SessionHandler) GetQRCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	qrCode, err := h.whatsappService.GetQRCode(sessionID)
	if err != nil {
		h.logger.Error("Failed to get QR code for session %s: %v", sessionID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := &models.QRResponse{
		QRCode: qrCode,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateSession handles session updates
func (h *SessionHandler) UpdateSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	var req models.UpdateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.whatsappService.UpdateSession(sessionID, &req); err != nil {
		h.logger.Error("Failed to update session %s: %v", sessionID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Session updated"})
}

// SendMessage handles sending text messages
func (h *SessionHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	var req models.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.To == "" || req.Message == "" {
		http.Error(w, "To and message fields are required", http.StatusBadRequest)
		return
	}

	// Send message
	messageID, err := h.whatsappService.SendMessage(sessionID, &req)
	if err != nil {
		h.logger.Error("Failed to send message from session %s: %v", sessionID, err)
		HandleError(w, err)
		return
	}

	WriteSuccessResponse(w, "Message sent successfully", map[string]interface{}{
		"message_id": messageID,
	})
}

// SendLocation handles sending location messages
func (h *SessionHandler) SendLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	var req models.SendLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.To == "" {
		http.Error(w, "To field is required", http.StatusBadRequest)
		return
	}
	
	// Validate latitude and longitude ranges
	if req.Latitude < -90 || req.Latitude > 90 {
		http.Error(w, "Latitude must be between -90 and 90", http.StatusBadRequest)
		return
	}
	
	if req.Longitude < -180 || req.Longitude > 180 {
		http.Error(w, "Longitude must be between -180 and 180", http.StatusBadRequest)
		return
	}

	// Send location
	messageID, err := h.whatsappService.SendLocation(sessionID, &req)
	if err != nil {
		h.logger.Error("Failed to send location from session %s: %v", sessionID, err)
		HandleError(w, err)
		return
	}

	WriteSuccessResponse(w, "Location sent successfully", map[string]interface{}{
		"message_id": messageID,
	})
}

// SendAttachment handles sending file attachments
func (h *SessionHandler) SendAttachment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	var req models.SendFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.To == "" || req.File == "" {
		http.Error(w, "To and file fields are required", http.StatusBadRequest)
		return
	}

	// Send attachment
	messageID, err := h.whatsappService.SendAttachment(sessionID, &req)
	if err != nil {
		h.logger.Error("Failed to send attachment from session %s: %v", sessionID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      messageID,
		"message": "Attachment sent successfully",
	})
}

// SendFileFromURL handles sending files from URL
func (h *SessionHandler) SendFileFromURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	var req models.SendFileURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.To == "" || req.URL == "" {
		http.Error(w, "To and URL fields are required", http.StatusBadRequest)
		return
	}

	// Send file from URL
	messageID, err := h.whatsappService.SendFileFromURL(sessionID, &req)
	if err != nil {
		h.logger.Error("Failed to send file from URL for session %s: %v", sessionID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      messageID,
		"message": "File sent successfully",
	})
}

// SendImage handles sending images
func (h *SessionHandler) SendImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	var req models.SendImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.To == "" || req.Image == "" {
		http.Error(w, "To and image fields are required", http.StatusBadRequest)
		return
	}

	// Send image
	messageID, err := h.whatsappService.SendImage(sessionID, &req)
	if err != nil {
		h.logger.Error("Failed to send image from session %s: %v", sessionID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      messageID,
		"message": "Image sent successfully",
	})
}

// CheckNumber handles checking if a number is on WhatsApp
func (h *SessionHandler) CheckNumber(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	var req struct {
		Number string `json:"number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Number == "" {
		http.Error(w, "Number field is required", http.StatusBadRequest)
		return
	}

	// Check number
	exists, jid, err := h.whatsappService.CheckNumber(sessionID, req.Number)
	if err != nil {
		h.logger.Error("Failed to check number from session %s: %v", sessionID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"exists": exists,
		"number": req.Number,
	}
	if exists {
		response["jid"] = jid
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SendTyping handles sending typing indicator
func (h *SessionHandler) SendTyping(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	var req struct {
		To string `json:"to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.To == "" {
		http.Error(w, "To field is required", http.StatusBadRequest)
		return
	}

	if err := h.whatsappService.SendTyping(sessionID, req.To, true); err != nil {
		h.logger.Error("Failed to send typing indicator from session %s: %v", sessionID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Typing indicator sent"})
}

// StopTyping handles stopping typing indicator
func (h *SessionHandler) StopTyping(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	var req struct {
		To string `json:"to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.To == "" {
		http.Error(w, "To field is required", http.StatusBadRequest)
		return
	}

	if err := h.whatsappService.SendTyping(sessionID, req.To, false); err != nil {
		h.logger.Error("Failed to stop typing indicator from session %s: %v", sessionID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Typing indicator stopped"})
}

// GetGroups handles getting groups for a session
func (h *SessionHandler) GetGroups(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	groups, err := h.whatsappService.GetGroups(sessionID)
	if err != nil {
		h.logger.Error("Failed to get groups from session %s: %v", sessionID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    groups,
	})
}

// SendMessageGeneral handles sending messages via API with phone selection (for compatibility)
func (h *SessionHandler) SendMessageGeneral(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone     string `json:"phone"`      // Session phone to use (can be session ID or actual phone)
		SessionID string `json:"session_id"` // Legacy field for backward compatibility
		To        string `json:"to"`         // Recipient
		Message   string `json:"message"`    // Message content
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Support both 'phone' and 'session_id' for backward compatibility
	phoneIdentifier := req.Phone
	if phoneIdentifier == "" {
		phoneIdentifier = req.SessionID
	}

	// Validate input
	if phoneIdentifier == "" || req.To == "" || req.Message == "" {
		http.Error(w, "phone (or session_id), to, and message fields are required", http.StatusBadRequest)
		return
	}

	// Find session by phone identifier (session ID or actual phone)
	sessionID := h.whatsappService.FindSessionByPhone(phoneIdentifier)
	if sessionID == "" {
		HandleErrorWithMessage(w, http.StatusNotFound, "Session not found for phone: "+phoneIdentifier, models.ErrCodeNotFound)
		return
	}

	// Create message request
	msgReq := &models.SendMessageRequest{
		To:      req.To,
		Message: req.Message,
	}

	// Send message
	messageID, err := h.whatsappService.SendMessage(sessionID, msgReq)
	if err != nil {
		h.logger.Error("Failed to send message from session %s: %v", sessionID, err)
		HandleError(w, err)
		return
	}

	h.logger.Info("API message sent successfully with ID: %s", messageID)

	WriteSuccessResponse(w, "Message sent successfully", map[string]interface{}{
		"message_id": messageID,
		"timestamp":  time.Now().Unix(),
		"session":    sessionID,
	})
}

// WebSocketHandler handles WebSocket connections for real-time updates
func (h *SessionHandler) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	// Authenticate via token query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Authentication token required", http.StatusUnauthorized)
		return
	}

	// Validate JWT token
	if err := h.validateJWT(token); err != nil {
		http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Check if session exists
	session, exists := h.whatsappService.GetSession(sessionID)
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Upgrade connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade WebSocket connection: %v", err)
		return
	}
	defer conn.Close()

	h.logger.Info("WebSocket connection established for session %s", sessionID)

	// Start QR code streaming if not logged in
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	if !session.LoggedIn {
		h.logger.Info("Starting QR code generation for unauthenticated session %s", sessionID)
		
		// Disconnect if already connected but not logged in (like original implementation)
		if session.Connected && session.Client.IsConnected() && !session.Client.IsLoggedIn() {
			h.logger.Info("Disconnecting stale connection for session %s", sessionID)
			session.Client.Disconnect()
			// Give it a moment to disconnect cleanly
			time.Sleep(1 * time.Second)
		}

		// Get QR channel before connecting (like original implementation)
		h.logger.Debug("Getting QR channel for session %s", sessionID)
		qrChan, err := session.Client.GetQRChannel(ctx)
		if err != nil {
			h.logger.Error("Failed to get QR channel for session %s: %v", sessionID, err)
			conn.WriteJSON(models.WebSocketMessage{
				Type: "error",
				Data: map[string]string{"error": "Failed to get QR channel: " + err.Error()},
			})
			return
		}

		// Start QR code streaming
		h.logger.Debug("Starting QR code streaming for session %s", sessionID)
		go h.streamQRUpdatesFromChannel(ctx, conn, qrChan, sessionID)

		// Now connect after getting QR channel (only if not already connected)
		if !session.Connected {
			h.logger.Info("Connecting session %s for QR generation", sessionID)
			if err := h.whatsappService.ConnectSession(sessionID); err != nil {
				h.logger.Error("Failed to connect session %s: %v", sessionID, err)
				conn.WriteJSON(models.WebSocketMessage{
					Type: "error",
					Data: map[string]string{"error": "Failed to connect: " + err.Error()},
				})
				return
			}
		} else {
			h.logger.Info("Session %s already connected, starting QR generation", sessionID)
		}
	} else {
		h.logger.Info("Session %s already logged in, no QR needed", sessionID)
	}

	// Handle WebSocket messages
	for {
		var msg models.WebSocketMessage
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Info("WebSocket connection closed for session %s", sessionID)
			} else {
				h.logger.Error("WebSocket read error for session %s: %v", sessionID, err)
			}
			break
		}

		// Handle different message types
		switch msg.Type {
		case "ping":
			if err := conn.WriteJSON(models.WebSocketMessage{Type: "pong"}); err != nil {
				h.logger.Error("WebSocket pong error for session %s: %v", sessionID, err)
				return
			}
		default:
			h.logger.Debug("Received WebSocket message type %s for session %s", msg.Type, sessionID)
		}
	}

	// Clean up: If WebSocket connection is closed and session is not logged in,
	// disconnect the session to avoid showing it as "connected" when it's not authenticated
	// Run this in a goroutine with delay to avoid interfering with immediate reconnections
	go func() {
		if session, exists := h.whatsappService.GetSession(sessionID); exists {
			if !session.LoggedIn && session.Connected {
				// Give a delay to allow for immediate reconnection attempts
				time.Sleep(5 * time.Second)
				
				// Check again after delay - if still not logged in, disconnect
				if session, exists := h.whatsappService.GetSession(sessionID); exists && !session.LoggedIn && session.Connected {
					h.logger.Info("WebSocket closed for unauthenticated session %s, disconnecting after delay", sessionID)
					if err := h.whatsappService.DisconnectSession(sessionID); err != nil {
						h.logger.Error("Failed to disconnect unauthenticated session %s: %v", sessionID, err)
					}
				}
			}
		}
	}()
}

// validateJWT validates a JWT token
func (h *SessionHandler) validateJWT(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.jwtSecret), nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}

// LoginSession handles session login (WhatsApp authentication)
func (h *SessionHandler) LoginSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	if err := h.whatsappService.LoginSession(sessionID); err != nil {
		h.logger.Error("Failed to login session %s: %v", sessionID, err)
		HandleError(w, err)
		return
	}

	WriteSuccessResponse(w, "Login process initiated", nil)
}

// LogoutSession handles session logout (WhatsApp logout)
func (h *SessionHandler) LogoutSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	if err := h.whatsappService.LogoutSession(sessionID); err != nil {
		h.logger.Error("Failed to logout session %s: %v", sessionID, err)
		HandleError(w, err)
		return
	}

	WriteSuccessResponse(w, "Session logged out successfully", nil)
}

// UpdateSessionWebhook handles updating session webhook URL
func (h *SessionHandler) UpdateSessionWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	var req struct {
		WebhookURL string `json:"webhook_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updateReq := &models.UpdateSessionRequest{
		WebhookURL: req.WebhookURL,
	}

	if err := h.whatsappService.UpdateSession(sessionID, updateReq); err != nil {
		h.logger.Error("Failed to update session webhook %s: %v", sessionID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Webhook updated successfully"})
}

// UpdateSessionName handles updating session name
func (h *SessionHandler) UpdateSessionName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	updateReq := &models.UpdateSessionRequest{
		Name: req.Name,
	}

	if err := h.whatsappService.UpdateSession(sessionID, updateReq); err != nil {
		h.logger.Error("Failed to update session name %s: %v", sessionID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Session name updated successfully"})
}

// generateSessionID generates a random 10-digit session ID
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

// streamQRUpdatesFromChannel streams QR code updates from a QR channel directly
func (h *SessionHandler) streamQRUpdatesFromChannel(ctx context.Context, conn *websocket.Conn, qrChan <-chan whatsmeow.QRChannelItem, sessionID string) {
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-qrChan:
			if !ok {
				h.logger.Info("QR channel closed for session %s", sessionID)
				return
			}

			h.logger.Debug("QR event for session %s: %s", sessionID, evt.Event)

			var msgType string
			var data interface{}

			switch evt.Event {
			case "code":
				msgType = "qr"
				data = map[string]interface{}{
					"qr":      evt.Code,
					"timeout": evt.Timeout,
				}
			case "success":
				msgType = "success"
				data = map[string]string{"message": "Login successful"}
			case "timeout":
				msgType = "qr_timeout"
				data = map[string]string{"message": "QR code timeout"}
			default:
				h.logger.Debug("Unknown QR event: %s", evt.Event)
				continue
			}

			wsMsg := models.WebSocketMessage{
				Type: msgType,
				Data: data,
			}

			if err := conn.WriteJSON(wsMsg); err != nil {
				h.logger.Error("Failed to send QR update for session %s: %v", sessionID, err)
				return
			}

			h.logger.Debug("Sent QR update (%s) for session %s", msgType, sessionID)

			// Stop streaming on success or timeout
			if evt.Event == "success" || evt.Event == "timeout" {
				return
			}
		}
	}
}
