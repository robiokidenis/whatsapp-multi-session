package handlers

import (
	"encoding/json"
	"net/http"

	"whatsapp-multi-session/internal/types"
)

// SessionResponse represents a session in API responses
type SessionResponse struct {
	ID          string `json:"id"`
	Phone       string `json:"phone"`
	ActualPhone string `json:"actual_phone,omitempty"`
	Name        string `json:"name"`
	Position    int    `json:"position"`
	Connected   bool   `json:"connected"`
	LoggedIn    bool   `json:"logged_in"`
	WebhookURL  string `json:"webhook_url,omitempty"`
	CreatedAt   int64  `json:"created_at"`
}

// ListSessionsResponse represents the response for listing sessions
type ListSessionsResponse struct {
	Success  bool              `json:"success"`
	Message  string            `json:"message"`
	Sessions []SessionResponse `json:"sessions"`
}

// ListSessionsHandler handles listing all sessions
func ListSessionsHandler(sessionManager *types.SessionManager, logger *types.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionManager.Mu.RLock()
		defer sessionManager.Mu.RUnlock()

		sessions := make([]SessionResponse, 0, len(sessionManager.Sessions))
		for _, session := range sessionManager.Sessions {
			sessions = append(sessions, SessionResponse{
				ID:          session.ID,
				Phone:       session.Phone,
				ActualPhone: session.ActualPhone,
				Name:        session.Name,
				Position:    session.Position,
				Connected:   session.Connected,
				LoggedIn:    session.LoggedIn,
				WebhookURL:  session.WebhookURL,
				CreatedAt:   session.CreatedAt,
			})
		}

		if logger != nil {
			logger.Info("Listed %d sessions", len(sessions))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ListSessionsResponse{
			Success:  true,
			Message:  "Sessions retrieved successfully",
			Sessions: sessions,
		})
	}
}

// CreateSessionRequest represents the request to create a new session
type CreateSessionRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone,omitempty"`
}

// CreateSessionResponse represents the response for creating a session
type CreateSessionResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Session SessionResponse `json:"session,omitempty"`
}

// CreateSessionHandler handles creating new sessions
func CreateSessionHandler(sessionManager *types.SessionManager, logger *types.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateSessionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(CreateSessionResponse{
				Success: false,
				Message: "Invalid request body",
			})
			return
		}

		// For now, return a placeholder response
		// TODO: Implement actual session creation logic
		if logger != nil {
			logger.Info("Create session request: name=%s, phone=%s", req.Name, req.Phone)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CreateSessionResponse{
			Success: true,
			Message: "Session creation endpoint ready (implementation pending)",
		})
	}
}