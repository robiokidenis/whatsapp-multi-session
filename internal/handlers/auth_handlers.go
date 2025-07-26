package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
	"whatsapp-multi-session/internal/auth"
	"whatsapp-multi-session/internal/types"
)

// LoginRequest represents the login request body
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
	User    *UserInfo `json:"user,omitempty"`
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserInfo represents user information in responses
type UserInfo struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// LoginHandler handles user authentication
func LoginHandler(sessionManager *types.SessionManager, loginLimiter *types.LoginRateLimiter, cfg *types.Config, logger *types.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := auth.GetClientIP(r)
		
		// Check if IP is blocked
		if loginLimiter.IsBlocked(clientIP) {
			remaining := loginLimiter.GetRemainingTime(clientIP)
			if logger != nil {
				logger.Warn("Login attempt from blocked IP %s, %v remaining", clientIP, remaining)
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(LoginResponse{
				Success: false,
				Message: "Too many failed attempts. Please try again later.",
			})
			return
		}

		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			if logger != nil {
				logger.Warn("Invalid request body from IP %s: %v", clientIP, err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(LoginResponse{
				Success: false,
				Message: "Invalid request body",
			})
			return
		}

		// Validate input
		if req.Username == "" || req.Password == "" {
			if logger != nil {
				logger.Warn("Empty credentials from IP %s", clientIP)
			}
			loginLimiter.RecordAttempt(clientIP, false, logger)
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(LoginResponse{
				Success: false,
				Message: "Username and password are required",
			})
			return
		}

		// Authenticate user
		user, err := auth.AuthenticateUser(req.Username, req.Password, sessionManager)
		if err != nil {
			if logger != nil {
				logger.Warn("Authentication failed for user %s from IP %s: %v", req.Username, clientIP, err)
			}
			loginLimiter.RecordAttempt(clientIP, false, logger)
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(LoginResponse{
				Success: false,
				Message: "Invalid username or password",
			})
			return
		}

		// Generate JWT token
		token, err := auth.GenerateJWT(user.ID, user.Username, user.Role, cfg)
		if err != nil {
			if logger != nil {
				logger.Error("Failed to generate JWT for user %s from IP %s: %v", req.Username, clientIP, err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(LoginResponse{
				Success: false,
				Message: "Failed to generate token",
			})
			return
		}

		// Record successful login
		loginLimiter.RecordAttempt(clientIP, true, logger)
		if logger != nil {
			logger.Info("User %s logged in successfully from IP %s", req.Username, clientIP)
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(LoginResponse{
			Success: true,
			Message: "Login successful",
			Token:   token,
			User: &UserInfo{
				ID:       user.ID,
				Username: user.Username,
				Role:     user.Role,
			},
		})
	}
}

// RegisterHandler handles user registration
func RegisterHandler(sessionManager *types.SessionManager, logger *types.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(LoginResponse{
				Success: false,
				Message: "Invalid request body",
			})
			return
		}

		// Validate input
		if req.Username == "" || req.Password == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(LoginResponse{
				Success: false,
				Message: "Username and password are required",
			})
			return
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			if logger != nil {
				logger.Error("Failed to hash password: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(LoginResponse{
				Success: false,
				Message: "Failed to create user",
			})
			return
		}

		// Create user in database
		result, err := sessionManager.MetadataDB.Exec(`
			INSERT INTO users (username, password_hash, role, session_limit, is_active, created_at) 
			VALUES (?, ?, ?, ?, ?, ?)
		`, req.Username, string(hashedPassword), "user", 5, 1, time.Now().Unix())

		if err != nil {
			if logger != nil {
				logger.Error("Failed to create user %s: %v", req.Username, err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(LoginResponse{
				Success: false,
				Message: "Username already exists",
			})
			return
		}

		userID, _ := result.LastInsertId()
		if logger != nil {
			logger.Info("User %s registered successfully with ID %d", req.Username, userID)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(LoginResponse{
			Success: true,
			Message: "User registered successfully",
		})
	}
}