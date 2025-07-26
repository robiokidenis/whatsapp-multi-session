package auth

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"whatsapp-multi-session/internal/types"
)

// GenerateJWT generates a JWT token for the user
func GenerateJWT(userID int, username string, role string, config *types.Config) (string, error) {
	claims := &types.Claims{
		Username: username,
		UserID:   userID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.JWTSecret))
}

// ValidateJWT validates a JWT token and returns claims
func ValidateJWT(tokenString string, config *types.Config) (*types.Claims, error) {
	claims := &types.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// Middleware validates JWT tokens for protected routes
func Middleware(config *types.Config) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
			claims, err := ValidateJWT(tokenString, config)
			if err != nil {
				http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
				return
			}

			// Check admin privileges for admin endpoints
			if strings.HasPrefix(r.URL.Path, "/api/admin/") {
				// Role-based access control: only admin role can access admin endpoints
				if claims.Role != "admin" {
					http.Error(w, "Admin privileges required", http.StatusForbidden)
					return
				}
			}

			// Add user info to request context
			ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "username", claims.Username)
			ctx = context.WithValue(ctx, "role", claims.Role)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		}
	}
}

// AuthenticateUser validates username and password
func AuthenticateUser(username, password string, sessionManager *types.SessionManager) (*types.User, error) {
	var user types.User
	var hashedPassword string
	var createdAtInterface interface{}

	err := sessionManager.MetadataDB.QueryRow(`
		SELECT id, username, password_hash, role, session_limit, is_active, created_at 
		FROM users 
		WHERE username = ? AND is_active = 1
	`, username).Scan(&user.ID, &user.Username, &hashedPassword, &user.Role, &user.SessionLimit, &user.IsActive, &createdAtInterface)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	// Handle different timestamp formats
	switch v := createdAtInterface.(type) {
	case int64:
		user.CreatedAt = v
	case time.Time:
		user.CreatedAt = v.Unix()
	case []uint8: // MySQL timestamp as bytes
		if t, err := time.Parse("2006-01-02 15:04:05", string(v)); err == nil {
			user.CreatedAt = t.Unix()
		} else {
			user.CreatedAt = time.Now().Unix()
		}
	default:
		user.CreatedAt = time.Now().Unix()
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return &user, nil
}

// GetClientIP extracts the real client IP from request headers
func GetClientIP(r *http.Request) string {
	// Check for forwarded IP in headers (for reverse proxy setups)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP if multiple are listed
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check for real IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to remote address
	ip := r.RemoteAddr
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}
	return ip
}