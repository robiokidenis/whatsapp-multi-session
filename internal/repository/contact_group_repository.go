package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"whatsapp-multi-session/internal/models"
)

type ContactGroupRepository struct {
	db *sql.DB
}

func NewContactGroupRepository(db *sql.DB) *ContactGroupRepository {
	return &ContactGroupRepository{db: db}
}

// CreateContactGroup creates a new contact group
func (r *ContactGroupRepository) CreateContactGroup(group *models.ContactGroup) error {
	query := `
		INSERT INTO contact_groups (name, description, color, is_active, created_at)
		VALUES (?, ?, ?, ?, ?)`
	
	result, err := r.db.Exec(query,
		group.Name,
		group.Description,
		group.Color,
		group.IsActive,
		time.Now().Unix(),
	)
	
	if err != nil {
		return fmt.Errorf("failed to create contact group: %v", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get contact group ID: %v", err)
	}
	
	group.ID = int(id)
	group.CreatedAt = time.Now()
	
	return nil
}

// GetContactGroup retrieves a contact group by ID
func (r *ContactGroupRepository) GetContactGroup(id int) (*models.ContactGroup, error) {
	group := &models.ContactGroup{}
	var updatedAt sql.NullInt64
	var createdAt int64
	
	query := `
		SELECT cg.id, cg.name, cg.description, cg.color, cg.is_active, cg.created_at, cg.updated_at,
		       COUNT(c.id) as contact_count
		FROM contact_groups cg
		LEFT JOIN contacts c ON cg.id = c.group_id AND c.is_active = true
		WHERE cg.id = ?
		GROUP BY cg.id`
	
	err := r.db.QueryRow(query, id).Scan(
		&group.ID,
		&group.Name,
		&group.Description,
		&group.Color,
		&group.IsActive,
		&createdAt,
		&updatedAt,
		&group.ContactCount,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("contact group not found")
		}
		return nil, fmt.Errorf("failed to get contact group: %v", err)
	}
	
	// Parse timestamps
	group.CreatedAt = time.Unix(createdAt, 0)
	if updatedAt.Valid {
		t := time.Unix(updatedAt.Int64, 0)
		group.UpdatedAt = &t
	}
	
	return group, nil
}

// GetContactGroups retrieves all contact groups
func (r *ContactGroupRepository) GetContactGroups() ([]models.ContactGroup, error) {
	query := `
		SELECT cg.id, cg.name, cg.description, cg.color, cg.is_active, cg.created_at, cg.updated_at,
		       COUNT(c.id) as contact_count
		FROM contact_groups cg
		LEFT JOIN contacts c ON cg.id = c.group_id AND c.is_active = true
		GROUP BY cg.id
		ORDER BY cg.name`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query contact groups: %v", err)
	}
	defer rows.Close()
	
	var groups []models.ContactGroup
	
	for rows.Next() {
		group := models.ContactGroup{}
		var updatedAt sql.NullInt64
		var createdAt int64
		
		err := rows.Scan(
			&group.ID,
			&group.Name,
			&group.Description,
			&group.Color,
			&group.IsActive,
			&createdAt,
			&updatedAt,
			&group.ContactCount,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan contact group: %v", err)
		}
		
		// Parse timestamps
		group.CreatedAt = time.Unix(createdAt, 0)
		if updatedAt.Valid {
			t := time.Unix(updatedAt.Int64, 0)
			group.UpdatedAt = &t
		}
		
		groups = append(groups, group)
	}
	
	return groups, nil
}

// GetActiveContactGroups retrieves only active contact groups
func (r *ContactGroupRepository) GetActiveContactGroups() ([]models.ContactGroup, error) {
	query := `
		SELECT cg.id, cg.name, cg.description, cg.color, cg.is_active, cg.created_at, cg.updated_at,
		       COUNT(c.id) as contact_count
		FROM contact_groups cg
		LEFT JOIN contacts c ON cg.id = c.group_id AND c.is_active = true
		WHERE cg.is_active = true
		GROUP BY cg.id
		ORDER BY cg.name`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active contact groups: %v", err)
	}
	defer rows.Close()
	
	var groups []models.ContactGroup
	
	for rows.Next() {
		group := models.ContactGroup{}
		var updatedAt sql.NullInt64
		var createdAt int64
		
		err := rows.Scan(
			&group.ID,
			&group.Name,
			&group.Description,
			&group.Color,
			&group.IsActive,
			&createdAt,
			&updatedAt,
			&group.ContactCount,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan contact group: %v", err)
		}
		
		// Parse timestamps
		group.CreatedAt = time.Unix(createdAt, 0)
		if updatedAt.Valid {
			t := time.Unix(updatedAt.Int64, 0)
			group.UpdatedAt = &t
		}
		
		groups = append(groups, group)
	}
	
	return groups, nil
}

// UpdateContactGroup updates an existing contact group
func (r *ContactGroupRepository) UpdateContactGroup(id int, req models.UpdateContactGroupRequest) error {
	setParts := []string{}
	args := []interface{}{}
	
	if req.Name != "" {
		setParts = append(setParts, "name = ?")
		args = append(args, req.Name)
	}
	
	if req.Description != "" {
		setParts = append(setParts, "description = ?")
		args = append(args, req.Description)
	}
	
	if req.Color != "" {
		setParts = append(setParts, "color = ?")
		args = append(args, req.Color)
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
	
	query := fmt.Sprintf("UPDATE contact_groups SET %s WHERE id = ?", strings.Join(setParts, ", "))
	
	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update contact group: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("contact group not found")
	}
	
	return nil
}

// DeleteContactGroup deletes a contact group
func (r *ContactGroupRepository) DeleteContactGroup(id int) error {
	// Check if group has contacts
	var contactCount int
	err := r.db.QueryRow("SELECT COUNT(*) FROM contacts WHERE group_id = ?", id).Scan(&contactCount)
	if err != nil {
		return fmt.Errorf("failed to check contact count: %v", err)
	}
	
	if contactCount > 0 {
		return fmt.Errorf("cannot delete group with %d contacts. Move or delete contacts first", contactCount)
	}
	
	query := "DELETE FROM contact_groups WHERE id = ?"
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete contact group: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("contact group not found")
	}
	
	return nil
}

// CheckGroupNameExists checks if a group name already exists
func (r *ContactGroupRepository) CheckGroupNameExists(name string, excludeID *int) (bool, error) {
	query := "SELECT COUNT(*) FROM contact_groups WHERE name = ?"
	args := []interface{}{name}
	
	if excludeID != nil {
		query += " AND id != ?"
		args = append(args, *excludeID)
	}
	
	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check group name: %v", err)
	}
	
	return count > 0, nil
}

// GetGroupStats returns statistics for contact groups
func (r *ContactGroupRepository) GetGroupStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Total groups
	var totalGroups, activeGroups int
	err := r.db.QueryRow("SELECT COUNT(*), SUM(CASE WHEN is_active THEN 1 ELSE 0 END) FROM contact_groups").Scan(&totalGroups, &activeGroups)
	if err != nil {
		return nil, fmt.Errorf("failed to get group counts: %v", err)
	}
	
	stats["total_groups"] = totalGroups
	stats["active_groups"] = activeGroups
	
	// Total contacts in groups
	var contactsInGroups int
	err = r.db.QueryRow("SELECT COUNT(*) FROM contacts WHERE group_id IS NOT NULL AND is_active = true").Scan(&contactsInGroups)
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts in groups count: %v", err)
	}
	
	stats["contacts_in_groups"] = contactsInGroups
	
	// Contacts without groups
	var contactsWithoutGroup int
	err = r.db.QueryRow("SELECT COUNT(*) FROM contacts WHERE group_id IS NULL AND is_active = true").Scan(&contactsWithoutGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts without group count: %v", err)
	}
	
	stats["contacts_without_group"] = contactsWithoutGroup
	
	// Average contacts per group
	if activeGroups > 0 {
		stats["avg_contacts_per_group"] = float64(contactsInGroups) / float64(activeGroups)
	} else {
		stats["avg_contacts_per_group"] = 0.0
	}
	
	return stats, nil
}