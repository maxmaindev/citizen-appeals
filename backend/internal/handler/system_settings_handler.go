package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"citizen-appeals/internal/models"
)

// SystemSettingsHandler handles reading and updating system-wide settings
// that are stored in a JSON file on disk.
type SystemSettingsHandler struct {
	filePath string
	mu       sync.RWMutex
}

func NewSystemSettingsHandler(filePath string) *SystemSettingsHandler {
	return &SystemSettingsHandler{
		filePath: filePath,
	}
}

// GetSettings returns current system settings (public method for internal use)
func (h *SystemSettingsHandler) GetSettings() (*models.SystemSettings, error) {
	return h.load()
}

func (h *SystemSettingsHandler) load() (*models.SystemSettings, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := os.ReadFile(h.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return sensible defaults if file does not exist yet
			return &models.SystemSettings{
				CityName:            "Київ",
				MapCenterLat:        50.4501,
				MapCenterLng:        30.5234,
				MapZoom:             13,
				ConfidenceThreshold: 0.8,
			}, nil
		}
		return nil, err
	}

	var settings models.SystemSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	// Set default confidence threshold if not set
	if settings.ConfidenceThreshold == 0 {
		settings.ConfidenceThreshold = 0.8
	}

	return &settings, nil
}

func (h *SystemSettingsHandler) save(settings *models.SystemSettings) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(h.filePath), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(h.filePath, data, 0o644)
}

// Get returns current system settings (admin-only, routed in main.go).
func (h *SystemSettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	settings, err := h.load()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to load system settings", err)
		return
	}

	respondJSON(w, http.StatusOK, settings)
}

// Update updates and persists system settings (admin-only, routed in main.go).
func (h *SystemSettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req models.SystemSettings
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Basic sanity defaults if client sends zero values
	if req.CityName == "" {
		req.CityName = "Київ"
	}
	if req.MapZoom == 0 {
		req.MapZoom = 13
	}
	if req.ConfidenceThreshold == 0 {
		req.ConfidenceThreshold = 0.8
	}
	// Validate confidence threshold range (0.0 to 1.0)
	if req.ConfidenceThreshold < 0.0 {
		req.ConfidenceThreshold = 0.0
	}
	if req.ConfidenceThreshold > 1.0 {
		req.ConfidenceThreshold = 1.0
	}

	if err := h.save(&req); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save system settings", err)
		return
	}

	respondJSON(w, http.StatusOK, req)
}
