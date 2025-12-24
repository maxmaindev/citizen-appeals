package handler

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"citizen-appeals/internal/models"
	"citizen-appeals/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type ServiceHandler struct {
	serviceRepo              *repository.ServiceRepository
	embeddingRepo            *repository.ServiceEmbeddingRepository
	validator                *validator.Validate
	classificationServiceURL string
	backendURL               string
}

func NewServiceHandler(
	serviceRepo *repository.ServiceRepository,
	embeddingRepo *repository.ServiceEmbeddingRepository,
	classificationServiceURL string,
	backendURL string,
) *ServiceHandler {
	return &ServiceHandler{
		serviceRepo:              serviceRepo,
		embeddingRepo:            embeddingRepo,
		validator:                validator.New(),
		classificationServiceURL: classificationServiceURL,
		backendURL:               backendURL,
	}
}

// List retrieves all services
func (h *ServiceHandler) List(w http.ResponseWriter, r *http.Request) {
	includeInactive := r.URL.Query().Get("include_inactive") == "true"

	services, err := h.serviceRepo.List(r.Context(), includeInactive)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list services", err)
		return
	}

	// Load keywords from MongoDB (keywords are only stored in MongoDB)
	for _, service := range services {
		keywords, err := h.embeddingRepo.GetKeywords(r.Context(), service.ID)
		if err == nil {
			service.Keywords = keywords
		}
	}

	respondJSON(w, http.StatusOK, services)
}

// GetByID retrieves a service by ID
func (h *ServiceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid service ID", err)
		return
	}

	service, err := h.serviceRepo.GetByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrServiceNotFound {
			respondError(w, http.StatusNotFound, "Service not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get service", err)
		return
	}

	// Load keywords from MongoDB (keywords are only stored in MongoDB)
	keywords, err := h.embeddingRepo.GetKeywords(r.Context(), service.ID)
	if err == nil {
		service.Keywords = keywords
	}

	respondJSON(w, http.StatusOK, service)
}

// Create creates a new service (admin only)
func (h *ServiceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	service := &models.Service{
		Name:          req.Name,
		Description:   req.Description,
		ContactPerson: req.ContactPerson,
		ContactPhone:  req.ContactPhone,
		ContactEmail:  req.ContactEmail,
		IsActive:      true,
	}

	if err := h.serviceRepo.Create(r.Context(), service); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create service", err)
		return
	}

	// Save keywords to MongoDB if provided
	if req.Keywords != "" {
		if err := h.embeddingRepo.UpdateKeywords(r.Context(), service.ID, service.Name, req.Keywords); err != nil {
			log.Printf("Warning: Failed to save keywords to MongoDB: %v", err)
		} else {
			service.Keywords = req.Keywords
		}
	}

	respondJSON(w, http.StatusCreated, service)
}

// Update updates a service (admin only)
func (h *ServiceHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid service ID", err)
		return
	}

	service, err := h.serviceRepo.GetByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrServiceNotFound {
			respondError(w, http.StatusNotFound, "Service not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get service", err)
		return
	}

	var req models.UpdateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Зберігаємо старе значення keywords для перевірки змін (з MongoDB)
	var oldKeywords string
	keywordsChanged := false

	oldKeywords, _ = h.embeddingRepo.GetKeywords(r.Context(), service.ID)

	// Update fields
	if req.Name != nil {
		service.Name = *req.Name
	}
	if req.Description != nil {
		service.Description = *req.Description
	}
	if req.Keywords != nil {
		keywordsChanged = (oldKeywords != *req.Keywords)

		// Save keywords to MongoDB (keywords are only stored in MongoDB)
		if err := h.embeddingRepo.UpdateKeywords(r.Context(), service.ID, service.Name, *req.Keywords); err != nil {
			log.Printf("Warning: Failed to update keywords in MongoDB: %v", err)
		}
	}
	if req.ContactPerson != nil {
		service.ContactPerson = *req.ContactPerson
	}
	if req.ContactPhone != nil {
		service.ContactPhone = *req.ContactPhone
	}
	if req.ContactEmail != nil {
		service.ContactEmail = *req.ContactEmail
	}
	if req.IsActive != nil {
		service.IsActive = *req.IsActive
	}

	if err := h.serviceRepo.Update(r.Context(), service); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update service", err)
		return
	}

	// Load keywords from MongoDB for response (keywords are only stored in MongoDB)
	keywords, err := h.embeddingRepo.GetKeywords(r.Context(), service.ID)
	if err == nil {
		service.Keywords = keywords
	}

	// Автоматична синхронізація з сервісом класифікації при зміні keywords
	if keywordsChanged && h.classificationServiceURL != "" {
		go h.syncClassificationService()
	}

	respondJSON(w, http.StatusOK, service)
}

// syncClassificationService викликає endpoint синхронізації ML-сервісу
// Виконується асинхронно, щоб не блокувати відповідь API
func (h *ServiceHandler) syncClassificationService() {
	if h.classificationServiceURL == "" {
		return
	}

	// Створюємо HTTP клієнт з таймаутом
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Формуємо URL для синхронізації
	syncURL := h.classificationServiceURL + "/sync"

	// Формуємо URL бекенду для отримання служб
	backendServicesURL := h.backendURL + "/api/services/for-classification"

	// Створюємо запит
	reqBody := map[string]interface{}{
		"backend_url": backendServicesURL,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("Failed to marshal sync request: %v", err)
		return
	}

	req, err := http.NewRequest("POST", syncURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to create sync request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	// Виконуємо запит
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to sync classification service: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("Classification service synced successfully after keywords update")
	} else {
		log.Printf("Classification service sync returned status %d", resp.StatusCode)
	}
}

// Delete soft deletes a service (admin only)
func (h *ServiceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid service ID", err)
		return
	}

	if err := h.serviceRepo.Delete(r.Context(), id); err != nil {
		if err == repository.ErrServiceNotFound {
			respondError(w, http.StatusNotFound, "Service not found", err)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete service", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Service deleted successfully",
	})
}

// GetForClassification returns services with name and combined description+keywords for ML classification
func (h *ServiceHandler) GetForClassification(w http.ResponseWriter, r *http.Request) {
	// Get from MongoDB (keywords are only stored in MongoDB)
	services, err := h.embeddingRepo.GetAllForClassification(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get services from MongoDB", err)
		return
	}

	// Convert to array format
	result := make([]map[string]string, 0, len(services))
	for name, description := range services {
		result = append(result, map[string]string{
			"name":        name,
			"description": description,
		})
	}

	respondJSON(w, http.StatusOK, result)
}
