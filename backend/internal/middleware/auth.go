package middleware

import (
	"context"
	"log"
	"net/http"
	"runtime/debug"
	"strings"

	"citizen-appeals/internal/models"
	"citizen-appeals/pkg/auth"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserEmailKey contextKey = "user_email"
	UserRoleKey  contextKey = "user_role"
)

// AuthMiddleware validates JWT token and adds user info to context
func AuthMiddleware(tokenService *auth.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondError(w, http.StatusUnauthorized, "Missing authorization header")
				return
			}

			// Check Bearer prefix
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondError(w, http.StatusUnauthorized, "Invalid authorization header format")
				return
			}

			tokenString := parts[1]

			// Validate token
			claims, err := tokenService.ValidateToken(tokenString)
			if err != nil {
				respondError(w, http.StatusUnauthorized, "Invalid or expired token", err)
				return
			}

			// Add user info to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

			// Call next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole middleware checks if user has required role
func RequireRole(roles ...models.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(UserRoleKey).(models.UserRole)
			if !ok {
				respondError(w, http.StatusUnauthorized, "User role not found")
				return
			}

			// Check if user has required role
			hasRole := false
			for _, role := range roles {
				if userRole == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				respondError(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}

// GetUserRole extracts user role from context
func GetUserRole(ctx context.Context) (models.UserRole, bool) {
	role, ok := ctx.Value(UserRoleKey).(models.UserRole)
	return role, ok
}

func respondError(w http.ResponseWriter, status int, message string, details ...error) {
	var err error
	if len(details) > 0 {
		err = details[0]
	}

	if err != nil {
		log.Printf("[ERROR] status=%d message=%s error=%v", status, message, err)
		if status >= http.StatusInternalServerError {
			log.Printf("[STACK]\n%s", debug.Stack())
		}
	} else {
		log.Printf("[ERROR] status=%d message=%s", status, message)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"success": false, "error": "` + message + `"}`))
}
