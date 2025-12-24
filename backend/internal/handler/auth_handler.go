package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/go-playground/validator/v10"
	"citizen-appeals/internal/middleware"
	"citizen-appeals/internal/models"
	"citizen-appeals/internal/repository"
	"citizen-appeals/pkg/auth"
)

type AuthHandler struct {
	userRepo     *repository.UserRepository
	tokenService *auth.TokenService
	validator    *validator.Validate
}

func NewAuthHandler(userRepo *repository.UserRepository, tokenService *auth.TokenService) *AuthHandler {
	return &AuthHandler{
		userRepo:     userRepo,
		tokenService: tokenService,
		validator:    validator.New(),
	}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to process password", err)
		return
	}

	// Create user
	user := &models.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		Role:         models.RoleCitizen,
		IsActive:     true,
	}

	if err := h.userRepo.Create(r.Context(), user); err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyExists) {
			respondError(w, http.StatusConflict, "Email already exists", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to create user", err)
		return
	}

	// Generate token
	token, err := h.tokenService.GenerateToken(user)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token", err)
		return
	}

	response := models.LoginResponse{
		Token: token,
		User:  user,
	}

	respondJSON(w, http.StatusCreated, response)
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Get user by email
	user, err := h.userRepo.GetByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			respondError(w, http.StatusUnauthorized, "Invalid email or password", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get user", err)
		return
	}

	// Check if user is active
	if !user.IsActive {
		respondError(w, http.StatusUnauthorized, "Account is not active")
		return
	}

	// Check password
	if err := auth.CheckPassword(req.Password, user.PasswordHash); err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid email or password", err)
		return
	}

	// Generate token
	token, err := h.tokenService.GenerateToken(user)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token", err)
		return
	}

	response := models.LoginResponse{
		Token: token,
		User:  user,
	}

	respondJSON(w, http.StatusOK, response)
}

// Me returns the current authenticated user
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "User ID not found in context or token is invalid")
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found", err)
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// RefreshToken refreshes the JWT token
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Get token from Authorization header
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		respondError(w, http.StatusUnauthorized, "Missing token")
		return
	}

	// Remove "Bearer " prefix
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	// Refresh token
	newToken, err := h.tokenService.RefreshToken(tokenString)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid or expired token", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"token": newToken,
	})
}

// UpdateProfile updates the current user's profile (first name, last name, phone)
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Get existing user
	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "User not found", err)
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to get user", err)
		}
		return
	}

	// Update fields
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}

	if err := h.userRepo.Update(r.Context(), user); err != nil {
		if err == repository.ErrUserNotFound {
			respondError(w, http.StatusNotFound, "User not found", err)
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to update profile", err)
		}
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// ChangePassword changes the current user's password
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	var req models.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Get existing user
	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "User not found", err)
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to get user", err)
		}
		return
	}

	// Verify current password
	if err := auth.CheckPassword(req.CurrentPassword, user.PasswordHash); err != nil {
		respondError(w, http.StatusUnauthorized, "Current password is incorrect", err)
		return
	}

	// Hash new password
	newPasswordHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to process password", err)
		return
	}

	// Update password
	if err := h.userRepo.UpdatePassword(r.Context(), userID, newPasswordHash); err != nil {
		if err == repository.ErrUserNotFound {
			respondError(w, http.StatusNotFound, "User not found", err)
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to update password", err)
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Password changed successfully",
	})
}

// Helper functions
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    data,
	})
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
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: false,
		Error:   message,
	})
}
