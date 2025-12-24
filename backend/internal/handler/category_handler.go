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

type CategoryHandler struct {
	categoryRepo *repository.CategoryRepository
	validator    *validator.Validate
}

func NewCategoryHandler(categoryRepo *repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{
		categoryRepo: categoryRepo,
		validator:    validator.New(),
	}
}

// List retrieves all categories
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	includeInactive := r.URL.Query().Get("include_inactive") == "true"

	categories, err := h.categoryRepo.List(r.Context(), includeInactive)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list categories", err)
		return
	}

	respondJSON(w, http.StatusOK, categories)
}

// GetByID retrieves a category by ID
func (h *CategoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid category ID", err)
		return
	}

	category, err := h.categoryRepo.GetByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrCategoryNotFound {
			respondError(w, http.StatusNotFound, "Category not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get category", err)
		return
	}

	respondJSON(w, http.StatusOK, category)
}

// Create creates a new category (admin only)
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	category := &models.Category{
		Name:            req.Name,
		Description:     req.Description,
		DefaultPriority: req.DefaultPriority,
		IsActive:        true,
	}

	if err := h.categoryRepo.Create(r.Context(), category); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create category", err)
		return
	}

	respondJSON(w, http.StatusCreated, category)
}

// Update updates a category (admin only)
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid category ID", err)
		return
	}

	category, err := h.categoryRepo.GetByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrCategoryNotFound {
			respondError(w, http.StatusNotFound, "Category not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get category", err)
		return
	}

	var req models.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Update fields
	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.Description != nil {
		category.Description = *req.Description
	}
	if req.DefaultPriority != nil {
		category.DefaultPriority = *req.DefaultPriority
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if err := h.categoryRepo.Update(r.Context(), category); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update category", err)
		return
	}

	respondJSON(w, http.StatusOK, category)
}

// Delete soft deletes a category (admin only)
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid category ID", err)
		return
	}

	if err := h.categoryRepo.Delete(r.Context(), id); err != nil {
		if err == repository.ErrCategoryNotFound {
			respondError(w, http.StatusNotFound, "Category not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete category", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Category deleted successfully",
	})
}
