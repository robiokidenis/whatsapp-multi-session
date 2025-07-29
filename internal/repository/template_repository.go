package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"whatsapp-multi-session/internal/models"
)

type TemplateRepository struct {
	db *sql.DB
}

func NewTemplateRepository(db *sql.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// CreateTemplate creates a new message template
func (r *TemplateRepository) CreateTemplate(template *models.MessageTemplate) error {
	variablesJSON, _ := json.Marshal(template.Variables)
	
	query := `
		INSERT INTO message_templates (name, content, type, variables, media_url, media_type, category, is_active, usage_count, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	result, err := r.db.Exec(query,
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

// GetTemplate retrieves a template by ID
func (r *TemplateRepository) GetTemplate(id int) (*models.MessageTemplate, error) {
	template := &models.MessageTemplate{}
	var variablesJSON sql.NullString
	var updatedAt sql.NullInt64
	var createdAt int64
	
	query := `
		SELECT id, name, content, type, variables, media_url, media_type, category, 
		       is_active, usage_count, created_at, updated_at
		FROM message_templates
		WHERE id = ?`
	
	err := r.db.QueryRow(query, id).Scan(
		&template.ID,
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

// GetTemplates retrieves templates with filtering
func (r *TemplateRepository) GetTemplates(category, templateType string, isActive *bool) ([]models.MessageTemplate, error) {
	query := `
		SELECT id, name, content, type, variables, media_url, media_type, category, 
		       is_active, usage_count, created_at, updated_at
		FROM message_templates
		WHERE 1=1`
	
	args := []interface{}{}
	
	if category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}
	
	if templateType != "" {
		query += " AND type = ?"
		args = append(args, templateType)
	}
	
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
func (r *TemplateRepository) GetActiveTemplates() ([]models.MessageTemplate, error) {
	isActive := true
	return r.GetTemplates("", "", &isActive)
}

// UpdateTemplate updates an existing template
func (r *TemplateRepository) UpdateTemplate(id int, req models.UpdateTemplateRequest) error {
	setParts := []string{}
	args := []interface{}{}
	
	if req.Name != "" {
		setParts = append(setParts, "name = ?")
		args = append(args, req.Name)
	}
	
	if req.Content != "" {
		setParts = append(setParts, "content = ?")
		args = append(args, req.Content)
	}
	
	if req.Type != "" {
		setParts = append(setParts, "type = ?")
		args = append(args, req.Type)
	}
	
	if req.Variables != nil {
		variablesJSON, _ := json.Marshal(req.Variables)
		setParts = append(setParts, "variables = ?")
		args = append(args, string(variablesJSON))
	}
	
	if req.MediaURL != "" {
		setParts = append(setParts, "media_url = ?")
		args = append(args, req.MediaURL)
	}
	
	if req.MediaType != "" {
		setParts = append(setParts, "media_type = ?")
		args = append(args, req.MediaType)
	}
	
	if req.Category != "" {
		setParts = append(setParts, "category = ?")
		args = append(args, req.Category)
	}
	
	if req.IsActive != nil {
		setParts = append(setParts, "is_active = ?")
		args = append(args, *req.IsActive)
	}
	
	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}
	
	setParts = append(setParts, "updated_at = ?")
	args = append(args, time.Now().Unix())
	args = append(args, id)
	
	query := fmt.Sprintf("UPDATE message_templates SET %s WHERE id = ?", strings.Join(setParts, ", "))
	
	result, err := r.db.Exec(query, args...)
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

// DeleteTemplate deletes a template
func (r *TemplateRepository) DeleteTemplate(id int) error {
	// Check if template is used in any campaigns
	var campaignCount int
	err := r.db.QueryRow("SELECT COUNT(*) FROM campaigns WHERE template_id = ?", id).Scan(&campaignCount)
	if err != nil {
		return fmt.Errorf("failed to check campaign usage: %v", err)
	}
	
	if campaignCount > 0 {
		return fmt.Errorf("cannot delete template used in %d campaigns", campaignCount)
	}
	
	query := "DELETE FROM message_templates WHERE id = ?"
	
	result, err := r.db.Exec(query, id)
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
func (r *TemplateRepository) GetTemplateCategories() ([]string, error) {
	query := `
		SELECT DISTINCT category 
		FROM message_templates 
		WHERE category IS NOT NULL AND category != '' 
		ORDER BY category`
	
	rows, err := r.db.Query(query)
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