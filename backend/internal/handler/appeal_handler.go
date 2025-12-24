package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"citizen-appeals/internal/middleware"
	"citizen-appeals/internal/models"
	"citizen-appeals/internal/repository"
	"citizen-appeals/internal/service"
)

type AppealHandler struct {
	appealRepo         *repository.AppealRepository
	validator          *validator.Validate
	service            *service.AppealService
	notificationService *service.NotificationService
}

func NewAppealHandler(
	appealRepo *repository.AppealRepository,
	service *service.AppealService,
	notificationService *service.NotificationService,
) *AppealHandler {
	return &AppealHandler{
		appealRepo:          appealRepo,
		validator:           validator.New(),
		service:             service,
		notificationService: notificationService,
	}
}

// Classify classifies appeal text and returns suggested service
func (h *AppealHandler) Classify(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.Text == "" {
		respondError(w, http.StatusBadRequest, "Text is required", nil)
		return
	}

	// Use the service's classifier (it's accessible through the service)
	serviceName, confidence, err := h.service.ClassifyText(r.Context(), req.Text)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Classification failed", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"service":    serviceName,
		"confidence": confidence,
	})
}

// Create creates a new appeal
func (h *AppealHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	var req models.CreateAppealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	appeal, err := h.service.CreateAppeal(r.Context(), userID, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create appeal", err)
		return
	}

	// Send notification to dispatchers about new appeal
	if h.notificationService != nil {
		if err := h.notificationService.SendAppealCreated(r.Context(), appeal); err != nil {
			// Log error but don't fail the request
			log.Printf("Failed to send appeal created notification: %v", err)
		}
	}

	respondJSON(w, http.StatusCreated, appeal)
}

// GetByID retrieves an appeal by ID
func (h *AppealHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid appeal ID", err)
		return
	}

	appeal, err := h.appealRepo.GetByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrAppealNotFound {
			respondError(w, http.StatusNotFound, "Appeal not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get appeal", err)
		return
	}

	// All authenticated users can view any appeal
	respondJSON(w, http.StatusOK, appeal)
}

// List retrieves appeals with filters
func (h *AppealHandler) List(w http.ResponseWriter, r *http.Request) {
	filters := &models.AppealFilters{
		Page:  1,
		Limit: 20,
	}

	// Parse query parameters
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			filters.Page = page
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filters.Limit = limit
		}
	}

	if status := r.URL.Query().Get("status"); status != "" {
		appealStatus := models.AppealStatus(status)
		filters.Status = &appealStatus
	}

	if categoryIDStr := r.URL.Query().Get("category_id"); categoryIDStr != "" {
		if categoryID, err := strconv.ParseInt(categoryIDStr, 10, 64); err == nil {
			filters.CategoryID = &categoryID
		}
	}

	if serviceIDStr := r.URL.Query().Get("service_id"); serviceIDStr != "" {
		if serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64); err == nil {
			filters.ServiceID = &serviceID
		}
	}

	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			filters.UserID = &userID
		}
	}

	if search := r.URL.Query().Get("search"); search != "" {
		filters.Search = &search
	}

	if fromDateStr := r.URL.Query().Get("from_date"); fromDateStr != "" {
		if fromDate, err := time.Parse(time.RFC3339, fromDateStr); err == nil {
			filters.FromDate = &fromDate
		}
	}

	if toDateStr := r.URL.Query().Get("to_date"); toDateStr != "" {
		if toDate, err := time.Parse(time.RFC3339, toDateStr); err == nil {
			filters.ToDate = &toDate
		}
	}

	filters.SortBy = r.URL.Query().Get("sort_by")
	filters.SortOrder = r.URL.Query().Get("sort_order")

	// All authenticated users can see all appeals
	// Executors see all appeals (like dispatcher), not just assigned ones
	// They can then assign appeals to themselves

	appeals, total, err := h.appealRepo.List(r.Context(), filters)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list appeals", err)
		return
	}

	totalPages := int(total) / filters.Limit
	if int(total)%filters.Limit > 0 {
		totalPages++
	}

	response := models.PaginatedResponse{
		Items:      appeals,
		Total:      total,
		Page:       filters.Page,
		Limit:      filters.Limit,
		TotalPages: totalPages,
	}

	respondJSON(w, http.StatusOK, response)
}

// Update updates an appeal (only for appeal creator)
func (h *AppealHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid appeal ID", err)
		return
	}

	// Get existing appeal
	appeal, err := h.appealRepo.GetByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrAppealNotFound {
			respondError(w, http.StatusNotFound, "Appeal not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get appeal", err)
		return
	}

	// Get user role
	userRole, _ := middleware.GetUserRole(r.Context())

	// Check permissions: creator can update, dispatcher/admin can update category
	isCreator := appeal.UserID == userID
	isDispatcherOrAdmin := userRole == models.RoleDispatcher || userRole == models.RoleAdmin

	if !isCreator && !isDispatcherOrAdmin {
		respondError(w, http.StatusForbidden, "You don't have permission to update this appeal")
		return
	}

	// Creator can only update if status is 'new'
	// Dispatcher/admin can update category at any time
	if isCreator && appeal.Status != models.StatusNew {
		respondError(w, http.StatusBadRequest, "Cannot update appeal that is already being processed")
		return
	}

	var req models.UpdateAppealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Update fields
	if req.Title != nil && isCreator {
		appeal.Title = *req.Title
	}
	if req.Description != nil && isCreator {
		appeal.Description = *req.Description
	}
	if req.CategoryID != nil {
		appeal.CategoryID = req.CategoryID
	}
	if req.Address != nil {
		appeal.Address = *req.Address
	}
	if req.Latitude != nil {
		appeal.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		appeal.Longitude = *req.Longitude
	}

	if err := h.appealRepo.Update(r.Context(), appeal); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update appeal", err)
		return
	}

	respondJSON(w, http.StatusOK, appeal)
}

// UpdateStatus updates appeal status
func (h *AppealHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	userRole, _ := middleware.GetUserRole(r.Context())

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid appeal ID", err)
		return
	}

	var req models.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Get existing appeal
	appeal, err := h.appealRepo.GetByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrAppealNotFound {
			respondError(w, http.StatusNotFound, "Appeal not found", err)
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to get appeal", err)
		}
		return
	}

	// Check permissions based on role
	canUpdate := false
	switch userRole {
	case models.RoleAdmin, models.RoleDispatcher:
		canUpdate = true
	case models.RoleExecutor:
		// Executors can update appeals if they belong to the service assigned to the appeal
		if appeal.ServiceID != nil {
			// Check if executor belongs to the service (would need to query user_services table)
			// For now, allow executors to update any appeal (can be refined later)
			canUpdate = true
		}
	}

	if !canUpdate {
		respondError(w, http.StatusForbidden, "You don't have permission to update this appeal status")
		return
	}

	// Log the update attempt
	if req.Comment != nil && *req.Comment != "" {
		log.Printf("Updating appeal %d status from handler: %s -> %s with comment: %s", id, appeal.Status, req.Status, *req.Comment)
	} else {
		log.Printf("Updating appeal %d status from handler: %s -> %s (no comment)", id, appeal.Status, req.Status)
	}

	oldStatus := appeal.Status
	if err := h.appealRepo.UpdateStatus(r.Context(), id, req.Status, userID, req.Comment); err != nil {
		log.Printf("Error updating status for appeal %d: %v", id, err)
		respondError(w, http.StatusInternalServerError, "Failed to update status", err)
		return
	}

	// Send notification to appeal creator about status change
	if h.notificationService != nil {
		// Get updated appeal
		updatedAppeal, err := h.appealRepo.GetByID(r.Context(), id)
		if err == nil {
			updatedAppeal.Status = req.Status
			if err := h.notificationService.SendStatusChanged(r.Context(), updatedAppeal, oldStatus); err != nil {
				log.Printf("Failed to send status change notification: %v", err)
			}
		}
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Status updated successfully",
	})
}

// UpdatePriority updates appeal priority
func (h *AppealHandler) UpdatePriority(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	userRole, _ := middleware.GetUserRole(r.Context())

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid appeal ID", err)
		return
	}

	var req models.UpdatePriorityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Permissions: same as status change
	canUpdate := false
	switch userRole {
	case models.RoleAdmin, models.RoleDispatcher, models.RoleExecutor:
		canUpdate = true
	}

	if !canUpdate {
		respondError(w, http.StatusForbidden, "You don't have permission to update this appeal priority")
		return
	}

	if err := h.appealRepo.UpdatePriority(r.Context(), id, req.Priority, userID); err != nil {
		if err == repository.ErrAppealNotFound {
			respondError(w, http.StatusNotFound, "Appeal not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update priority", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Priority updated successfully",
	})
}

// Assign assigns appeal to service (dispatcher/admin)
func (h *AppealHandler) Assign(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	userRole, _ := middleware.GetUserRole(r.Context())

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid appeal ID", err)
		return
	}

	var req models.AssignAppealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Only dispatcher and admin can assign appeals
	if userRole != models.RoleDispatcher && userRole != models.RoleAdmin {
		respondError(w, http.StatusForbidden, "Only dispatchers and admins can assign appeals")
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	if err := h.appealRepo.Assign(r.Context(), id, req.ServiceID, req.Priority, userID); err != nil {
		if err == repository.ErrAppealNotFound {
			respondError(w, http.StatusNotFound, "Appeal not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to assign appeal", err)
		return
	}

	// Send notification to service executors
	if h.notificationService != nil {
		appeal, err := h.appealRepo.GetByID(r.Context(), id)
		if err == nil {
			if err := h.notificationService.SendAppealAssigned(r.Context(), appeal); err != nil {
				log.Printf("Failed to send assignment notification: %v", err)
			}
		}
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Appeal assigned successfully",
	})
}

// GetStatistics returns appeal statistics (admin/dispatcher only)
func (h *AppealHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	var fromDate, toDate *time.Time

	if fromDateStr := r.URL.Query().Get("from_date"); fromDateStr != "" {
		if fd, err := time.Parse(time.RFC3339, fromDateStr); err == nil {
			fromDate = &fd
		}
	}

	if toDateStr := r.URL.Query().Get("to_date"); toDateStr != "" {
		if td, err := time.Parse(time.RFC3339, toDateStr); err == nil {
			toDate = &td
		}
	}

	stats, err := h.appealRepo.GetStatistics(r.Context(), fromDate, toDate)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get statistics", err)
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// GetDispatcherDashboard returns dashboard data for dispatcher
func (h *AppealHandler) GetDispatcherDashboard(w http.ResponseWriter, r *http.Request) {
	dashboard, err := h.appealRepo.GetDispatcherDashboard(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get dispatcher dashboard", err)
		return
	}

	respondJSON(w, http.StatusOK, dashboard)
}

// GetAdminDashboard returns dashboard data for admin
func (h *AppealHandler) GetAdminDashboard(w http.ResponseWriter, r *http.Request) {
	dashboard, err := h.appealRepo.GetAdminDashboard(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get admin dashboard", err)
		return
	}

	respondJSON(w, http.StatusOK, dashboard)
}

// GetExecutorDashboard returns dashboard data for executor
func (h *AppealHandler) GetExecutorDashboard(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	dashboard, err := h.appealRepo.GetExecutorDashboard(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get executor dashboard", err)
		return
	}

	respondJSON(w, http.StatusOK, dashboard)
}

// GetServiceStatistics returns detailed statistics for a specific service
func (h *AppealHandler) GetServiceStatistics(w http.ResponseWriter, r *http.Request) {
	serviceIDStr := chi.URLParam(r, "service_id")
	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid service ID", err)
		return
	}

	stats, err := h.appealRepo.GetServiceStatistics(r.Context(), serviceID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get service statistics", err)
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// GetHistory retrieves appeal history (status changes)
func (h *AppealHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid appeal ID", err)
		return
	}

	history, err := h.appealRepo.GetHistory(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get appeal history", err)
		return
	}

	respondJSON(w, http.StatusOK, history)
}
