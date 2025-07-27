package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// GenerateSessionID generates a unique session identifier
func GenerateSessionID() string {
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

// GeneratePhoneJID generates a WhatsApp JID format from session ID
func GeneratePhoneJID(sessionID string) string {
	return sessionID + "@s.whatsapp.net"
}