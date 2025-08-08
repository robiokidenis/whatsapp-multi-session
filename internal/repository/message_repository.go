package repository

import (
	"database/sql"
	"fmt"
	"time"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Message represents a WhatsApp message
type Message struct {
	ID           int64     `json:"id"`
	SessionID    string    `json:"session_id"`
	MessageID    string    `json:"message_id"`
	SenderJID    string    `json:"sender_jid"`
	RecipientJID string    `json:"recipient_jid"`
	MessageType  string    `json:"message_type"`
	Content      string    `json:"content"`
	MediaURL     string    `json:"media_url"`
	Direction    string    `json:"direction"` // 'sent' or 'received'
	Status       string    `json:"status"`    // 'pending', 'sent', 'delivered', 'read', 'failed'
	ErrorMessage string    `json:"error_message"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// LogMessage logs a message to the database
func (r *MessageRepository) LogMessage(message *Message) error {
	// Check if messages table exists, create if not
	if err := r.ensureMessagesTable(); err != nil {
		return err
	}

	query := `
		INSERT INTO messages (
			session_id, message_id, sender_jid, recipient_jid, 
			message_type, content, media_url, direction, 
			status, error_message, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		message.SessionID,
		message.MessageID,
		message.SenderJID,
		message.RecipientJID,
		message.MessageType,
		message.Content,
		message.MediaURL,
		message.Direction,
		message.Status,
		message.ErrorMessage,
		message.CreatedAt,
		message.UpdatedAt,
	)

	return err
}

// UpdateMessageStatus updates the status of a message
func (r *MessageRepository) UpdateMessageStatus(messageID, status, errorMessage string) error {
	query := `
		UPDATE messages 
		SET status = ?, error_message = ?, updated_at = ?
		WHERE message_id = ?
	`

	_, err := r.db.Exec(query, status, errorMessage, time.Now(), messageID)
	return err
}

// GetMessagesBySession gets messages for a specific session
func (r *MessageRepository) GetMessagesBySession(sessionID string, limit int) ([]*Message, error) {
	query := `
		SELECT id, session_id, message_id, sender_jid, recipient_jid,
		       message_type, content, media_url, direction, status,
		       error_message, created_at, updated_at
		FROM messages 
		WHERE session_id = ? 
		ORDER BY created_at DESC 
		LIMIT ?
	`

	rows, err := r.db.Query(query, sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		msg := &Message{}
		err := rows.Scan(
			&msg.ID, &msg.SessionID, &msg.MessageID,
			&msg.SenderJID, &msg.RecipientJID, &msg.MessageType,
			&msg.Content, &msg.MediaURL, &msg.Direction,
			&msg.Status, &msg.ErrorMessage, &msg.CreatedAt, &msg.UpdatedAt,
		)
		if err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// ensureMessagesTable creates the messages table if it doesn't exist
func (r *MessageRepository) ensureMessagesTable() error {
	// Get database driver name to detect MySQL vs SQLite
	driver := r.db.Driver()
	driverName := fmt.Sprintf("%T", driver)
	
	if contains(driverName, "mysql") {
		// For MySQL, check if table exists using INFORMATION_SCHEMA
		var count int
		checkQuery := `
			SELECT COUNT(*) 
			FROM INFORMATION_SCHEMA.TABLES 
			WHERE TABLE_SCHEMA = DATABASE() 
			AND TABLE_NAME = 'messages'
		`
		err := r.db.QueryRow(checkQuery).Scan(&count)
		if err != nil {
			return err
		}
		
		if count > 0 {
			return nil // Table already exists
		}
		
		// Create the messages table for MySQL
		createQuery := `
			CREATE TABLE IF NOT EXISTS messages (
				id INT AUTO_INCREMENT PRIMARY KEY,
				session_id VARCHAR(255) NOT NULL,
				message_id VARCHAR(255) UNIQUE,
				sender_jid VARCHAR(100),
				recipient_jid VARCHAR(100),
				message_type VARCHAR(50) NOT NULL DEFAULT 'text',
				content TEXT,
				media_url TEXT,
				direction VARCHAR(20) NOT NULL,
				status VARCHAR(50),
				error_message TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
				INDEX idx_session_id (session_id),
				INDEX idx_sender_jid (sender_jid),
				INDEX idx_recipient_jid (recipient_jid),
				INDEX idx_direction (direction),
				INDEX idx_status (status),
				INDEX idx_created_at (created_at),
				FOREIGN KEY (session_id) REFERENCES session_metadata(id) ON DELETE CASCADE
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
		`
		
		_, err = r.db.Exec(createQuery)
		return err
		
	} else {
		// For SQLite, check if table exists using sqlite_master
		var count int
		checkQuery := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='messages'`
		err := r.db.QueryRow(checkQuery).Scan(&count)
		if err != nil {
			return err
		}
		
		if count > 0 {
			return nil // Table already exists
		}
		
		// Create the messages table for SQLite
		createQuery := `
			CREATE TABLE messages (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				session_id TEXT NOT NULL,
				message_id TEXT UNIQUE,
				sender_jid TEXT,
				recipient_jid TEXT,
				message_type TEXT NOT NULL DEFAULT 'text',
				content TEXT,
				media_url TEXT,
				direction TEXT NOT NULL,
				status TEXT,
				error_message TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (session_id) REFERENCES session_metadata(id) ON DELETE CASCADE
			)
		`
		
		_, err = r.db.Exec(createQuery)
		if err != nil {
			return err
		}
		
		// Create indexes for SQLite
		indexes := []string{
			"CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id)",
			"CREATE INDEX IF NOT EXISTS idx_messages_direction ON messages(direction)",
			"CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status)",
			"CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at)",
			"CREATE INDEX IF NOT EXISTS idx_messages_message_id ON messages(message_id)",
		}
		
		for _, indexQuery := range indexes {
			if _, err := r.db.Exec(indexQuery); err != nil {
				return err
			}
		}
		
		return nil
	}
}