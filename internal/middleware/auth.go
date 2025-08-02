package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	
	"whatsapp-multi-session/internal/services"
)

// Claims represents JWT claims
type Claims struct {
	Username string `json:"username"`
	UserID   int    `json:"user_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// ContextKey type for context keys
type ContextKey string

const (
	// UserContextKey is the key for user claims in context
	UserContextKey ContextKey = "user"
)

// AuthMiddleware creates JWT authentication middleware
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			// Check Bearer prefix
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := tokenParts[1]

			// Parse and validate token
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Add claims to context (individual keys for compatibility)
			ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "username", claims.Username)
			ctx = context.WithValue(ctx, "role", claims.Role)
			ctx = context.WithValue(ctx, UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FlexibleAuthMiddleware creates middleware that supports both JWT and API key authentication
func FlexibleAuthMiddleware(jwtSecret string, userService *services.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			// Check Bearer prefix
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := tokenParts[1]

			// Try API key authentication first (if it looks like an API key)
			if strings.HasPrefix(tokenString, "wams_") {
				user, err := userService.AuthenticateAPIKey(tokenString)
				if err != nil {
					http.Error(w, "Invalid API key", http.StatusUnauthorized)
					return
				}

				// Create claims-like context for compatibility
				ctx := context.WithValue(r.Context(), "user_id", user.ID)
				ctx = context.WithValue(ctx, "username", user.Username)
				ctx = context.WithValue(ctx, "role", user.Role)
				
				// Create claims object for compatibility with existing code
				claims := &Claims{
					Username: user.Username,
					UserID:   user.ID,
					Role:     user.Role,
				}
				ctx = context.WithValue(ctx, UserContextKey, claims)
				
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Otherwise, try JWT authentication
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Add claims to context (individual keys for compatibility)
			ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "username", claims.Username)
			ctx = context.WithValue(ctx, "role", claims.Role)
			ctx = context.WithValue(ctx, UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole creates middleware that requires specific role
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserContextKey).(*Claims)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if claims.Role != role {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserClaims extracts user claims from context
func GetUserClaims(r *http.Request) (*Claims, bool) {
	claims, ok := r.Context().Value(UserContextKey).(*Claims)
	return claims, ok
}