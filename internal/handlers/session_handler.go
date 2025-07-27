package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"whatsapp-multi-session/internal/models"
	"whatsapp-multi-session/internal/services"
	"whatsapp-multi-session/pkg/logger"
)

// SessionHandler handles session-related endpoints
type SessionHandler struct {
	whatsappService *services.WhatsAppService
	logger          *logger.Logger
	upgrader        websocket.Upgrader
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(
	whatsappService *services.WhatsAppService,
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

	// Validate input
	if req.Name == "" {
		http.Error(w, "Session name is required", http.StatusBadRequest)
		return
	}

	// Create session
	session, err := h.whatsappService.CreateSession(&req)
	if err != nil {
		h.logger.Error("Failed to create session: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to response
	response := &models.SessionResponse{
		ID:         session.ID,
		Phone:      session.Phone,
		ActualPhone: session.ActualPhone,
		Name:       session.Name,
		Position:   session.Position,
		WebhookURL: session.WebhookURL,
		Connected:  session.Connected,
		LoggedIn:   session.LoggedIn,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSessions handles getting all sessions
func (h *SessionHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
	sessions := h.whatsappService.GetAllSessions()
	
	// Convert to response format
	responses := make([]*models.SessionResponse, len(sessions))
	for i, session := range sessions {
		responses[i] = &models.SessionResponse{
			ID:         session.ID,
			Phone:      session.Phone,
			ActualPhone: session.ActualPhone,
			Name:       session.Name,
			Position:   session.Position,
			WebhookURL: session.WebhookURL,
			Connected:  session.Connected,
			LoggedIn:   session.LoggedIn,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// GetSession handles getting a specific session
func (h *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	session, exists := h.whatsappService.GetSession(sessionID)
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	response := &models.SessionResponse{
		ID:         session.ID,
		Phone:      session.Phone,
		ActualPhone: session.ActualPhone,
		Name:       session.Name,
		Position:   session.Position,
		WebhookURL: session.WebhookURL,
		Connected:  session.Connected,
		LoggedIn:   session.LoggedIn,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ConnectSession handles session connection
func (h *SessionHandler) ConnectSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	if err := h.whatsappService.ConnectSession(sessionID); err != nil {
		h.logger.Error("Failed to connect session %s: %v", sessionID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Connection initiated"})
}

// DisconnectSession handles session disconnection
func (h *SessionHandler) DisconnectSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	if err := h.whatsappService.DisconnectSession(sessionID); err != nil {
		h.logger.Error("Failed to disconnect session %s: %v", sessionID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Session disconnected"})
}

// DeleteSession handles session deletion
func (h *SessionHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	if err := h.whatsappService.DeleteSession(sessionID); err != nil {
		h.logger.Error("Failed to delete session %s: %v", sessionID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Session deleted"})
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := &models.MessageResponse{
		Success: true,
		ID:      messageID,
		Message: "Message sent successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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

	response := &models.MessageResponse{
		Success: true,
		ID:      messageID,
		Message: "Attachment sent successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
		"groups":  groups,
	})
}

// SendMessageGeneral handles sending messages to any session (for compatibility)
func (h *SessionHandler) SendMessageGeneral(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID string `json:"session_id"`
		To        string `json:"to"`
		Message   string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.SessionID == "" || req.To == "" || req.Message == "" {
		http.Error(w, "session_id, to, and message fields are required", http.StatusBadRequest)
		return
	}

	// Create message request
	msgReq := &models.SendMessageRequest{
		To:      req.To,
		Message: req.Message,
	}

	// Send message
	messageID, err := h.whatsappService.SendMessage(req.SessionID, msgReq)
	if err != nil {
		h.logger.Error("Failed to send message from session %s: %v", req.SessionID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := &models.MessageResponse{
		Success: true,
		ID:      messageID,
		Message: "Message sent successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// WebSocketHandler handles WebSocket connections for real-time updates
func (h *SessionHandler) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	// Check if session exists
	_, exists := h.whatsappService.GetSession(sessionID)
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

		// Echo message back (for now)
		if err := conn.WriteJSON(msg); err != nil {
			h.logger.Error("WebSocket write error for session %s: %v", sessionID, err)
			break
		}
	}
}