package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"citizen-appeals/internal/models"
	"citizen-appeals/internal/repository"
)

type CategoryServiceHandler struct {
	categoryServiceRepo *repository.CategoryServiceRepository
	validator           *validator.Validate
}

func NewCategoryServiceHandler(categoryServiceRepo *repository.CategoryServiceRepository) *CategoryServiceHandler {
	return &CategoryServiceHandler{
		categoryServiceRepo: categoryServiceRepo,
		validator:           validator.New(),
	}
}

// GetAll retrieves all categories with their assigned services
func (h *CategoryServiceHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	data, err := h.categoryServiceRepo.GetAllWithDetails(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get category services", err)
		return
	}

	respondJSON(w, http.StatusOK, data)
}

// GetByCategoryID retrieves all services assigned to a category
func (h *CategoryServiceHandler) GetByCategoryID(w http.ResponseWriter, r *http.Request) {
	categoryIDStr := chi.URLParam(r, "category_id")
	categoryID, err := strconv.ParseInt(categoryIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid category ID", err)
		return
	}

	services, err := h.categoryServiceRepo.GetByCategoryID(r.Context(), categoryID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get category services", err)
		return
	}

	respondJSON(w, http.StatusOK, services)
}

// AssignServices assigns services to a category (replaces existing assignments)
func (h *CategoryServiceHandler) AssignServices(w http.ResponseWriter, r *http.Request) {
	var req models.CategoryServiceAssignment
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	if err := h.categoryServiceRepo.AssignServices(r.Context(), req.CategoryID, req.ServiceIDs); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to assign services", err)
		return
	}

	// Return updated assignments
	services, err := h.categoryServiceRepo.GetByCategoryID(r.Context(), req.CategoryID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get updated assignments", err)
		return
	}

	respondJSON(w, http.StatusOK, services)
}

// Delete removes a service assignment from a category
func (h *CategoryServiceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	categoryIDStr := chi.URLParam(r, "category_id")
	categoryID, err := strconv.ParseInt(categoryIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid category ID", err)
		return
	}

	serviceIDStr := chi.URLParam(r, "service_id")
	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid service ID", err)
		return
	}

	if err := h.categoryServiceRepo.Delete(r.Context(), categoryID, serviceID); err != nil {
		if err == repository.ErrCategoryServiceNotFound {
			respondError(w, http.StatusNotFound, "Assignment not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete assignment", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Service assignment removed successfully",
	})
}

