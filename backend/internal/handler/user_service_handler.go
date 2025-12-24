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

type UserServiceHandler struct {
	userServiceRepo *repository.UserServiceRepository
	validator       *validator.Validate
}

func NewUserServiceHandler(userServiceRepo *repository.UserServiceRepository) *UserServiceHandler {
	return &UserServiceHandler{
		userServiceRepo: userServiceRepo,
		validator:       validator.New(),
	}
}

// GetAll retrieves all services with their assigned users
func (h *UserServiceHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	result, err := h.userServiceRepo.GetAllWithDetails(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get service users", err)
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// GetByServiceID retrieves all users assigned to a specific service
func (h *UserServiceHandler) GetByServiceID(w http.ResponseWriter, r *http.Request) {
	serviceIDStr := chi.URLParam(r, "service_id")
	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid service ID", err)
		return
	}

	users, err := h.userServiceRepo.GetByServiceID(r.Context(), serviceID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get service users", err)
		return
	}

	respondJSON(w, http.StatusOK, users)
}

// GetMyServices retrieves all services assigned to the current user (for executors)
func (h *UserServiceHandler) GetMyServices(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	services, err := h.userServiceRepo.GetByUserID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get user services", err)
		return
	}

	respondJSON(w, http.StatusOK, services)
}

// AssignUsers assigns users (executors) to a service
func (h *UserServiceHandler) AssignUsers(w http.ResponseWriter, r *http.Request) {
	var req models.UserServiceAssignment
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	if err := h.userServiceRepo.AssignUsers(r.Context(), req.ServiceID, req.UserIDs); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to assign users to service", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Users assigned to service successfully",
	})
}

// Delete removes a user-service relationship
func (h *UserServiceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	serviceIDStr := chi.URLParam(r, "service_id")
	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid service ID", err)
		return
	}

	userIDStr := chi.URLParam(r, "user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	if err := h.userServiceRepo.Delete(r.Context(), userID, serviceID); err != nil {
		if err == repository.ErrUserServiceNotFound {
			respondError(w, http.StatusNotFound, "User-service relationship not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete user-service relationship", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "User removed from service successfully",
	})
}
