package repository

import (
	"fmt"
)

func (d *Database) createMessagesTable() error {
	var query string
	driver := d.db.Driver()
	driverName := fmt.Sprintf("%T", driver)
	
	if contains(driverName, "mysql") {
		query = `
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
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
	} else {
		query = `
		CREATE TABLE IF NOT EXISTS messages (
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
		)`
		
		_, err := d.db.Exec(query)
		if err != nil {
			return err
		}
		
		indexes := []string{
			"CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id)",
			"CREATE INDEX IF NOT EXISTS idx_messages_sender_jid ON messages(sender_jid)",
			"CREATE INDEX IF NOT EXISTS idx_messages_recipient_jid ON messages(recipient_jid)",
			"CREATE INDEX IF NOT EXISTS idx_messages_direction ON messages(direction)",
			"CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status)",
			"CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at)",
		}
		
		for _, indexQuery := range indexes {
			if _, err := d.db.Exec(indexQuery); err != nil {
				return err
			}
		}
		
		return nil
	}

	_, err := d.db.Exec(query)
	return err
}