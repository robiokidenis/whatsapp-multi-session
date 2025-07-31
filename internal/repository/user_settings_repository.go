package repository

import (
	"database/sql"
	"fmt"
	"time"

	"whatsapp-multi-session/internal/models"
)

// UserSettingsRepository handles user settings database operations
type UserSettingsRepository struct {
	db *sql.DB
}

// NewUserSettingsRepository creates a new user settings repository
func NewUserSettingsRepository(db *sql.DB) *UserSettingsRepository {
	return &UserSettingsRepository{db: db}
}

// GetByUserID retrieves user settings by user ID
func (r *UserSettingsRepository) GetByUserID(userID int) (*models.UserSettings, error) {
	query := `
		SELECT id, user_id, timezone, date_format, time_format, language,
			   email_notifications, push_notifications, sms_notifications,
			   created_at, updated_at
		FROM user_settings 
		WHERE user_id = ?`

	var us models.UserSettings
	var updatedAt sql.NullTime

	err := r.db.QueryRow(query, userID).Scan(
		&us.ID,
		&us.UserID,
		&us.Timezone,
		&us.DateFormat,
		&us.TimeFormat,
		&us.Language,
		&us.EmailNotifications,
		&us.PushNotifications,
		&us.SMSNotifications,
		&us.CreatedAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No settings found, return nil (not an error)
		}
		return nil, fmt.Errorf("failed to get user settings: %v", err)
	}

	if updatedAt.Valid {
		us.UpdatedAt = &updatedAt.Time
	}

	return &us, nil
}

// Create creates new user settings
func (r *UserSettingsRepository) Create(userID int, req *models.CreateUserSettingsRequest) (*models.UserSettings, error) {
	now := time.Now().Unix()
	
	query := `
		INSERT INTO user_settings (
			user_id, timezone, date_format, time_format, language,
			email_notifications, push_notifications, sms_notifications,
			created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		userID,
		req.Timezone,
		req.DateFormat,
		req.TimeFormat,
		req.Language,
		req.EmailNotifications,
		req.PushNotifications,
		req.SMSNotifications,
		now,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user settings: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get inserted ID: %v", err)
	}

	// Return the created settings
	return r.GetByID(int(id))
}

// Update updates existing user settings
func (r *UserSettingsRepository) Update(userID int, req *models.UpdateUserSettingsRequest) (*models.UserSettings, error) {
	// Build dynamic query based on provided fields
	setParts := []string{}
	args := []interface{}{}

	if req.Timezone != nil {
		setParts = append(setParts, "timezone = ?")
		args = append(args, *req.Timezone)
	}
	if req.DateFormat != nil {
		setParts = append(setParts, "date_format = ?")
		args = append(args, *req.DateFormat)
	}
	if req.TimeFormat != nil {
		setParts = append(setParts, "time_format = ?")
		args = append(args, *req.TimeFormat)
	}
	if req.Language != nil {
		setParts = append(setParts, "language = ?")
		args = append(args, *req.Language)
	}
	if req.EmailNotifications != nil {
		setParts = append(setParts, "email_notifications = ?")
		args = append(args, *req.EmailNotifications)
	}
	if req.PushNotifications != nil {
		setParts = append(setParts, "push_notifications = ?")
		args = append(args, *req.PushNotifications)
	}
	if req.SMSNotifications != nil {
		setParts = append(setParts, "sms_notifications = ?")
		args = append(args, *req.SMSNotifications)
	}

	if len(setParts) == 0 {
		// No fields to update, return current settings
		return r.GetByUserID(userID)
	}

	// Add updated_at and user_id to query
	setParts = append(setParts, "updated_at = ?")
	args = append(args, time.Now().Unix())
	args = append(args, userID)

	query := fmt.Sprintf("UPDATE user_settings SET %s WHERE user_id = ?", 
		fmt.Sprintf("%s", setParts[0]))
	
	for i := 1; i < len(setParts); i++ {
		query = fmt.Sprintf("%s, %s", query, setParts[i])
	}

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update user settings: %v", err)
	}

	// Return updated settings
	return r.GetByUserID(userID)
}

// GetByID retrieves user settings by ID
func (r *UserSettingsRepository) GetByID(id int) (*models.UserSettings, error) {
	query := `
		SELECT id, user_id, timezone, date_format, time_format, language,
			   email_notifications, push_notifications, sms_notifications,
			   created_at, updated_at
		FROM user_settings 
		WHERE id = ?`

	var us models.UserSettings
	var updatedAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&us.ID,
		&us.UserID,
		&us.Timezone,
		&us.DateFormat,
		&us.TimeFormat,
		&us.Language,
		&us.EmailNotifications,
		&us.PushNotifications,
		&us.SMSNotifications,
		&us.CreatedAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user settings not found")
		}
		return nil, fmt.Errorf("failed to get user settings: %v", err)
	}

	if updatedAt.Valid {
		us.UpdatedAt = &updatedAt.Time
	}

	return &us, nil
}

// CreateOrUpdate creates new settings or updates existing ones
func (r *UserSettingsRepository) CreateOrUpdate(userID int, req *models.CreateUserSettingsRequest) (*models.UserSettings, error) {
	// Check if settings exist
	existing, err := r.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		// Create new settings
		return r.Create(userID, req)
	}

	// Update existing settings
	updateReq := &models.UpdateUserSettingsRequest{
		Timezone:            &req.Timezone,
		DateFormat:          &req.DateFormat,
		TimeFormat:          &req.TimeFormat,
		Language:            &req.Language,
		EmailNotifications:  &req.EmailNotifications,
		PushNotifications:   &req.PushNotifications,
		SMSNotifications:    &req.SMSNotifications,
	}

	return r.Update(userID, updateReq)
}

// Delete removes user settings
func (r *UserSettingsRepository) Delete(userID int) error {
	query := "DELETE FROM user_settings WHERE user_id = ?"
	
	result, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user settings: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user settings not found")
	}

	return nil
}