package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"citizen-appeals/internal/middleware"
	"citizen-appeals/internal/repository"
)

type NotificationHandler struct {
	repo *repository.NotificationRepository
}

func NewNotificationHandler(repo *repository.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{repo: repo}
}

// List retrieves all notifications for the current user
func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	// Parse pagination params
	page := 1
	limit := 50
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

	offset := (page - 1) * limit
	notifications, err := h.repo.GetByUserID(r.Context(), userID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get notifications", err)
		return
	}

	respondJSON(w, http.StatusOK, notifications)
}

// GetUnreadCount returns the count of unread notifications
func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	count, err := h.repo.GetUnreadCount(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get unread count", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"count": count,
	})
}

// MarkAsRead marks a notification as read
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid notification ID", err)
		return
	}

	if err := h.repo.MarkAsRead(r.Context(), id, userID); err != nil {
		if err == repository.ErrNotificationNotFound {
			respondError(w, http.StatusNotFound, "Notification not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to mark notification as read", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Notification marked as read"})
}

// MarkAllAsRead marks all notifications for the current user as read
func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	if err := h.repo.MarkAllAsRead(r.Context(), userID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to mark all notifications as read", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "All notifications marked as read"})
}

// Delete deletes a notification
func (h *NotificationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid notification ID", err)
		return
	}

	if err := h.repo.Delete(r.Context(), id, userID); err != nil {
		if err == repository.ErrNotificationNotFound {
			respondError(w, http.StatusNotFound, "Notification not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete notification", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Notification deleted"})
}

