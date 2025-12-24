package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"citizen-appeals/internal/models"
	"citizen-appeals/pkg/auth"
)

func TestAuthMiddleware_ValidToken(t *testing.T) {
	// Arrange
	jwtSecret := "test-secret-key-for-testing-purposes-only"
	tokenService := auth.NewTokenService(jwtSecret, 24*time.Hour)
	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  models.RoleCitizen,
	}
	token, err := tokenService.GenerateToken(user)
	assert.NoError(t, err)

	handler := AuthMiddleware(tokenService)
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		assert.True(t, ok)
		assert.Equal(t, int64(1), userID)

		role, ok := GetUserRole(r.Context())
		assert.True(t, ok)
		assert.Equal(t, models.RoleCitizen, role)

		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	// Act
	handler(nextHandler).ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	// Arrange
	jwtSecret := "test-secret-key-for-testing-purposes-only"
	tokenService := auth.NewTokenService(jwtSecret, 24*time.Hour)
	handler := AuthMiddleware(tokenService)
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	// Act
	handler(nextHandler).ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	// Arrange
	jwtSecret := "test-secret-key-for-testing-purposes-only"
	tokenService := auth.NewTokenService(jwtSecret, 24*time.Hour)
	handler := AuthMiddleware(tokenService)
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Act
	handler(nextHandler).ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	// Arrange
	jwtSecret := "test-secret-key-for-testing-purposes-only"
	tokenService := auth.NewTokenService(jwtSecret, 24*time.Hour)
	handler := AuthMiddleware(tokenService)
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat token")
	w := httptest.NewRecorder()

	// Act
	handler(nextHandler).ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireRole_AllowedRole(t *testing.T) {
	// Arrange
	jwtSecret := "test-secret-key-for-testing-purposes-only"
	tokenService := auth.NewTokenService(jwtSecret, 24*time.Hour)
	user := &models.User{
		ID:    1,
		Email: "dispatcher@example.com",
		Role:  models.RoleDispatcher,
	}
	token, err := tokenService.GenerateToken(user)
	assert.NoError(t, err)

	handler := RequireRole(models.RoleDispatcher, models.RoleAdmin)
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	authMiddleware := AuthMiddleware(tokenService)

	// Act
	authMiddleware(handler(nextHandler)).ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_ForbiddenRole(t *testing.T) {
	// Arrange
	jwtSecret := "test-secret-key-for-testing-purposes-only"
	tokenService := auth.NewTokenService(jwtSecret, 24*time.Hour)
	user := &models.User{
		ID:    1,
		Email: "citizen@example.com",
		Role:  models.RoleCitizen,
	}
	token, err := tokenService.GenerateToken(user)
	assert.NoError(t, err)

	handler := RequireRole(models.RoleDispatcher, models.RoleAdmin)
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	authMiddleware := AuthMiddleware(tokenService)

	// Act
	authMiddleware(handler(nextHandler)).ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
}
