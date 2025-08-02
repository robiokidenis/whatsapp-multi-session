package repository

import (
	"database/sql"
	"fmt"
	"time"

	"whatsapp-multi-session/internal/models"
)

// UserRepository handles user data persistence
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (username, password_hash, api_key, role, session_limit, is_active, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := r.db.Exec(query, 
		user.Username, 
		user.Password,
		user.APIKey,
		user.Role, 
		user.SessionLimit, 
		user.IsActive,
		user.CreatedAt.Unix(),
	)
	
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %v", err)
	}
	
	user.ID = int(id)
	return nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, password_hash, api_key, role, session_limit, is_active, created_at, updated_at
		FROM users
		WHERE username = ?
	`
	
	var createdAtUnix int64
	var updatedAtUnix sql.NullInt64
	var apiKey sql.NullString
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&apiKey,
		&user.Role,
		&user.SessionLimit,
		&user.IsActive,
		&createdAtUnix,
		&updatedAtUnix,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	
	user.CreatedAt = time.Unix(createdAtUnix, 0)
	if updatedAtUnix.Valid {
		updatedTime := time.Unix(updatedAtUnix.Int64, 0)
		user.UpdatedAt = &updatedTime
	}
	if apiKey.Valid {
		user.APIKey = apiKey.String
	}
	
	return user, nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id int) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, password_hash, api_key, role, session_limit, is_active, created_at, updated_at
		FROM users
		WHERE id = ?
	`
	
	var createdAtUnix int64
	var updatedAtUnix sql.NullInt64
	var apiKey sql.NullString
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&apiKey,
		&user.Role,
		&user.SessionLimit,
		&user.IsActive,
		&createdAtUnix,
		&updatedAtUnix,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	
	user.CreatedAt = time.Unix(createdAtUnix, 0)
	if updatedAtUnix.Valid {
		updatedTime := time.Unix(updatedAtUnix.Int64, 0)
		user.UpdatedAt = &updatedTime
	}
	if apiKey.Valid {
		user.APIKey = apiKey.String
	}
	
	return user, nil
}

// GetByAPIKey retrieves a user by API key
func (r *UserRepository) GetByAPIKey(apiKey string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, password_hash, api_key, role, session_limit, is_active, created_at, updated_at
		FROM users
		WHERE api_key = ?
	`
	
	var createdAtUnix int64
	var updatedAtUnix sql.NullInt64
	var userAPIKey sql.NullString
	err := r.db.QueryRow(query, apiKey).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&userAPIKey,
		&user.Role,
		&user.SessionLimit,
		&user.IsActive,
		&createdAtUnix,
		&updatedAtUnix,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to get user by API key: %v", err)
	}
	
	user.CreatedAt = time.Unix(createdAtUnix, 0)
	if updatedAtUnix.Valid {
		updatedTime := time.Unix(updatedAtUnix.Int64, 0)
		user.UpdatedAt = &updatedTime
	}
	if userAPIKey.Valid {
		user.APIKey = userAPIKey.String
	}
	
	return user, nil
}

// Update updates a user
func (r *UserRepository) Update(user *models.User) error {
	now := time.Now()
	user.UpdatedAt = &now
	
	query := `
		UPDATE users
		SET username = ?, password_hash = ?, role = ?, session_limit = ?, is_active = ?, updated_at = ?
		WHERE id = ?
	`
	
	_, err := r.db.Exec(query,
		user.Username,
		user.Password,
		user.Role,
		user.SessionLimit,
		user.IsActive,
		user.UpdatedAt.Unix(),
		user.ID,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}
	
	return nil
}

// UpdateAPIKey updates a user's API key
func (r *UserRepository) UpdateAPIKey(userID int, apiKey string) error {
	query := `
		UPDATE users
		SET api_key = ?, updated_at = ?
		WHERE id = ?
	`
	
	_, err := r.db.Exec(query, apiKey, time.Now().Unix(), userID)
	if err != nil {
		return fmt.Errorf("failed to update API key: %v", err)
	}
	
	return nil
}

// RemoveAPIKey removes a user's API key
func (r *UserRepository) RemoveAPIKey(userID int) error {
	query := `
		UPDATE users
		SET api_key = NULL, updated_at = ?
		WHERE id = ?
	`
	
	_, err := r.db.Exec(query, time.Now().Unix(), userID)
	if err != nil {
		return fmt.Errorf("failed to remove API key: %v", err)
	}
	
	return nil
}

// Delete deletes a user
func (r *UserRepository) Delete(id int) error {
	query := `DELETE FROM users WHERE id = ?`
	
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}
	
	return nil
}

// GetAll retrieves all users
func (r *UserRepository) GetAll() ([]*models.User, error) {
	query := `
		SELECT id, username, password_hash, role, session_limit, is_active, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %v", err)
	}
	defer rows.Close()
	
	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		var createdAtUnix int64
		var updatedAtUnix sql.NullInt64
		
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Password,
			&user.Role,
			&user.SessionLimit,
			&user.IsActive,
			&createdAtUnix,
			&updatedAtUnix,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %v", err)
		}
		
		user.CreatedAt = time.Unix(createdAtUnix, 0)
		if updatedAtUnix.Valid {
			updatedTime := time.Unix(updatedAtUnix.Int64, 0)
			user.UpdatedAt = &updatedTime
		}
		
		users = append(users, user)
	}
	
	return users, nil
}

// CountByRole counts users by role
func (r *UserRepository) CountByRole(role string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE role = ?`
	
	err := r.db.QueryRow(query, role).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %v", err)
	}
	
	return count, nil
}