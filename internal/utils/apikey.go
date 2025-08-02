package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
)

// GenerateAPIKey generates a secure API key
func GenerateAPIKey() (string, error) {
	// Generate 32 random bytes (256 bits)
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %v", err)
	}

	// Encode to base64 and make URL-safe
	apiKey := base64.URLEncoding.EncodeToString(bytes)
	
	// Remove padding and make it cleaner
	apiKey = strings.TrimRight(apiKey, "=")
	
	// Add prefix to identify it as an API key
	return "wams_" + apiKey, nil
}

// ValidateAPIKeyFormat checks if an API key has the correct format
func ValidateAPIKeyFormat(apiKey string) bool {
	// Should start with wams_ and be at least 47 characters total
	return strings.HasPrefix(apiKey, "wams_") && len(apiKey) >= 47
}