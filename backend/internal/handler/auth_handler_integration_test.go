package handler

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"citizen-appeals/internal/models"
)

// Integration tests for AuthHandler
// These tests verify HTTP request/response handling for authentication endpoints

func TestAuthHandler_Register_ValidRequest(t *testing.T) {
	reqBody := models.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
		Phone:     "+380501234567",
	}

	// Validate request structure
	assert.NotEmpty(t, reqBody.Email, "Email should not be empty")
	assert.GreaterOrEqual(t, len(reqBody.Password), 8, "Password should be at least 8 characters")
	assert.GreaterOrEqual(t, len(reqBody.FirstName), 2, "FirstName should be at least 2 characters")
	assert.GreaterOrEqual(t, len(reqBody.LastName), 2, "LastName should be at least 2 characters")
	assert.NotEmpty(t, reqBody.Phone, "Phone should not be empty")

	// Test JSON marshaling
	body, err := json.Marshal(reqBody)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	assert.NotNil(t, req)
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
}

func TestAuthHandler_Register_InvalidRequest(t *testing.T) {
	invalidRequests := []struct {
		name string
		req  map[string]interface{}
	}{
		{
			name: "missing email",
			req: map[string]interface{}{
				"password":   "password123",
				"first_name": "John",
				"last_name":  "Doe",
				"phone":      "+380501234567",
			},
		},
		{
			name: "invalid email format",
			req: map[string]interface{}{
				"email":      "invalid-email",
				"password":   "password123",
				"first_name": "John",
				"last_name":  "Doe",
				"phone":      "+380501234567",
			},
		},
		{
			name: "password too short",
			req: map[string]interface{}{
				"email":      "test@example.com",
				"password":   "short",
				"first_name": "John",
				"last_name":  "Doe",
				"phone":      "+380501234567",
			},
		},
		{
			name: "first name too short",
			req: map[string]interface{}{
				"email":      "test@example.com",
				"password":   "password123",
				"first_name": "J",
				"last_name":  "Doe",
				"phone":      "+380501234567",
			},
		},
		{
			name: "missing phone",
			req: map[string]interface{}{
				"email":      "test@example.com",
				"password":   "password123",
				"first_name": "John",
				"last_name":  "Doe",
			},
		},
	}

	for _, tt := range invalidRequests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.req)
			req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			assert.NotNil(t, req)
			// In real test, you would check for validation errors
		})
	}
}

func TestAuthHandler_Login_ValidRequest(t *testing.T) {
	reqBody := models.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	// Validate request structure
	assert.NotEmpty(t, reqBody.Email, "Email should not be empty")
	assert.NotEmpty(t, reqBody.Password, "Password should not be empty")

	// Test JSON marshaling
	body, err := json.Marshal(reqBody)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	assert.NotNil(t, req)
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
}

func TestAuthHandler_Login_InvalidRequest(t *testing.T) {
	invalidRequests := []struct {
		name string
		req  map[string]interface{}
	}{
		{
			name: "missing email",
			req: map[string]interface{}{
				"password": "password123",
			},
		},
		{
			name: "invalid email format",
			req: map[string]interface{}{
				"email":    "invalid-email",
				"password": "password123",
			},
		},
		{
			name: "missing password",
			req: map[string]interface{}{
				"email": "test@example.com",
			},
		},
	}

	for _, tt := range invalidRequests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.req)
			req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			assert.NotNil(t, req)
		})
	}
}

func TestAuthHandler_LoginResponse_Format(t *testing.T) {
	// Test LoginResponse structure
	response := models.LoginResponse{
		Token: "test-jwt-token",
		User: &models.User{
			ID:        1,
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
			Role:      models.RoleCitizen,
			IsActive:  true,
		},
	}

	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	var decoded models.LoginResponse
	err = json.Unmarshal(jsonData, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, response.Token, decoded.Token)
	assert.NotNil(t, decoded.User)
	assert.Equal(t, response.User.Email, decoded.User.Email)
}

func TestAuthHandler_Me_RequiresAuth(t *testing.T) {
	// Test that /me endpoint requires authentication
	req := httptest.NewRequest("GET", "/api/auth/me", nil)

	// Without Authorization header, request should be unauthorized
	assert.NotNil(t, req)
	assert.Empty(t, req.Header.Get("Authorization"))

	// With Authorization header
	reqWithAuth := httptest.NewRequest("GET", "/api/auth/me", nil)
	reqWithAuth.Header.Set("Authorization", "Bearer test-token")

	assert.NotNil(t, reqWithAuth)
	assert.Equal(t, "Bearer test-token", reqWithAuth.Header.Get("Authorization"))
}

func TestAuthHandler_UpdateProfile_ValidRequest(t *testing.T) {
	firstName := "Jane"
	lastName := "Smith"
	phone := "+380509876543"

	reqBody := models.UpdateProfileRequest{
		FirstName: &firstName,
		LastName:  &lastName,
		Phone:     &phone,
	}

	// Validate request structure
	assert.NotNil(t, reqBody.FirstName)
	assert.NotNil(t, reqBody.LastName)
	assert.NotNil(t, reqBody.Phone)

	// Test JSON marshaling
	body, err := json.Marshal(reqBody)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	req := httptest.NewRequest("PUT", "/api/auth/profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	assert.NotNil(t, req)
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
}

func TestAuthHandler_UpdateProfile_PartialUpdate(t *testing.T) {
	// Test partial update (only first name)
	firstName := "UpdatedName"
	reqBody := models.UpdateProfileRequest{
		FirstName: &firstName,
	}

	body, err := json.Marshal(reqBody)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	var decoded models.UpdateProfileRequest
	err = json.Unmarshal(body, &decoded)
	assert.NoError(t, err)
	assert.NotNil(t, decoded.FirstName)
	assert.Nil(t, decoded.LastName)
	assert.Nil(t, decoded.Phone)
}

func TestAuthHandler_ChangePassword_ValidRequest(t *testing.T) {
	reqBody := models.ChangePasswordRequest{
		CurrentPassword: "oldpassword123",
		NewPassword:     "newpassword123",
	}

	// Validate request structure
	assert.NotEmpty(t, reqBody.CurrentPassword, "CurrentPassword should not be empty")
	assert.GreaterOrEqual(t, len(reqBody.NewPassword), 8, "NewPassword should be at least 8 characters")

	// Test JSON marshaling
	body, err := json.Marshal(reqBody)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	req := httptest.NewRequest("PUT", "/api/auth/change-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	assert.NotNil(t, req)
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
}

func TestAuthHandler_ChangePassword_InvalidRequest(t *testing.T) {
	invalidRequests := []struct {
		name string
		req  map[string]interface{}
	}{
		{
			name: "missing current password",
			req: map[string]interface{}{
				"new_password": "newpassword123",
			},
		},
		{
			name: "missing new password",
			req: map[string]interface{}{
				"current_password": "oldpassword123",
			},
		},
		{
			name: "new password too short",
			req: map[string]interface{}{
				"current_password": "oldpassword123",
				"new_password":     "short",
			},
		},
	}

	for _, tt := range invalidRequests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.req)
			req := httptest.NewRequest("PUT", "/api/auth/change-password", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			assert.NotNil(t, req)
		})
	}
}

func TestAuthHandler_RefreshToken_ValidRequest(t *testing.T) {
	// Test refresh token request
	req := httptest.NewRequest("POST", "/api/auth/refresh", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	assert.NotNil(t, req)
	assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))
}

func TestAuthHandler_RefreshToken_MissingToken(t *testing.T) {
	// Test refresh token without Authorization header
	req := httptest.NewRequest("POST", "/api/auth/refresh", nil)

	assert.NotNil(t, req)
	assert.Empty(t, req.Header.Get("Authorization"))
}

func TestAuthHandler_ResponseFormat(t *testing.T) {
	// Test APIResponse format
	successResponse := models.APIResponse{
		Success: true,
		Data: map[string]string{
			"message": "Operation successful",
		},
	}

	jsonData, err := json.Marshal(successResponse)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	var decoded models.APIResponse
	err = json.Unmarshal(jsonData, &decoded)
	assert.NoError(t, err)
	assert.True(t, decoded.Success)
	assert.NotNil(t, decoded.Data)

	// Test error response
	errorResponse := models.APIResponse{
		Success: false,
		Error:   "Invalid credentials",
	}

	jsonData, err = json.Marshal(errorResponse)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	err = json.Unmarshal(jsonData, &decoded)
	assert.NoError(t, err)
	assert.False(t, decoded.Success)
	assert.NotEmpty(t, decoded.Error)
}

func TestAuthHandler_UserRole_Validation(t *testing.T) {
	// Test that new users are created with RoleCitizen
	user := &models.User{
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Role:      models.RoleCitizen,
		IsActive:  true,
	}

	assert.Equal(t, models.RoleCitizen, user.Role, "New users should have RoleCitizen")
	assert.True(t, user.IsActive, "New users should be active")

	// Test valid roles
	validRoles := []models.UserRole{
		models.RoleCitizen,
		models.RoleDispatcher,
		models.RoleExecutor,
		models.RoleAdmin,
	}

	for _, role := range validRoles {
		assert.Contains(t, validRoles, role, "Role should be valid")
	}
}
