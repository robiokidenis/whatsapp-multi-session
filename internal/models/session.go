package models

import (
	"time"
	
	"go.mau.fi/whatsmeow"
)

// ProxyConfig represents proxy configuration for a session
type ProxyConfig struct {
	Enabled  bool   `json:"enabled"`
	Type     string `json:"type"`     // http, https, socks5
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// Session represents a WhatsApp session
type Session struct {
	ID            string                         `json:"id"`
	Phone         string                         `json:"phone"`        // Session identifier
	ActualPhone   string                         `json:"actual_phone"` // Actual WhatsApp phone number
	Name          string                         `json:"name"`
	Position      int                            `json:"position"`
	WebhookURL    string                         `json:"webhook_url"`
	AutoReplyText *string                        `json:"auto_reply_text,omitempty"` // Auto reply text, nullable
	ProxyConfig   *ProxyConfig                   `json:"proxy_config,omitempty"`    // Proxy configuration, nullable
	Enabled       bool                           `json:"enabled"`                   // Session enabled/disabled status
	Client        *whatsmeow.Client              `json:"-"`
	QRChan        <-chan whatsmeow.QRChannelItem `json:"-"`
	Connected     bool                           `json:"connected"`
	LoggedIn      bool                           `json:"logged_in"`
	Connecting    bool                           `json:"-"`
}

// SessionMetadata represents session data stored in database
type SessionMetadata struct {
	ID            string       `json:"id"`
	Phone         string       `json:"phone"`
	ActualPhone   string       `json:"actual_phone"`
	Name          string       `json:"name"`
	Position      int          `json:"position"`
	WebhookURL    string       `json:"webhook_url"`
	AutoReplyText *string      `json:"auto_reply_text,omitempty"` // Auto reply text, nullable
	ProxyConfig   *ProxyConfig `json:"proxy_config,omitempty"`    // Proxy configuration, nullable
	Enabled       bool         `json:"enabled"`                   // Session enabled/disabled status
	UserID        int          `json:"user_id"`
	CreatedAt     time.Time    `json:"created_at"`
}

// CreateSessionRequest represents session creation request
type CreateSessionRequest struct {
	Phone         string       `json:"phone,omitempty"`
	Name          string       `json:"name"`
	Position      int          `json:"position,omitempty"`
	WebhookURL    string       `json:"webhook_url,omitempty"`
	AutoReplyText *string      `json:"auto_reply_text,omitempty"` // Auto reply text, nullable
	ProxyConfig   *ProxyConfig `json:"proxy_config,omitempty"`    // Proxy configuration, nullable
	Enabled       bool         `json:"enabled,omitempty"`         // Session enabled status, defaults to true
}

// UpdateSessionRequest represents session update request
type UpdateSessionRequest struct {
	Name          string       `json:"name,omitempty"`
	WebhookURL    string       `json:"webhook_url,omitempty"`
	Position      int          `json:"position,omitempty"`
	AutoReplyText *string      `json:"auto_reply_text,omitempty"` // Auto reply text, nullable
	ProxyConfig   *ProxyConfig `json:"proxy_config,omitempty"`    // Proxy configuration, nullable
	Enabled       *bool        `json:"enabled,omitempty"`         // Session enabled status, nullable for explicit updates
}

// SessionResponse represents session response
type SessionResponse struct {
	ID            string       `json:"id"`
	Phone         string       `json:"phone"`
	ActualPhone   string       `json:"actual_phone,omitempty"`
	Name          string       `json:"name"`
	Position      int          `json:"position"`
	WebhookURL    string       `json:"webhook_url,omitempty"`
	AutoReplyText *string      `json:"auto_reply_text,omitempty"` // Auto reply text, nullable
	ProxyConfig   *ProxyConfig `json:"proxy_config,omitempty"`    // Proxy configuration, nullable
	Enabled       bool         `json:"enabled"`                   // Session enabled/disabled status
	Connected     bool         `json:"connected"`
	LoggedIn      bool         `json:"logged_in"`
	QRCode        string       `json:"qr_code,omitempty"`
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