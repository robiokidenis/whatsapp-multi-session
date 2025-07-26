package session

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"whatsapp-multi-session/internal/database"
	"whatsapp-multi-session/internal/types"
)

// RestoreSession recreates a session from stored metadata
func RestoreSession(metadata types.SessionMetadata, container *sqlstore.Container, sessionManager *types.SessionManager, logger *types.Logger) *types.Session {
	if logger != nil {
		logger.Info("Restoring session %s (%s)", metadata.ID, metadata.Name)
	}

	// Find existing device in store
	var device *store.Device
	devices, err := container.GetAllDevices(context.Background())
	if err != nil {
		if logger != nil {
			logger.Error("Error getting devices: %v", err)
		}
		return nil
	}

	// Look for existing device
	for _, d := range devices {
		if d != nil && d.ID != nil && d.ID.User == strings.Replace(metadata.ActualPhone, "@s.whatsapp.net", "", 1) {
			device = d
			if logger != nil {
				logger.Debug("Found existing device for session %s", metadata.ID)
			}
			break
		}
	}

	// If no existing device found, create new one (will need re-authentication)
	if device == nil {
		device = container.NewDevice()
		if logger != nil {
			logger.Info("Created new device for session %s (will need re-authentication)", metadata.ID)
		}
	}

	// Create client
	clientLog := waLog.Stdout("Client:"+metadata.ID, "INFO", true)
	client := whatsmeow.NewClient(device, clientLog)
	client.EnableAutoReconnect = true
	client.AutoTrustIdentity = true

	// Create session
	session := &types.Session{
		ID:          metadata.ID,
		Phone:       metadata.Phone,
		ActualPhone: metadata.ActualPhone,
		Name:        metadata.Name,
		Position:    metadata.Position,
		WebhookURL:  metadata.WebhookURL,
		Client:      client,
		Connected:   false,
		LoggedIn:    false,
	}

	// Register event handlers
	client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Connected:
			if logger != nil {
				logger.Info("Session %s connected", metadata.ID)
			}
			sessionManager.Mu.Lock()
			if s, ok := sessionManager.Sessions[metadata.ID]; ok {
				s.Connected = true
				s.LoggedIn = client.IsLoggedIn()

				// Update actual phone number if logged in
				if client.IsLoggedIn() && client.Store.ID != nil {
					s.ActualPhone = client.Store.ID.User + "@s.whatsapp.net"
					if logger != nil {
						logger.Info("Session %s actual phone: %s", metadata.ID, s.ActualPhone)
					}
					// Save updated metadata
					database.SaveSessionMetadata(sessionManager, s)
				}
			}
			sessionManager.Mu.Unlock()
		case *events.Disconnected:
			if logger != nil {
				logger.Info("Session %s disconnected", metadata.ID)
			}
			sessionManager.Mu.Lock()
			if s, ok := sessionManager.Sessions[metadata.ID]; ok {
				s.Connected = false
			}
			sessionManager.Mu.Unlock()
		case *events.Message:
			if logger != nil {
				logger.Info("Received message in session %s: %s", metadata.ID, v.Message.GetConversation())
			}
			// Send webhook if configured
			sessionManager.Mu.RLock()
			if s, ok := sessionManager.Sessions[metadata.ID]; ok {
				if logger != nil {
					logger.Info("Session %s webhook URL: '%s'", metadata.ID, s.WebhookURL)
				}
				if s.WebhookURL != "" {
					go SendWebhook(s.WebhookURL, metadata.ID, v, logger)
				} else {
					if logger != nil {
						logger.Info("No webhook URL configured for session %s", metadata.ID)
					}
				}
			} else {
				if logger != nil {
					logger.Warn("Session %s not found in sessionManager.sessions", metadata.ID)
				}
			}
			sessionManager.Mu.RUnlock()
		}
	})

	// Try to connect if device has stored credentials
	if client.Store.ID != nil {
		go func() {
			if logger != nil {
				logger.Info("Auto-connecting restored session %s", metadata.ID)
			}
			err := client.Connect()
			if err != nil {
				if logger != nil {
					logger.Error("Failed to auto-connect session %s: %v", metadata.ID, err)
				}
			}
		}()
	}

	return session
}

// SendWebhook sends incoming message data to configured webhook URL
func SendWebhook(webhookURL, sessionID string, message *events.Message, logger *types.Logger) {
	if logger != nil {
		logger.Info("Sending webhook for session %s to URL: %s", sessionID, webhookURL)
	}

	// Detect message type and content
	var messageType string
	var content interface{}
	var mediaInfo map[string]interface{}

	// Check for different message types
	if message.Message.GetConversation() != "" {
		messageType = "text"
		content = message.Message.GetConversation()
	} else if imageMessage := message.Message.GetImageMessage(); imageMessage != nil {
		messageType = "image"
		content = imageMessage.GetCaption()
		mediaInfo = map[string]interface{}{
			"url":         imageMessage.GetURL(),
			"mime_type":   imageMessage.GetMimetype(),
			"file_length": imageMessage.GetFileLength(),
			"width":       imageMessage.GetWidth(),
			"height":      imageMessage.GetHeight(),
			"caption":     imageMessage.GetCaption(),
		}
	} else if videoMessage := message.Message.GetVideoMessage(); videoMessage != nil {
		messageType = "video"
		content = videoMessage.GetCaption()
		mediaInfo = map[string]interface{}{
			"url":         videoMessage.GetURL(),
			"mime_type":   videoMessage.GetMimetype(),
			"file_length": videoMessage.GetFileLength(),
			"duration":    videoMessage.GetSeconds(),
			"width":       videoMessage.GetWidth(),
			"height":      videoMessage.GetHeight(),
			"caption":     videoMessage.GetCaption(),
		}
	} else if audioMessage := message.Message.GetAudioMessage(); audioMessage != nil {
		messageType = "audio"
		content = ""
		mediaInfo = map[string]interface{}{
			"url":         audioMessage.GetURL(),
			"mime_type":   audioMessage.GetMimetype(),
			"file_length": audioMessage.GetFileLength(),
			"duration":    audioMessage.GetSeconds(),
			"voice_note":  audioMessage.GetPTT(), // Push-to-talk (voice note)
		}
	} else if documentMessage := message.Message.GetDocumentMessage(); documentMessage != nil {
		messageType = "document"
		content = documentMessage.GetTitle()
		mediaInfo = map[string]interface{}{
			"url":         documentMessage.GetURL(),
			"mime_type":   documentMessage.GetMimetype(),
			"file_length": documentMessage.GetFileLength(),
			"filename":    documentMessage.GetFileName(),
			"title":       documentMessage.GetTitle(),
		}
	} else if stickerMessage := message.Message.GetStickerMessage(); stickerMessage != nil {
		messageType = "sticker"
		content = ""
		mediaInfo = map[string]interface{}{
			"url":         stickerMessage.GetURL(),
			"mime_type":   stickerMessage.GetMimetype(),
			"file_length": stickerMessage.GetFileLength(),
			"width":       stickerMessage.GetWidth(),
			"height":      stickerMessage.GetHeight(),
		}
	} else if locationMessage := message.Message.GetLocationMessage(); locationMessage != nil {
		messageType = "location"
		content = locationMessage.GetName()
		mediaInfo = map[string]interface{}{
			"latitude":  locationMessage.GetDegreesLatitude(),
			"longitude": locationMessage.GetDegreesLongitude(),
			"name":      locationMessage.GetName(),
			"address":   locationMessage.GetAddress(),
		}
	} else if contactMessage := message.Message.GetContactMessage(); contactMessage != nil {
		messageType = "contact"
		content = contactMessage.GetDisplayName()
		mediaInfo = map[string]interface{}{
			"display_name": contactMessage.GetDisplayName(),
			"vcard":        contactMessage.GetVcard(),
		}
	} else {
		messageType = "unknown"
		content = "Unsupported message type"
	}

	// Prepare webhook payload
	webhookData := map[string]interface{}{
		"session_id":   sessionID,
		"timestamp":    message.Info.Timestamp.Unix(),
		"message_id":   message.Info.ID,
		"from": map[string]interface{}{
			"jid":       message.Info.Sender.String(),
			"phone":     message.Info.Sender.User,
			"push_name": message.Info.PushName,
		},
		"message_type": messageType,
		"content":      content,
		"is_from_me":   message.Info.IsFromMe,
		"is_group":     message.Info.Sender.Server == "g.us",
	}

	// Add media info if present
	if mediaInfo != nil {
		webhookData["media"] = mediaInfo
	}

	// Add group info if it's a group message
	if message.Info.Sender.Server == "g.us" {
		webhookData["group"] = map[string]interface{}{
			"jid":  message.Info.Chat.String(),
			"name": "", // Group name would need to be fetched separately
		}
	}

	// Convert to JSON
	jsonData, err := json.Marshal(webhookData)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to marshal webhook data for session %s: %v", sessionID, err)
		}
		return
	}

	// Send HTTP POST request to webhook URL
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Post(webhookURL, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		if logger != nil {
			logger.Error("Failed to send webhook for session %s to %s: %v", sessionID, webhookURL, err)
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if logger != nil {
			logger.Info("Webhook sent successfully for session %s to %s", sessionID, webhookURL)
		}
	} else {
		if logger != nil {
			logger.Warn("Webhook failed for session %s to %s: HTTP %d", sessionID, webhookURL, resp.StatusCode)
		}
	}
}