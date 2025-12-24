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
	"citizen-appeals/internal/service"
)

type CommentHandler struct {
	commentRepo         *repository.CommentRepository
	appealRepo          *repository.AppealRepository
	validator           *validator.Validate
	notificationService *service.NotificationService
}

func NewCommentHandler(
	commentRepo *repository.CommentRepository,
	appealRepo *repository.AppealRepository,
	notificationService *service.NotificationService,
) *CommentHandler {
	return &CommentHandler{
		commentRepo:         commentRepo,
		appealRepo:          appealRepo,
		validator:           validator.New(),
		notificationService: notificationService,
	}
}

// Create creates a new comment
func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	userRole, _ := middleware.GetUserRole(r.Context())

	appealIDStr := chi.URLParam(r, "appeal_id")
	appealID, err := strconv.ParseInt(appealIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid appeal ID", err)
		return
	}

	// Verify appeal exists
	_, err = h.appealRepo.GetByID(r.Context(), appealID)
	if err != nil {
		if err == repository.ErrAppealNotFound {
			respondError(w, http.StatusNotFound, "Appeal not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get appeal", err)
		return
	}

	var req models.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Citizens cannot create internal comments
	if userRole == models.RoleCitizen && req.IsInternal {
		respondError(w, http.StatusForbidden, "Citizens cannot create internal comments")
		return
	}

	comment := &models.Comment{
		AppealID:   appealID,
		UserID:     userID,
		Text:       req.Text,
		IsInternal: req.IsInternal,
	}

	if err := h.commentRepo.Create(r.Context(), comment); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create comment", err)
		return
	}

	// Get full comment with user info
	fullComment, err := h.commentRepo.GetByID(r.Context(), comment.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get created comment", err)
		return
	}

	// Send notification to appeal creator (if not internal comment or if comment is from different user)
	if h.notificationService != nil && !req.IsInternal {
		if err := h.notificationService.SendCommentAdded(r.Context(), appealID, userID, req.Text); err != nil {
			// Log error but don't fail the request
		}
	}

	respondJSON(w, http.StatusCreated, fullComment)
}

// GetByAppealID retrieves all comments for an appeal
func (h *CommentHandler) GetByAppealID(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	userRole, _ := middleware.GetUserRole(r.Context())

	appealIDStr := chi.URLParam(r, "appeal_id")
	appealID, err := strconv.ParseInt(appealIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid appeal ID", err)
		return
	}

	// Verify appeal exists
	_, err = h.appealRepo.GetByID(r.Context(), appealID)
	if err != nil {
		if err == repository.ErrAppealNotFound {
			respondError(w, http.StatusNotFound, "Appeal not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get appeal", err)
		return
	}

	comments, err := h.commentRepo.GetByAppealID(r.Context(), appealID, userID, userRole)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get comments", err)
		return
	}

	respondJSON(w, http.StatusOK, comments)
}

// Update updates a comment
func (h *CommentHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	userRole, _ := middleware.GetUserRole(r.Context())

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid comment ID", err)
		return
	}

	// Get existing comment
	comment, err := h.commentRepo.GetByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrCommentNotFound {
			respondError(w, http.StatusNotFound, "Comment not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get comment", err)
		return
	}

	// Check permissions: user can only update their own comments
	// Admins and dispatchers can update any comment
	if comment.UserID != userID && userRole != models.RoleAdmin && userRole != models.RoleDispatcher {
		respondError(w, http.StatusForbidden, "You don't have permission to update this comment")
		return
	}

	var req models.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Citizens cannot create internal comments
	if userRole == models.RoleCitizen && req.IsInternal {
		respondError(w, http.StatusForbidden, "Citizens cannot create internal comments")
		return
	}

	comment.Text = req.Text
	comment.IsInternal = req.IsInternal

	if err := h.commentRepo.Update(r.Context(), comment); err != nil {
		if err == repository.ErrCommentNotFound {
			respondError(w, http.StatusNotFound, "Comment not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update comment", err)
		return
	}

	// Get updated comment with user info
	updatedComment, err := h.commentRepo.GetByID(r.Context(), comment.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get updated comment", err)
		return
	}

	respondJSON(w, http.StatusOK, updatedComment)
}

// Delete deletes a comment
func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	userRole, _ := middleware.GetUserRole(r.Context())

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid comment ID", err)
		return
	}

	// Get existing comment
	comment, err := h.commentRepo.GetByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrCommentNotFound {
			respondError(w, http.StatusNotFound, "Comment not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get comment", err)
		return
	}

	// Check permissions: user can only delete their own comments
	// Admins and dispatchers can delete any comment
	if comment.UserID != userID && userRole != models.RoleAdmin && userRole != models.RoleDispatcher {
		respondError(w, http.StatusForbidden, "You don't have permission to delete this comment")
		return
	}

	if err := h.commentRepo.Delete(r.Context(), id); err != nil {
		if err == repository.ErrCommentNotFound {
			respondError(w, http.StatusNotFound, "Comment not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete comment", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Comment deleted successfully",
	})
}

