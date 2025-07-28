package models

import "time"

// SendMessageRequest represents a message send request
type SendMessageRequest struct {
	To      string `json:"to"`
	Message string `json:"message"`
}

// SendImageRequest represents an image send request
type SendImageRequest struct {
	To      string `json:"to"`
	Image   string `json:"image"`   // Base64 encoded image
	Caption string `json:"caption"`
}

// SendFileRequest represents a file send request
type SendFileRequest struct {
	To       string `json:"to"`
	File     string `json:"file"`     // Base64 encoded file
	FileName string `json:"filename"`
	Caption  string `json:"caption"`
}

// SendFileURLRequest represents a file send request from URL
type SendFileURLRequest struct {
	To       string `json:"to"`
	URL      string `json:"url"`      // File URL to download
	FileName string `json:"filename,omitempty"`
	Caption  string `json:"caption,omitempty"`
	Type     string `json:"type,omitempty"` // image, video, audio, document
}

// MessageResponse represents a message response
type MessageResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// WebhookMessage represents a message for webhook delivery
type WebhookMessage struct {
	SessionID   string    `json:"session_id"`
	From        string    `json:"from"`
	To          string    `json:"to"`
	Message     string    `json:"message"`
	MessageType string    `json:"message_type"`
	Timestamp   time.Time `json:"timestamp"`
	ID          string    `json:"id"`
	IsGroup     bool      `json:"is_group"`
	GroupID     string    `json:"group_id,omitempty"`
	MediaURL    string    `json:"media_url,omitempty"`
}

// ContactInfo represents WhatsApp contact information
type ContactInfo struct {
	JID         string `json:"jid"`
	Name        string `json:"name"`
	BusinessName string `json:"business_name,omitempty"`
	PushName    string `json:"push_name,omitempty"`
	Verified    bool   `json:"verified"`
}