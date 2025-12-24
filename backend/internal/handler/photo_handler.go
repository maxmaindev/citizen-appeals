package handler

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"citizen-appeals/internal/middleware"
	"citizen-appeals/internal/models"
	"citizen-appeals/internal/repository"
	"citizen-appeals/pkg/storage"
)

const (
	maxPhotosPerAppeal = 5
	maxPhotoSize       = 5 * 1024 * 1024 // 5MB
)

type PhotoHandler struct {
	photoRepo  *repository.PhotoRepository
	appealRepo *repository.AppealRepository
	storage    storage.Storage
}

func NewPhotoHandler(photoRepo *repository.PhotoRepository, appealRepo *repository.AppealRepository, storage storage.Storage) *PhotoHandler {
	return &PhotoHandler{
		photoRepo:  photoRepo,
		appealRepo: appealRepo,
		storage:    storage,
	}
}

// Upload handles photo upload for an appeal
func (h *PhotoHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	// Parse appeal ID
	appealIDStr := chi.URLParam(r, "id")
	appealID, err := strconv.ParseInt(appealIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid appeal ID", err)
		return
	}

	// Check if appeal exists and user has permission
	appeal, err := h.appealRepo.GetByID(r.Context(), appealID)
	if err != nil {
		if err == repository.ErrAppealNotFound {
			respondError(w, http.StatusNotFound, "Appeal not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get appeal", err)
		return
	}

	// Check permissions: citizen can only upload to their own appeals, executor can upload result photos
	userRole, _ := middleware.GetUserRole(r.Context())
	canUpload := false
	isResultPhoto := r.URL.Query().Get("result") == "true"

	if userRole == models.RoleCitizen {
		// Citizens can only upload to their own appeals, and only initial photos (not result photos)
		if appeal.UserID == userID && !isResultPhoto {
			canUpload = true
		}
	} else if userRole == models.RoleExecutor {
		// Executors can upload result photos to appeals assigned to their service
		if appeal.ServiceID != nil && isResultPhoto {
			canUpload = true
		}
	} else if userRole == models.RoleDispatcher || userRole == models.RoleAdmin {
		// Dispatchers and admins can upload to any appeal
		canUpload = true
	}

	if !canUpload {
		respondError(w, http.StatusForbidden, "You don't have permission to upload photos to this appeal")
		return
	}

	// Check current photo count
	currentCount, err := h.photoRepo.CountByAppealID(r.Context(), appealID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to count photos", err)
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(maxPhotoSize * maxPhotosPerAppeal)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse form", err)
		return
	}

	files := r.MultipartForm.File["photos"]
	if len(files) == 0 {
		respondError(w, http.StatusBadRequest, "No files provided")
		return
	}

	// Check total file count
	if currentCount+len(files) > maxPhotosPerAppeal {
		respondError(w, http.StatusBadRequest, "Too many photos. Maximum 5 photos per appeal")
		return
	}

	uploadedPhotos := make([]*models.UploadPhotoResponse, 0)

	// Process each file
	for _, fileHeader := range files {
		// Validate file
		if err := storage.ValidateFile(fileHeader, maxPhotoSize, maxPhotosPerAppeal, currentCount+len(uploadedPhotos)); err != nil {
			respondError(w, http.StatusBadRequest, err.Error(), err)
			return
		}

		// Open file
		file, err := fileHeader.Open()
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to open file", err)
			return
		}

		// Save file to storage
		filePath, mimeType, fileSize, err := h.storage.Save(file, fileHeader, appealID, isResultPhoto)
		file.Close()
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to save file", err)
			return
		}

		// Create photo record in database
		photo := &models.Photo{
			AppealID:      &appealID,
			FilePath:      filePath,
			FileName:      fileHeader.Filename,
			FileSize:      fileSize,
			MimeType:      mimeType,
			IsResultPhoto: isResultPhoto,
		}

		if err := h.photoRepo.Create(r.Context(), photo); err != nil {
			// Clean up file if database insert fails
			h.storage.Delete(filePath)
			respondError(w, http.StatusInternalServerError, "Failed to save photo record", err)
			return
		}

		uploadedPhotos = append(uploadedPhotos, &models.UploadPhotoResponse{
			ID:       photo.ID,
			FileName: photo.FileName,
			FileSize: photo.FileSize,
			URL:      h.storage.GetURL(filePath),
		})
	}

	respondJSON(w, http.StatusCreated, uploadedPhotos)
}

// Get retrieves a photo by ID
func (h *PhotoHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid photo ID", err)
		return
	}

	photo, err := h.photoRepo.GetByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrPhotoNotFound {
			respondError(w, http.StatusNotFound, "Photo not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get photo", err)
		return
	}

	// Check if user has permission to view the appeal
	userID, _ := middleware.GetUserID(r.Context())
	userRole, _ := middleware.GetUserRole(r.Context())

	appeal, err := h.appealRepo.GetByID(r.Context(), *photo.AppealID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get appeal", err)
		return
	}

	// Check permissions
	if userRole == models.RoleCitizen && appeal.UserID != userID {
		respondError(w, http.StatusForbidden, "You can only view photos of your own appeals")
		return
	}

	// Get file from storage
	file, err := h.storage.Get(photo.FilePath)
	if err != nil {
		if err == storage.ErrFileNotFound {
			respondError(w, http.StatusNotFound, "Photo file not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve photo", err)
		return
	}
	defer file.Close()

	// Set headers
	w.Header().Set("Content-Type", photo.MimeType)
	w.Header().Set("Content-Disposition", `inline; filename="`+photo.FileName+`"`)
	w.Header().Set("Content-Length", strconv.FormatInt(photo.FileSize, 10))

	// Stream file to response
	_, err = io.Copy(w, file)
	if err != nil {
		log.Printf("Error streaming photo: %v", err)
	}
}

// List retrieves all photos for an appeal
func (h *PhotoHandler) List(w http.ResponseWriter, r *http.Request) {
	appealIDStr := chi.URLParam(r, "id")
	appealID, err := strconv.ParseInt(appealIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid appeal ID", err)
		return
	}

	// Check if appeal exists and user has permission
	appeal, err := h.appealRepo.GetByID(r.Context(), appealID)
	if err != nil {
		if err == repository.ErrAppealNotFound {
			respondError(w, http.StatusNotFound, "Appeal not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get appeal", err)
		return
	}

	// Check permissions
	userID, _ := middleware.GetUserID(r.Context())
	userRole, _ := middleware.GetUserRole(r.Context())

	if userRole == models.RoleCitizen && appeal.UserID != userID {
		respondError(w, http.StatusForbidden, "You can only view photos of your own appeals")
		return
	}

	// Get photos
	photos, err := h.photoRepo.GetByAppealID(r.Context(), appealID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get photos", err)
		return
	}

	// Add URLs to photos
	response := make([]map[string]interface{}, len(photos))
	for i, photo := range photos {
		response[i] = map[string]interface{}{
			"id":              photo.ID,
			"file_name":       photo.FileName,
			"file_size":       photo.FileSize,
			"mime_type":       photo.MimeType,
			"is_result_photo": photo.IsResultPhoto,
			"url":             h.storage.GetURL(photo.FilePath),
			"uploaded_at":     photo.UploadedAt,
		}
	}

	respondJSON(w, http.StatusOK, response)
}

// Delete deletes a photo
func (h *PhotoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid photo ID", err)
		return
	}

	// Get photo
	photo, err := h.photoRepo.GetByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrPhotoNotFound {
			respondError(w, http.StatusNotFound, "Photo not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get photo", err)
		return
	}

	// Check permissions
	userID, _ := middleware.GetUserID(r.Context())
	userRole, _ := middleware.GetUserRole(r.Context())

	appeal, err := h.appealRepo.GetByID(r.Context(), *photo.AppealID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get appeal", err)
		return
	}

	canDelete := false
	if userRole == models.RoleCitizen {
		// Citizens can only delete their own photos from their own appeals
		if appeal.UserID == userID && !photo.IsResultPhoto {
			canDelete = true
		}
	} else if userRole == models.RoleExecutor {
		// Executors can delete result photos from appeals assigned to their service
		if appeal.ServiceID != nil && photo.IsResultPhoto {
			canDelete = true
		}
	} else if userRole == models.RoleDispatcher || userRole == models.RoleAdmin {
		// Dispatchers and admins can delete any photo
		canDelete = true
	}

	if !canDelete {
		respondError(w, http.StatusForbidden, "You don't have permission to delete this photo")
		return
	}

	// Delete file from storage
	if err := h.storage.Delete(photo.FilePath); err != nil && err != storage.ErrFileNotFound {
		log.Printf("Warning: failed to delete file from storage: %v", err)
		// Continue with database deletion even if file deletion fails
	}

	// Delete photo record
	if err := h.photoRepo.Delete(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete photo", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Photo deleted successfully",
	})
}
