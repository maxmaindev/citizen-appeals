package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"citizen-appeals/internal/middleware"
	"citizen-appeals/internal/models"
	"citizen-appeals/internal/repository"
)

type UserHandler struct {
	userRepo  *repository.UserRepository
	validator *validator.Validate
}

func NewUserHandler(userRepo *repository.UserRepository, validator *validator.Validate) *UserHandler {
	return &UserHandler{
		userRepo:  userRepo,
		validator: validator,
	}
}

// List retrieves all users (admin only) or executors (dispatcher/admin)
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	// Check if requesting executors only
	roleFilter := r.URL.Query().Get("role")
	
	if roleFilter == "executor" {
		// Get all executors (for dispatcher/admin to assign appeals)
		serviceIDStr := r.URL.Query().Get("service_id")
		var executors []*models.User
		var err error
		
		if serviceIDStr != "" {
			serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
			if err == nil {
				executors, err = h.userRepo.GetExecutorsByService(r.Context(), serviceID)
			}
		} else {
			// Get all executors
			executors, err = h.userRepo.GetExecutorsByService(r.Context(), 0)
		}
		
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to get executors", err)
			return
		}
		
		respondJSON(w, http.StatusOK, executors)
		return
	}

	page := 1
	limit := 20

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	users, total, err := h.userRepo.List(r.Context(), page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get users", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"items":       users,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	})
}

// GetByID retrieves a user by ID (admin only)
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrUserNotFound {
			respondError(w, http.StatusNotFound, "User not found", err)
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to get user", err)
		}
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// Update updates a user (admin only)
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Get existing user
	user, err := h.userRepo.GetByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrUserNotFound {
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
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := h.userRepo.Update(r.Context(), user); err != nil {
		if err == repository.ErrUserNotFound {
			respondError(w, http.StatusNotFound, "User not found", err)
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to update user", err)
		}
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// Delete soft deletes a user (admin only)
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	// Prevent deleting yourself
	userID, _ := middleware.GetUserID(r.Context())
	if id == userID {
		respondError(w, http.StatusBadRequest, "Cannot delete your own account")
		return
	}

	if err := h.userRepo.Delete(r.Context(), id); err != nil {
		if err == repository.ErrUserNotFound {
			respondError(w, http.StatusNotFound, "User not found", err)
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to delete user", err)
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

