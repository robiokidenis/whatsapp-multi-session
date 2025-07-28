package models

import (
	"time"
	
	"go.mau.fi/whatsmeow"
)

// Session represents a WhatsApp session
type Session struct {
	ID          string                         `json:"id"`
	Phone       string                         `json:"phone"`        // Session identifier
	ActualPhone string                         `json:"actual_phone"` // Actual WhatsApp phone number
	Name        string                         `json:"name"`
	Position    int                            `json:"position"`
	WebhookURL  string                         `json:"webhook_url"`
	Client      *whatsmeow.Client              `json:"-"`
	QRChan      <-chan whatsmeow.QRChannelItem `json:"-"`
	Connected   bool                           `json:"connected"`
	LoggedIn    bool                           `json:"logged_in"`
	Connecting  bool                           `json:"-"`
}

// SessionMetadata represents session data stored in database
type SessionMetadata struct {
	ID          string    `json:"id"`
	Phone       string    `json:"phone"`
	ActualPhone string    `json:"actual_phone"`
	Name        string    `json:"name"`
	Position    int       `json:"position"`
	WebhookURL  string    `json:"webhook_url"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateSessionRequest represents session creation request
type CreateSessionRequest struct {
	Phone      string `json:"phone,omitempty"`
	Name       string `json:"name"`
	Position   int    `json:"position,omitempty"`
	WebhookURL string `json:"webhook_url,omitempty"`
}

// UpdateSessionRequest represents session update request
type UpdateSessionRequest struct {
	Name       string `json:"name,omitempty"`
	WebhookURL string `json:"webhook_url,omitempty"`
	Position   int    `json:"position,omitempty"`
}

// SessionResponse represents session response
type SessionResponse struct {
	ID          string `json:"id"`
	Phone       string `json:"phone"`
	ActualPhone string `json:"actual_phone,omitempty"`
	Name        string `json:"name"`
	Position    int    `json:"position"`
	WebhookURL  string `json:"webhook_url,omitempty"`
	Connected   bool   `json:"connected"`
	LoggedIn    bool   `json:"logged_in"`
	QRCode      string `json:"qr_code,omitempty"`
}

// QRResponse represents QR code response
type QRResponse struct {
	QRCode string `json:"qr_code"`
}

// WebSocketMessage represents WebSocket messages
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}