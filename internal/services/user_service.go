package services

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"whatsapp-multi-session/internal/models"
	"whatsapp-multi-session/internal/repository"
	"whatsapp-multi-session/internal/utils"
	"whatsapp-multi-session/pkg/logger"
)

// UserService handles user-related business logic
type UserService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
	logger    *logger.Logger
}

// NewUserService creates a new user service
func NewUserService(
	userRepo *repository.UserRepository,
	jwtSecret string,
	log *logger.Logger,
) *UserService {
	return &UserService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		logger:    log,
	}
}

// Login authenticates a user and returns a JWT token
func (s *UserService) Login(req *models.LoginRequest) (*models.LoginResponse, error) {
	// Get user by username
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	if user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("account is disabled")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	token, err := s.generateJWT(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	s.logger.Info("User %s logged in successfully", user.Username)

	return &models.LoginResponse{
		Success: true,
		Token:   token,
		User:    user,
	}, nil
}

// Register registers a new user account
func (s *UserService) Register(req *models.RegisterRequest) (*models.RegisterResponse, error) {
	// Check if username already exists
	existing, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %v", err)
	}

	if existing != nil {
		return nil, fmt.Errorf("username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// Create user with default role and session limit
	user := &models.User{
		Username:     req.Username,
		Password:     string(hashedPassword),
		Role:         models.RoleUser, // Default role
		SessionLimit: 5,               // Default session limit
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	s.logger.Info("User %s registered successfully", user.Username)

	// Don't return password in response
	user.Password = ""

	return &models.RegisterResponse{
		Success: true,
		Message: "User registered successfully",
		User:    user,
	}, nil
}

// CreateUser creates a new user
func (s *UserService) CreateUser(req *models.CreateUserRequest) (*models.User, error) {
	// Check if username already exists
	existing, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %v", err)
	}

	if existing != nil {
		return nil, fmt.Errorf("username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// Create user
	user := &models.User{
		Username:     req.Username,
		Password:     string(hashedPassword),
		Role:         req.Role,
		SessionLimit: req.SessionLimit,
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	s.logger.Info("User %s created successfully", user.Username)
	return user, nil
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(id int, req *models.UpdateUserRequest) (*models.User, error) {
	// Get existing user
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update fields
	if req.Username != "" && req.Username != user.Username {
		// Check if new username already exists
		existing, err := s.userRepo.GetByUsername(req.Username)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing username: %v", err)
		}
		if existing != nil && existing.ID != user.ID {
			return nil, fmt.Errorf("username already exists")
		}
		user.Username = req.Username
	}

	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %v", err)
		}
		user.Password = string(hashedPassword)
	}

	if req.Role != "" {
		user.Role = req.Role
	}

	if req.SessionLimit > 0 {
		user.SessionLimit = req.SessionLimit
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	// Save changes
	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %v", err)
	}

	s.logger.Info("User %s updated successfully", user.Username)
	return user, nil
}

// ChangePassword changes a user's password
func (s *UserService) ChangePassword(userID int, req *models.ChangePasswordRequest) error {
	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}

	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return fmt.Errorf("invalid old password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %v", err)
	}

	// Update password
	user.Password = string(hashedPassword)
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update password: %v", err)
	}

	s.logger.Info("Password changed for user %s", user.Username)
	return nil
}

// GetUser returns a user by ID
func (s *UserService) GetUser(id int) (*models.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// GetAllUsers returns all users
func (s *UserService) GetAllUsers() ([]*models.User, error) {
	users, err := s.userRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %v", err)
	}

	return users, nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(id int) error {
	// Check if user exists
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}

	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Don't allow deleting the last admin
	if user.Role == models.RoleAdmin {
		adminCount, err := s.userRepo.CountByRole(models.RoleAdmin)
		if err != nil {
			return fmt.Errorf("failed to count admins: %v", err)
		}

		if adminCount <= 1 {
			return fmt.Errorf("cannot delete the last admin user")
		}
	}

	// Delete user
	if err := s.userRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}

	s.logger.Info("User %s deleted successfully", user.Username)
	return nil
}

// EnsureDefaultAdmin creates a default admin user if none exists
func (s *UserService) EnsureDefaultAdmin(username, password string) error {
	// Check if any admin exists
	adminCount, err := s.userRepo.CountByRole(models.RoleAdmin)
	if err != nil {
		return fmt.Errorf("failed to count admins: %v", err)
	}

	if adminCount > 0 {
		s.logger.Debug("Admin user already exists, skipping creation")
		return nil // Admin already exists
	}

	// Check if the username already exists
	existingUser, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return fmt.Errorf("failed to check existing user: %v", err)
	}

	if existingUser != nil {
		// User exists but is not an admin - upgrade to admin
		if existingUser.Role != models.RoleAdmin {
			existingUser.Role = models.RoleAdmin
			if err := s.userRepo.Update(existingUser); err != nil {
				return fmt.Errorf("failed to upgrade user to admin: %v", err)
			}
			s.logger.Info("Upgraded existing user %s to admin role", username)
		}
		return nil
	}

	// Create default admin
	req := &models.CreateUserRequest{
		Username:     username,
		Password:     password,
		Role:         models.RoleAdmin,
		SessionLimit: 10,
	}

	_, err = s.CreateUser(req)
	if err != nil {
		return fmt.Errorf("failed to create default admin: %v", err)
	}

	s.logger.Info("Default admin user created: %s", username)
	return nil
}

// Claims represents JWT claims (should match middleware)
type Claims struct {
	Username string `json:"username"`
	UserID   int    `json:"user_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// generateJWT generates a JWT token for a user
func (s *UserService) generateJWT(user *models.User) (string, error) {
	claims := &Claims{
		Username: user.Username,
		UserID:   user.ID,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", user.ID),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer:    "whatsapp-multi-session",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// GenerateAPIKey generates a new API key for a user
func (s *UserService) GenerateAPIKey(userID int) (*models.APIKeyResponse, error) {
	// Check if user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Generate new API key
	apiKey, err := s.generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %v", err)
	}

	// Update user with new API key
	if err := s.userRepo.UpdateAPIKey(userID, apiKey); err != nil {
		return nil, fmt.Errorf("failed to save API key: %v", err)
	}

	s.logger.Info("Generated new API key for user %s (ID: %d)", user.Username, userID)

	return &models.APIKeyResponse{
		Success: true,
		Message: "API key generated successfully",
		APIKey:  apiKey,
	}, nil
}

// RevokeAPIKey revokes a user's API key
func (s *UserService) RevokeAPIKey(userID int) error {
	// Check if user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Remove API key
	if err := s.userRepo.RemoveAPIKey(userID); err != nil {
		return fmt.Errorf("failed to revoke API key: %v", err)
	}

	s.logger.Info("Revoked API key for user %s (ID: %d)", user.Username, userID)
	return nil
}

// GetAPIKeyInfo returns information about a user's API key (without the key itself)
func (s *UserService) GetAPIKeyInfo(userID int) (*models.APIKeyInfo, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return &models.APIKeyInfo{
		HasKey:    user.APIKey != "",
		CreatedAt: user.CreatedAt,
		LastUsed:  user.UpdatedAt,
	}, nil
}

// AuthenticateAPIKey authenticates a user by API key
func (s *UserService) AuthenticateAPIKey(apiKey string) (*models.User, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	user, err := s.userRepo.GetByAPIKey(apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate API key: %v", err)
	}

	if user == nil {
		return nil, fmt.Errorf("invalid API key")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("account is disabled")
	}

	return user, nil
}

// generateAPIKey generates a secure API key
func (s *UserService) generateAPIKey() (string, error) {
	// Import the utils package
	return utils.GenerateAPIKey()
}