package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"whatsapp-multi-session/internal/models"
)

type TemplateRepository struct {
	db *sql.DB
}

func NewTemplateRepository(db *sql.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// CreateMessageTemplate creates a new message template
func (r *TemplateRepository) CreateMessageTemplate(template *models.MessageTemplate) error {
	variablesJSON, _ := json.Marshal(template.Variables)
	
	query := `
		INSERT INTO message_templates (user_id, name, content, type, variables, media_url, media_type, category, is_active, usage_count, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	result, err := r.db.Exec(query,
		template.UserID,
		template.Name,
		template.Content,
		template.Type,
		string(variablesJSON),
		template.MediaURL,
		template.MediaType,
		template.Category,
		template.IsActive,
		0, // initial usage count
		time.Now().Unix(),
	)
	
	if err != nil {
		return fmt.Errorf("failed to create template: %v", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get template ID: %v", err)
	}
	
	template.ID = int(id)
	template.CreatedAt = time.Now()
	template.UsageCount = 0
	
	return nil
}

// GetMessageTemplate retrieves a template by ID  
func (r *TemplateRepository) GetMessageTemplate(userID, id int) (*models.MessageTemplate, error) {
	template := &models.MessageTemplate{}
	var variablesJSON sql.NullString
	var updatedAt sql.NullInt64
	var createdAt int64
	
	query := `
		SELECT id, user_id, name, content, type, variables, media_url, media_type, category, 
		       is_active, usage_count, created_at, updated_at
		FROM message_templates
		WHERE id = ? AND user_id = ?`
	
	err := r.db.QueryRow(query, id, userID).Scan(
		&template.ID,
		&template.UserID,
		&template.Name,
		&template.Content,
		&template.Type,
		&variablesJSON,
		&template.MediaURL,
		&template.MediaType,
		&template.Category,
		&template.IsActive,
		&template.UsageCount,
		&createdAt,
		&updatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("template not found")
		}
		return nil, fmt.Errorf("failed to get template: %v", err)
	}
	
	// Parse timestamps
	template.CreatedAt = time.Unix(createdAt, 0)
	if updatedAt.Valid {
		t := time.Unix(updatedAt.Int64, 0)
		template.UpdatedAt = &t
	}
	
	// Parse variables
	if variablesJSON.Valid && variablesJSON.String != "" {
		json.Unmarshal([]byte(variablesJSON.String), &template.Variables)
	}
	
	return template, nil
}

// GetMessageTemplates retrieves templates with filtering
func (r *TemplateRepository) GetMessageTemplates(userID int, isActive *bool) ([]models.MessageTemplate, error) {
	query := `
		SELECT id, user_id, name, content, type, variables, media_url, media_type, category, 
		       is_active, usage_count, created_at, updated_at
		FROM message_templates
		WHERE user_id = ?`
	
	args := []interface{}{userID}
	
	if isActive != nil {
		query += " AND is_active = ?"
		args = append(args, *isActive)
	}
	
	query += " ORDER BY name"
	
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query templates: %v", err)
	}
	defer rows.Close()
	
	var templates []models.MessageTemplate
	
	for rows.Next() {
		template := models.MessageTemplate{}
		var variablesJSON sql.NullString
		var updatedAt sql.NullInt64
		var createdAt int64
		
		err := rows.Scan(
			&template.ID,
			&template.UserID,
			&template.Name,
			&template.Content,
			&template.Type,
			&variablesJSON,
			&template.MediaURL,
			&template.MediaType,
			&template.Category,
			&template.IsActive,
			&template.UsageCount,
			&createdAt,
			&updatedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan template: %v", err)
		}
		
		// Parse timestamps
		template.CreatedAt = time.Unix(createdAt, 0)
		if updatedAt.Valid {
			t := time.Unix(updatedAt.Int64, 0)
			template.UpdatedAt = &t
		}
		
		// Parse variables
		if variablesJSON.Valid && variablesJSON.String != "" {
			json.Unmarshal([]byte(variablesJSON.String), &template.Variables)
		}
		
		templates = append(templates, template)
	}
	
	return templates, nil
}

// GetActiveTemplates retrieves only active templates
func (r *TemplateRepository) GetActiveTemplates(userID int) ([]models.MessageTemplate, error) {
	isActive := true
	return r.GetMessageTemplates(userID, &isActive)
}

// UpdateMessageTemplate updates an existing template
func (r *TemplateRepository) UpdateMessageTemplate(template *models.MessageTemplate) error {
	variablesJSON, _ := json.Marshal(template.Variables)
	
	query := `
		UPDATE message_templates 
		SET name = ?, content = ?, type = ?, variables = ?, media_url = ?, media_type = ?, 
		    category = ?, is_active = ?, updated_at = ?
		WHERE id = ? AND user_id = ?`
	
	result, err := r.db.Exec(query,
		template.Name,
		template.Content,
		template.Type,
		string(variablesJSON),
		template.MediaURL,
		template.MediaType,
		template.Category,
		template.IsActive,
		time.Now().Unix(),
		template.ID,
		template.UserID,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update template: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("template not found")
	}
	
	return nil
}

// DeleteMessageTemplate deletes a template
func (r *TemplateRepository) DeleteMessageTemplate(userID, id int) error {
	// Check if template is used in any campaigns
	var campaignCount int
	err := r.db.QueryRow("SELECT COUNT(*) FROM campaigns WHERE template_id = ?", id).Scan(&campaignCount)
	if err != nil {
		return fmt.Errorf("failed to check campaign usage: %v", err)
	}
	
	if campaignCount > 0 {
		return fmt.Errorf("cannot delete template used in %d campaigns", campaignCount)
	}
	
	query := "DELETE FROM message_templates WHERE id = ? AND user_id = ?"
	
	result, err := r.db.Exec(query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete template: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("template not found")
	}
	
	return nil
}

// IncrementUsageCount increments the usage count for a template
func (r *TemplateRepository) IncrementUsageCount(id int) error {
	query := "UPDATE message_templates SET usage_count = usage_count + 1, updated_at = ? WHERE id = ?"
	
	result, err := r.db.Exec(query, time.Now().Unix(), id)
	if err != nil {
		return fmt.Errorf("failed to increment usage count: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("template not found")
	}
	
	return nil
}

// CheckTemplateNameExists checks if a template name already exists
func (r *TemplateRepository) CheckTemplateNameExists(name string, excludeID *int) (bool, error) {
	query := "SELECT COUNT(*) FROM message_templates WHERE name = ?"
	args := []interface{}{name}
	
	if excludeID != nil {
		query += " AND id != ?"
		args = append(args, *excludeID)
	}
	
	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check template name: %v", err)
	}
	
	return count > 0, nil
}

// GetTemplateCategories returns all unique template categories
func (r *TemplateRepository) GetTemplateCategories(userID int) ([]string, error) {
	query := `
		SELECT DISTINCT category 
		FROM message_templates 
		WHERE user_id = ? AND category IS NOT NULL AND category != ''
		ORDER BY category`
	
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %v", err)
	}
	defer rows.Close()
	
	var categories []string
	
	for rows.Next() {
		var category string
		err := rows.Scan(&category)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %v", err)
		}
		categories = append(categories, category)
	}
	
	return categories, nil
}

// GetTemplateTypes returns all unique template types
func (r *TemplateRepository) GetTemplateTypes() ([]string, error) {
	query := `
		SELECT DISTINCT type 
		FROM message_templates 
		ORDER BY type`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query types: %v", err)
	}
	defer rows.Close()
	
	var types []string
	
	for rows.Next() {
		var templateType string
		err := rows.Scan(&templateType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan type: %v", err)
		}
		types = append(types, templateType)
	}
	
	return types, nil
}

// GetTemplateStats returns statistics for templates
func (r *TemplateRepository) GetTemplateStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Total templates
	var totalTemplates, activeTemplates int
	err := r.db.QueryRow("SELECT COUNT(*), SUM(CASE WHEN is_active THEN 1 ELSE 0 END) FROM message_templates").Scan(&totalTemplates, &activeTemplates)
	if err != nil {
		return nil, fmt.Errorf("failed to get template counts: %v", err)
	}
	
	stats["total_templates"] = totalTemplates
	stats["active_templates"] = activeTemplates
	
	// Templates by type
	typeQuery := `
		SELECT type, COUNT(*) as count
		FROM message_templates
		GROUP BY type
		ORDER BY count DESC`
	
	rows, err := r.db.Query(typeQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query templates by type: %v", err)
	}
	defer rows.Close()
	
	typeStats := make(map[string]int)
	for rows.Next() {
		var templateType string
		var count int
		err := rows.Scan(&templateType, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan type stats: %v", err)
		}
		typeStats[templateType] = count
	}
	stats["templates_by_type"] = typeStats
	
	// Total usage count
	var totalUsage int
	err = r.db.QueryRow("SELECT SUM(usage_count) FROM message_templates").Scan(&totalUsage)
	if err != nil {
		return nil, fmt.Errorf("failed to get total usage: %v", err)
	}
	
	stats["total_usage"] = totalUsage
	
	// Most used templates
	mostUsedQuery := `
		SELECT id, name, usage_count
		FROM message_templates
		ORDER BY usage_count DESC
		LIMIT 5`
	
	rows, err = r.db.Query(mostUsedQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query most used templates: %v", err)
	}
	defer rows.Close()
	
	type TemplateUsage struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		UsageCount int    `json:"usage_count"`
	}
	
	var mostUsed []TemplateUsage
	for rows.Next() {
		var usage TemplateUsage
		err := rows.Scan(&usage.ID, &usage.Name, &usage.UsageCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan most used template: %v", err)
		}
		mostUsed = append(mostUsed, usage)
	}
	stats["most_used_templates"] = mostUsed
	
	return stats, nil
}