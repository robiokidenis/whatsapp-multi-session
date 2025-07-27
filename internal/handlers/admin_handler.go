package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"whatsapp-multi-session/internal/models"
	"whatsapp-multi-session/internal/services"
	"whatsapp-multi-session/pkg/logger"
)

// AdminHandler handles admin-related endpoints
type AdminHandler struct {
	userService *services.UserService
	logger      *logger.Logger
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	userService *services.UserService,
	log *logger.Logger,
) *AdminHandler {
	return &AdminHandler{
		userService: userService,
		logger:      log,
	}
}

// GetUsers handles getting all users
func (h *AdminHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		h.logger.Error("Failed to get users: %v", err)
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	// Remove passwords from response
	for _, user := range users {
		user.Password = ""
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    users,
	})
}

// GetUser handles getting a specific user
func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetUser(userID)
	if err != nil {
		h.logger.Error("Failed to get user %d: %v", userID, err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Remove password from response
	user.Password = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    user,
	})
}

// CreateUser handles creating a new user
func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	if len(req.Username) < 3 {
		http.Error(w, "Username must be at least 3 characters", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 6 {
		http.Error(w, "Password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.Role == "" {
		req.Role = models.RoleUser
	}
	if req.SessionLimit == 0 {
		req.SessionLimit = 5
	}

	// Validate role
	if req.Role != models.RoleAdmin && req.Role != models.RoleUser {
		http.Error(w, "Role must be 'admin' or 'user'", http.StatusBadRequest)
		return
	}

	user, err := h.userService.CreateUser(&req)
	if err != nil {
		h.logger.Error("Failed to create user: %v", err)
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	// Remove password from response
	user.Password = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User created successfully",
		"data":    user,
	})
}

// UpdateUser handles updating a user
func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Username != "" && len(req.Username) < 3 {
		http.Error(w, "Username must be at least 3 characters", http.StatusBadRequest)
		return
	}

	if req.Password != "" && len(req.Password) < 6 {
		http.Error(w, "Password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	if req.Role != "" && req.Role != models.RoleAdmin && req.Role != models.RoleUser {
		http.Error(w, "Role must be 'admin' or 'user'", http.StatusBadRequest)
		return
	}

	user, err := h.userService.UpdateUser(userID, &req)
	if err != nil {
		h.logger.Error("Failed to update user %d: %v", userID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Remove password from response
	user.Password = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User updated successfully",
		"data":    user,
	})
}

// DeleteUser handles deleting a user
func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := h.userService.DeleteUser(userID); err != nil {
		h.logger.Error("Failed to delete user %d: %v", userID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User deleted successfully",
	})
}