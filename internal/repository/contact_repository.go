package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"whatsapp-multi-session/internal/models"
)

type ContactRepository struct {
	db *sql.DB
}

func NewContactRepository(db *sql.DB) *ContactRepository {
	return &ContactRepository{db: db}
}

// CreateContact creates a new contact
func (r *ContactRepository) CreateContact(contact *models.Contact) error {
	tagsJSON, _ := json.Marshal(contact.Tags)
	
	query := `
		INSERT INTO contacts (name, phone, email, company, position, group_id, tags, notes, is_active, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	result, err := r.db.Exec(query,
		contact.Name,
		contact.Phone,
		contact.Email,
		contact.Company,
		contact.Position,
		contact.GroupID,
		string(tagsJSON),
		contact.Notes,
		contact.IsActive,
		time.Now().Unix(),
	)
	
	if err != nil {
		return fmt.Errorf("failed to create contact: %v", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get contact ID: %v", err)
	}
	
	contact.ID = int(id)
	contact.CreatedAt = time.Now()
	
	return nil
}

// GetContact retrieves a contact by ID
func (r *ContactRepository) GetContact(id int) (*models.Contact, error) {
	contact := &models.Contact{}
	var tagsJSON, groupName, groupColor sql.NullString
	var lastContact sql.NullInt64
	var updatedAt sql.NullInt64
	var createdAt int64
	
	query := `
		SELECT c.id, c.name, c.phone, c.email, c.company, c.position, c.group_id, c.tags, 
		       c.notes, c.is_active, c.last_contact, c.created_at, c.updated_at,
		       cg.name as group_name, cg.color as group_color
		FROM contacts c
		LEFT JOIN contact_groups cg ON c.group_id = cg.id
		WHERE c.id = ?`
	
	err := r.db.QueryRow(query, id).Scan(
		&contact.ID,
		&contact.Name,
		&contact.Phone,
		&contact.Email,
		&contact.Company,
		&contact.Position,
		&contact.GroupID,
		&tagsJSON,
		&contact.Notes,
		&contact.IsActive,
		&lastContact,
		&createdAt,
		&updatedAt,
		&groupName,
		&groupColor,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("contact not found")
		}
		return nil, fmt.Errorf("failed to get contact: %v", err)
	}
	
	// Parse timestamps
	contact.CreatedAt = time.Unix(createdAt, 0)
	if updatedAt.Valid {
		t := time.Unix(updatedAt.Int64, 0)
		contact.UpdatedAt = &t
	}
	if lastContact.Valid {
		t := time.Unix(lastContact.Int64, 0)
		contact.LastContact = &t
	}
	
	// Parse tags
	if tagsJSON.Valid && tagsJSON.String != "" {
		json.Unmarshal([]byte(tagsJSON.String), &contact.Tags)
	}
	
	// Set group info if exists
	if contact.GroupID != nil && groupName.Valid {
		contact.Group = &models.ContactGroup{
			ID:    *contact.GroupID,
			Name:  groupName.String,
			Color: groupColor.String,
		}
	}
	
	return contact, nil
}

// GetContacts retrieves contacts with filtering and pagination
func (r *ContactRepository) GetContacts(req models.ContactSearchRequest) (*models.ContactListResponse, error) {
	// Build base query
	baseQuery := `
		FROM contacts c
		LEFT JOIN contact_groups cg ON c.group_id = cg.id
		WHERE 1=1`
	
	args := []interface{}{}
	argIndex := 1
	
	// Add filters
	if req.Query != "" {
		baseQuery += fmt.Sprintf(" AND (c.name LIKE ? OR c.phone LIKE ? OR c.email LIKE ? OR c.company LIKE ?)")
		likeQuery := "%" + req.Query + "%"
		args = append(args, likeQuery, likeQuery, likeQuery, likeQuery)
		argIndex += 4
	}
	
	if req.GroupID != nil {
		baseQuery += fmt.Sprintf(" AND c.group_id = ?")
		args = append(args, *req.GroupID)
		argIndex++
	}
	
	if req.IsActive != nil {
		baseQuery += fmt.Sprintf(" AND c.is_active = ?")
		args = append(args, *req.IsActive)
		argIndex++
	}
	
	if len(req.Tags) > 0 {
		// For SQLite/MySQL JSON search - simplified approach
		for _, tag := range req.Tags {
			baseQuery += fmt.Sprintf(" AND c.tags LIKE ?")
			args = append(args, "%"+tag+"%")
			argIndex++
		}
	}
	
	// Count total records
	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count contacts: %v", err)
	}
	
	// Set pagination defaults
	page := req.Page
	if page <= 0 {
		page = 1
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	
	offset := (page - 1) * limit
	
	// Build main query with pagination
	selectQuery := `
		SELECT c.id, c.name, c.phone, c.email, c.company, c.position, c.group_id, c.tags,
		       c.notes, c.is_active, c.last_contact, c.created_at, c.updated_at,
		       cg.name as group_name, cg.color as group_color
		` + baseQuery + `
		ORDER BY c.created_at DESC
		LIMIT ? OFFSET ?`
	
	args = append(args, limit, offset)
	
	rows, err := r.db.Query(selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query contacts: %v", err)
	}
	defer rows.Close()
	
	var contacts []models.Contact
	
	for rows.Next() {
		contact := models.Contact{}
		var tagsJSON, groupName, groupColor sql.NullString
		var lastContact sql.NullInt64
		var updatedAt sql.NullInt64
		var createdAt int64
		
		err := rows.Scan(
			&contact.ID,
			&contact.Name,
			&contact.Phone,
			&contact.Email,
			&contact.Company,
			&contact.Position,
			&contact.GroupID,
			&tagsJSON,
			&contact.Notes,
			&contact.IsActive,
			&lastContact,
			&createdAt,
			&updatedAt,
			&groupName,
			&groupColor,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan contact: %v", err)
		}
		
		// Parse timestamps
		contact.CreatedAt = time.Unix(createdAt, 0)
		if updatedAt.Valid {
			t := time.Unix(updatedAt.Int64, 0)
			contact.UpdatedAt = &t
		}
		if lastContact.Valid {
			t := time.Unix(lastContact.Int64, 0)
			contact.LastContact = &t
		}
		
		// Parse tags
		if tagsJSON.Valid && tagsJSON.String != "" {
			json.Unmarshal([]byte(tagsJSON.String), &contact.Tags)
		}
		
		// Set group info if exists
		if contact.GroupID != nil && groupName.Valid {
			contact.Group = &models.ContactGroup{
				ID:    *contact.GroupID,
				Name:  groupName.String,
				Color: groupColor.String,
			}
		}
		
		contacts = append(contacts, contact)
	}
	
	pages := (total + limit - 1) / limit
	
	return &models.ContactListResponse{
		Contacts: contacts,
		Total:    total,
		Page:     page,
		Limit:    limit,
		Pages:    pages,
	}, nil
}

// UpdateContact updates an existing contact
func (r *ContactRepository) UpdateContact(id int, req models.UpdateContactRequest) error {
	setParts := []string{}
	args := []interface{}{}
	
	if req.Name != "" {
		setParts = append(setParts, "name = ?")
		args = append(args, req.Name)
	}
	
	if req.Phone != "" {
		setParts = append(setParts, "phone = ?")
		args = append(args, req.Phone)
	}
	
	if req.Email != "" {
		setParts = append(setParts, "email = ?")
		args = append(args, req.Email)
	}
	
	if req.Company != "" {
		setParts = append(setParts, "company = ?")
		args = append(args, req.Company)
	}
	
	if req.Position != "" {
		setParts = append(setParts, "position = ?")
		args = append(args, req.Position)
	}
	
	if req.GroupID != nil {
		setParts = append(setParts, "group_id = ?")
		args = append(args, req.GroupID)
	}
	
	if req.Tags != nil {
		tagsJSON, _ := json.Marshal(req.Tags)
		setParts = append(setParts, "tags = ?")
		args = append(args, string(tagsJSON))
	}
	
	if req.Notes != "" {
		setParts = append(setParts, "notes = ?")
		args = append(args, req.Notes)
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
	
	query := fmt.Sprintf("UPDATE contacts SET %s WHERE id = ?", strings.Join(setParts, ", "))
	
	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update contact: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("contact not found")
	}
	
	return nil
}

// DeleteContact deletes a contact
func (r *ContactRepository) DeleteContact(id int) error {
	query := "DELETE FROM contacts WHERE id = ?"
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete contact: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("contact not found")
	}
	
	return nil
}

// BulkCreateContacts creates multiple contacts in a transaction
func (r *ContactRepository) BulkCreateContacts(contacts []models.Contact) (*models.ContactImportResult, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()
	
	result := &models.ContactImportResult{
		Total:  len(contacts),
		Errors: []string{},
	}
	
	query := `
		INSERT INTO contacts (name, phone, email, company, position, group_id, tags, notes, is_active, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()
	
	for i, contact := range contacts {
		// Check for duplicate phone number
		var existingID int
		checkQuery := "SELECT id FROM contacts WHERE phone = ?"
		err := tx.QueryRow(checkQuery, contact.Phone).Scan(&existingID)
		if err == nil {
			result.Duplicates++
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: Phone %s already exists", i+1, contact.Phone))
			continue
		} else if err != sql.ErrNoRows {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: Database error checking duplicate: %v", i+1, err))
			continue
		}
		
		tagsJSON, _ := json.Marshal(contact.Tags)
		
		_, err = stmt.Exec(
			contact.Name,
			contact.Phone,
			contact.Email,
			contact.Company,
			contact.Position,
			contact.GroupID,
			string(tagsJSON),
			contact.Notes,
			contact.IsActive,
			time.Now().Unix(),
		)
		
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: %v", i+1, err))
		} else {
			result.Success++
		}
	}
	
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}
	
	return result, nil
}

// BulkUpdateContacts performs bulk operations on contacts
func (r *ContactRepository) BulkUpdateContacts(req models.BulkContactRequest) error {
	if len(req.ContactIDs) == 0 {
		return fmt.Errorf("no contact IDs provided")
	}
	
	placeholders := strings.Repeat("?,", len(req.ContactIDs))
	placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma
	
	args := make([]interface{}, len(req.ContactIDs))
	for i, id := range req.ContactIDs {
		args[i] = id
	}
	
	var query string
	
	switch req.Action {
	case "delete":
		query = fmt.Sprintf("DELETE FROM contacts WHERE id IN (%s)", placeholders)
		
	case "activate":
		query = fmt.Sprintf("UPDATE contacts SET is_active = true, updated_at = ? WHERE id IN (%s)", placeholders)
		args = append([]interface{}{time.Now().Unix()}, args...)
		
	case "deactivate":
		query = fmt.Sprintf("UPDATE contacts SET is_active = false, updated_at = ? WHERE id IN (%s)", placeholders)
		args = append([]interface{}{time.Now().Unix()}, args...)
		
	case "move_to_group":
		if req.GroupID == nil {
			return fmt.Errorf("group_id is required for move_to_group action")
		}
		query = fmt.Sprintf("UPDATE contacts SET group_id = ?, updated_at = ? WHERE id IN (%s)", placeholders)
		args = append([]interface{}{*req.GroupID, time.Now().Unix()}, args...)
		
	default:
		return fmt.Errorf("invalid action: %s", req.Action)
	}
	
	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute bulk operation: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("no contacts were affected")
	}
	
	return nil
}

// GetContactsByIDs retrieves multiple contacts by their IDs
func (r *ContactRepository) GetContactsByIDs(ids []int) ([]models.Contact, error) {
	if len(ids) == 0 {
		return []models.Contact{}, nil
	}
	
	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1]
	
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	
	query := fmt.Sprintf(`
		SELECT c.id, c.name, c.phone, c.email, c.company, c.position, c.group_id, c.tags,
		       c.notes, c.is_active, c.last_contact, c.created_at, c.updated_at,
		       cg.name as group_name, cg.color as group_color
		FROM contacts c
		LEFT JOIN contact_groups cg ON c.group_id = cg.id
		WHERE c.id IN (%s)`, placeholders)
	
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query contacts: %v", err)
	}
	defer rows.Close()
	
	var contacts []models.Contact
	
	for rows.Next() {
		contact := models.Contact{}
		var tagsJSON, groupName, groupColor sql.NullString
		var lastContact sql.NullInt64
		var updatedAt sql.NullInt64
		var createdAt int64
		
		err := rows.Scan(
			&contact.ID,
			&contact.Name,
			&contact.Phone,
			&contact.Email,
			&contact.Company,
			&contact.Position,
			&contact.GroupID,
			&tagsJSON,
			&contact.Notes,
			&contact.IsActive,
			&lastContact,
			&createdAt,
			&updatedAt,
			&groupName,
			&groupColor,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan contact: %v", err)
		}
		
		// Parse timestamps
		contact.CreatedAt = time.Unix(createdAt, 0)
		if updatedAt.Valid {
			t := time.Unix(updatedAt.Int64, 0)
			contact.UpdatedAt = &t
		}
		if lastContact.Valid {
			t := time.Unix(lastContact.Int64, 0)
			contact.LastContact = &t
		}
		
		// Parse tags
		if tagsJSON.Valid && tagsJSON.String != "" {
			json.Unmarshal([]byte(tagsJSON.String), &contact.Tags)
		}
		
		// Set group info if exists
		if contact.GroupID != nil && groupName.Valid {
			contact.Group = &models.ContactGroup{
				ID:    *contact.GroupID,
				Name:  groupName.String,
				Color: groupColor.String,
			}
		}
		
		contacts = append(contacts, contact)
	}
	
	return contacts, nil
}

// GetContactsByGroupID retrieves all contacts in a specific group
func (r *ContactRepository) GetContactsByGroupID(groupID int) ([]models.Contact, error) {
	query := `
		SELECT c.id, c.name, c.phone, c.email, c.company, c.position, c.group_id, c.tags,
		       c.notes, c.is_active, c.last_contact, c.created_at, c.updated_at,
		       cg.name as group_name, cg.color as group_color
		FROM contacts c
		LEFT JOIN contact_groups cg ON c.group_id = cg.id
		WHERE c.group_id = ? AND c.is_active = true
		ORDER BY c.name`
	
	rows, err := r.db.Query(query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to query contacts by group: %v", err)
	}
	defer rows.Close()
	
	var contacts []models.Contact
	
	for rows.Next() {
		contact := models.Contact{}
		var tagsJSON, groupName, groupColor sql.NullString
		var lastContact sql.NullInt64
		var updatedAt sql.NullInt64
		var createdAt int64
		
		err := rows.Scan(
			&contact.ID,
			&contact.Name,
			&contact.Phone,
			&contact.Email,
			&contact.Company,
			&contact.Position,
			&contact.GroupID,
			&tagsJSON,
			&contact.Notes,
			&contact.IsActive,
			&lastContact,
			&createdAt,
			&updatedAt,
			&groupName,
			&groupColor,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan contact: %v", err)
		}
		
		// Parse timestamps
		contact.CreatedAt = time.Unix(createdAt, 0)
		if updatedAt.Valid {
			t := time.Unix(updatedAt.Int64, 0)
			contact.UpdatedAt = &t
		}
		if lastContact.Valid {
			t := time.Unix(lastContact.Int64, 0)
			contact.LastContact = &t
		}
		
		// Parse tags
		if tagsJSON.Valid && tagsJSON.String != "" {
			json.Unmarshal([]byte(tagsJSON.String), &contact.Tags)
		}
		
		// Set group info if exists
		if contact.GroupID != nil && groupName.Valid {
			contact.Group = &models.ContactGroup{
				ID:    *contact.GroupID,
				Name:  groupName.String,
				Color: groupColor.String,
			}
		}
		
		contacts = append(contacts, contact)
	}
	
	return contacts, nil
}

// UpdateLastContact updates the last contact time for a contact
func (r *ContactRepository) UpdateLastContact(id int) error {
	query := "UPDATE contacts SET last_contact = ?, updated_at = ? WHERE id = ?"
	
	now := time.Now().Unix()
	result, err := r.db.Exec(query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to update last contact: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("contact not found")
	}
	
	return nil
}