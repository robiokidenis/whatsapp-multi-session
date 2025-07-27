package models

import "time"

// User represents a user account
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Password     string    `json:"-"` // Don't include in JSON responses
	Role         string    `json:"role"`
	SessionLimit int       `json:"session_limit"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

// UserRole constants
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the response after successful login
type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterResponse represents the response after successful registration
type RegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	User    *User  `json:"user"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// CreateUserRequest represents user creation request
type CreateUserRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	Role         string `json:"role"`
	SessionLimit int    `json:"session_limit"`
}

// UpdateUserRequest represents user update request
type UpdateUserRequest struct {
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	Role         string `json:"role,omitempty"`
	SessionLimit int    `json:"session_limit,omitempty"`
	IsActive     *bool  `json:"is_active,omitempty"`
}