package services

import (
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"

	"whatsapp-multi-session/internal/models"
	"whatsapp-multi-session/pkg/logger"
)

type ContactDetectionService struct {
	log logger.Logger
}

func NewContactDetectionService(log logger.Logger) *ContactDetectionService {
	return &ContactDetectionService{
		log: log,
	}
}

// DetectFromCSV parses CSV data and returns detected contacts with smart field detection
func (s *ContactDetectionService) DetectFromCSV(data io.Reader) ([]models.SmartContactDetection, error) {
	reader := csv.NewReader(data)
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %v", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// Detect header row and field mappings
	headers := records[0]
	fieldMapping := s.detectFieldMapping(headers)

	var contacts []models.SmartContactDetection

	// Process data rows (skip header)
	for i, record := range records[1:] {
		contact := s.detectContactFromRecord(record, fieldMapping, fmt.Sprintf("CSV row %d", i+2))
		if contact.Phone != "" || contact.Name != "" {
			contacts = append(contacts, contact)
		}
	}

	return contacts, nil
}

// DetectFromText parses plain text and extracts contact information
func (s *ContactDetectionService) DetectFromText(text string) ([]models.SmartContactDetection, error) {
	lines := strings.Split(text, "\n")
	var contacts []models.SmartContactDetection

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		contact := s.detectContactFromText(line, fmt.Sprintf("Text line %d", i+1))
		if contact.Phone != "" || contact.Name != "" {
			contacts = append(contacts, contact)
		}
	}

	return contacts, nil
}

// detectFieldMapping analyzes CSV headers to determine field types
func (s *ContactDetectionService) detectFieldMapping(headers []string) map[int]string {
	mapping := make(map[int]string)

	for i, header := range headers {
		header = strings.ToLower(strings.TrimSpace(header))

		// Phone number patterns
		if s.containsAny(header, []string{"phone", "mobile", "tel", "number", "whatsapp", "wa", "contact_number", "kontak", "nomor"}) {
			mapping[i] = "phone"
			continue
		}

		// Name patterns
		if s.containsAny(header, []string{"name", "fullname", "full_name", "contact", "person", "nama", "nama_lengkap", "nama_penuh"}) {
			mapping[i] = "name"
			continue
		}

		// Email patterns
		if s.containsAny(header, []string{"email", "mail", "e-mail", "email_address", "alamat_email"}) {
			mapping[i] = "email"
			continue
		}

		// Company patterns
		if s.containsAny(header, []string{"company", "organization", "org", "business", "enterprise", "perusahaan", "bisnis", "organisasi", "pt"}) {
			mapping[i] = "company"
			continue
		}

		// Position patterns
		if s.containsAny(header, []string{"position", "title", "job", "role", "designation", "jabatan", "tugas", "posisi", "pekerjaan"}) {
			mapping[i] = "position"
			continue
		}

		// If no specific pattern matches, try to infer from content in first few rows
		mapping[i] = "unknown"
	}

	return mapping
}

// detectContactFromRecord extracts contact info from CSV record
func (s *ContactDetectionService) detectContactFromRecord(record []string, fieldMapping map[int]string, source string) models.SmartContactDetection {
	contact := models.SmartContactDetection{
		Source:     source,
		RawData:    strings.Join(record, ","),
		Confidence: 0.0,
	}

	confidenceScore := 0.0
	totalFields := 0

	for i, value := range record {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		fieldType, exists := fieldMapping[i]
		if !exists {
			// Try to auto-detect field type from content
			fieldType = s.detectFieldType(value)
		}

		switch fieldType {
		case "phone":
			if phone := s.extractPhone(value); phone != "" {
				contact.Phone = phone
				confidenceScore += 40.0
			}
		case "name":
			if s.isValidName(value) {
				contact.Name = value
				confidenceScore += 30.0
			}
		case "email":
			if s.isValidEmail(value) {
				contact.Email = value
				confidenceScore += 20.0
			}
		case "company":
			contact.Company = value
			confidenceScore += 10.0
		case "position":
			contact.Position = value
			confidenceScore += 5.0
		case "unknown":
			// Try to auto-detect what this field might be
			if phone := s.extractPhone(value); phone != "" {
				contact.Phone = phone
				confidenceScore += 35.0 // Lower confidence for auto-detected
			} else if s.isValidEmail(value) {
				contact.Email = value
				confidenceScore += 15.0
			} else if s.isValidName(value) {
				if contact.Name == "" {
					contact.Name = value
					confidenceScore += 25.0
				}
			}
		}
		totalFields++
	}

	// Calculate final confidence score
	if totalFields > 0 {
		contact.Confidence = confidenceScore / 100.0
		if contact.Confidence > 1.0 {
			contact.Confidence = 1.0
		}
	}

	return contact
}

// detectContactFromText extracts contact info from plain text line
func (s *ContactDetectionService) detectContactFromText(text string, source string) models.SmartContactDetection {
	contact := models.SmartContactDetection{
		Source:     source,
		RawData:    text,
		Confidence: 0.0,
	}

	confidenceScore := 0.0

	// Extract phone number
	if phone := s.extractPhone(text); phone != "" {
		contact.Phone = phone
		confidenceScore += 50.0
	}

	// Extract email
	if email := s.extractEmail(text); email != "" {
		contact.Email = email
		confidenceScore += 20.0
	}

	// Extract name (remaining text after removing phone and email)
	cleanText := text
	if contact.Phone != "" {
		cleanText = s.removePhone(cleanText)
	}
	if contact.Email != "" {
		cleanText = s.removeEmail(cleanText)
	}

	// Clean and extract potential name
	cleanText = strings.TrimSpace(cleanText)
	cleanText = s.cleanPunctuation(cleanText)

	if s.isValidName(cleanText) {
		contact.Name = cleanText
		confidenceScore += 30.0
	}

	contact.Confidence = confidenceScore / 100.0
	if contact.Confidence > 1.0 {
		contact.Confidence = 1.0
	}

	return contact
}

// detectFieldType attempts to determine field type from content
func (s *ContactDetectionService) detectFieldType(value string) string {
	value = strings.TrimSpace(value)

	if s.extractPhone(value) != "" {
		return "phone"
	}

	if s.isValidEmail(value) {
		return "email"
	}

	if s.isValidName(value) {
		return "name"
	}

	return "unknown"
}

// extractPhone extracts and formats phone number from text
func (s *ContactDetectionService) extractPhone(text string) string {
	// Remove common non-digit characters but keep + for international numbers
	phoneRegexes := []*regexp.Regexp{
		regexp.MustCompile(`\+?[\d\s\-\(\)\.]{7,20}`),
		regexp.MustCompile(`\b\d{3,4}[\s\-]?\d{3,4}[\s\-]?\d{3,6}\b`),
		regexp.MustCompile(`\(\d{3,4}\)[\s\-]?\d{3,4}[\s\-]?\d{3,6}`),
	}

	for _, regex := range phoneRegexes {
		matches := regex.FindAllString(text, -1)
		for _, match := range matches {
			// Clean the phone number
			phone := s.cleanPhoneNumber(match)
			if s.isValidPhone(phone) {
				return phone
			}
		}
	}

	return ""
}

// cleanPhoneNumber removes formatting and validates phone number
func (s *ContactDetectionService) cleanPhoneNumber(phone string) string {
	// Remove all non-digit characters except +
	re := regexp.MustCompile(`[^\d+]`)
	cleaned := re.ReplaceAllString(phone, "")

	// Remove leading zeros but preserve international +
	if strings.HasPrefix(cleaned, "+") {
		return cleaned
	}

	// Remove leading zeros
	cleaned = strings.TrimLeft(cleaned, "0")

	// Add + if it looks like an international number
	if len(cleaned) > 10 {
		return "+" + cleaned
	}

	return cleaned
}

// isValidPhone validates if string is a valid phone number
func (s *ContactDetectionService) isValidPhone(phone string) bool {
	if len(phone) < 7 || len(phone) > 20 {
		return false
	}

	// Count digits
	digitCount := 0
	for _, r := range phone {
		if unicode.IsDigit(r) {
			digitCount++
		}
	}

	return digitCount >= 7 && digitCount <= 15
}

// extractEmail extracts email address from text
func (s *ContactDetectionService) extractEmail(text string) string {
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	matches := emailRegex.FindAllString(text, -1)

	for _, match := range matches {
		if s.isValidEmail(match) {
			return match
		}
	}

	return ""
}

// isValidEmail validates email format
func (s *ContactDetectionService) isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// isValidName checks if text could be a person's name
func (s *ContactDetectionService) isValidName(name string) bool {
	name = strings.TrimSpace(name)
	if len(name) < 2 || len(name) > 100 {
		return false
	}

	// Name should contain mostly letters and spaces
	letterCount := 0
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsSpace(r) {
			if unicode.IsLetter(r) {
				letterCount++
			}
		} else if !s.isAllowedNameCharacter(r) {
			return false
		}
	}

	// At least 50% should be letters
	return float64(letterCount)/float64(len(name)) >= 0.5
}

// isAllowedNameCharacter checks if character is allowed in names
func (s *ContactDetectionService) isAllowedNameCharacter(r rune) bool {
	allowedChars := []rune{'.', ',', '-', '\'', '(', ')'}
	for _, char := range allowedChars {
		if r == char {
			return true
		}
	}
	return false
}

// removePhone removes phone number from text
func (s *ContactDetectionService) removePhone(text string) string {
	phoneRegex := regexp.MustCompile(`\+?[\d\s\-\(\)\.]{7,20}`)
	return phoneRegex.ReplaceAllString(text, "")
}

// removeEmail removes email from text
func (s *ContactDetectionService) removeEmail(text string) string {
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	return emailRegex.ReplaceAllString(text, "")
}

// cleanPunctuation removes extra punctuation and whitespace
func (s *ContactDetectionService) cleanPunctuation(text string) string {
	// Remove multiple spaces
	spaceRegex := regexp.MustCompile(`\s+`)
	text = spaceRegex.ReplaceAllString(text, " ")

	// Remove leading/trailing punctuation
	text = strings.Trim(text, " .,;:-_()[]{}\"'")

	return text
}

// containsAny checks if string contains any of the given substrings
func (s *ContactDetectionService) containsAny(str string, substrs []string) bool {
	str = strings.ToLower(str)
	for _, substr := range substrs {
		if strings.Contains(str, strings.ToLower(substr)) {
			return true
		}
	}
	return false
}

// ValidateContacts validates and enhances detected contacts
func (s *ContactDetectionService) ValidateContacts(contacts []models.SmartContactDetection) []models.SmartContactDetection {
	var validated []models.SmartContactDetection

	for _, contact := range contacts {
		// Skip contacts with very low confidence
		if contact.Confidence < 0.3 {
			continue
		}

		// Enhance phone number formatting
		if contact.Phone != "" {
			contact.Phone = s.enhancePhoneFormat(contact.Phone)
		}

		// Clean up name
		if contact.Name != "" {
			contact.Name = s.enhanceName(contact.Name)
		}

		// Skip if no phone or name
		if contact.Phone == "" && contact.Name == "" {
			continue
		}

		validated = append(validated, contact)
	}

	return validated
}

// enhancePhoneFormat improves phone number formatting
func (s *ContactDetectionService) enhancePhoneFormat(phone string) string {
	// Remove all formatting
	clean := regexp.MustCompile(`[^\d+]`).ReplaceAllString(phone, "")

	// If no + and more than 10 digits, likely international
	if !strings.HasPrefix(clean, "+") && len(clean) > 10 {
		clean = "+" + clean
	}

	return clean
}

// enhanceName improves name formatting
func (s *ContactDetectionService) enhanceName(name string) string {
	// Title case each word
	words := strings.Fields(strings.ToLower(name))
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + word[1:]
		}
	}

	return strings.Join(words, " ")
}

// GetConfidenceExplanation returns explanation for confidence score
func (s *ContactDetectionService) GetConfidenceExplanation(contact models.SmartContactDetection) string {
	explanations := []string{}

	if contact.Phone != "" {
		explanations = append(explanations, "Phone number detected")
	}
	if contact.Name != "" {
		explanations = append(explanations, "Name detected")
	}
	if contact.Email != "" {
		explanations = append(explanations, "Email detected")
	}

	confidence := "Low"
	if contact.Confidence >= 0.7 {
		confidence = "High"
	} else if contact.Confidence >= 0.5 {
		confidence = "Medium"
	}

	return fmt.Sprintf("%s confidence: %s", confidence, strings.Join(explanations, ", "))
}
