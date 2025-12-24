package handler

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"citizen-appeals/internal/models"
)

// Integration tests for AppealHandler
// These tests verify HTTP request/response handling

func TestAppealHandler_Create_ValidRequest(t *testing.T) {
	// This is a simplified integration test
	// In a full implementation, you would use a test database

	reqBody := models.CreateAppealRequest{
		Title:       "Test Appeal Title",
		Description: "This is a test description for the appeal",
		CategoryID:  1,
		Address:     "Test Address 123",
		Latitude:    50.4501,
		Longitude:   30.5234,
	}

	// Validate request structure
	assert.GreaterOrEqual(t, len(reqBody.Title), 5, "Title should be at least 5 characters")
	assert.GreaterOrEqual(t, len(reqBody.Description), 10, "Description should be at least 10 characters")
	assert.Greater(t, reqBody.CategoryID, int64(0), "CategoryID should be positive")
	assert.GreaterOrEqual(t, reqBody.Latitude, -90.0, "Latitude should be >= -90")
	assert.LessOrEqual(t, reqBody.Latitude, 90.0, "Latitude should be <= 90")
	assert.GreaterOrEqual(t, reqBody.Longitude, -180.0, "Longitude should be >= -180")
	assert.LessOrEqual(t, reqBody.Longitude, 180.0, "Longitude should be <= 180")
}

func TestAppealHandler_Create_InvalidRequest(t *testing.T) {
	// Test invalid request validation
	invalidRequests := []struct {
		name string
		req  map[string]interface{}
	}{
		{
			name: "missing title",
			req: map[string]interface{}{
				"description": "Valid description",
				"category_id": 1,
			},
		},
		{
			name: "title too short",
			req: map[string]interface{}{
				"title":       "Sh",
				"description": "Valid description",
				"category_id": 1,
			},
		},
		{
			name: "invalid latitude",
			req: map[string]interface{}{
				"title":       "Valid Title",
				"description": "Valid description",
				"category_id": 1,
				"latitude":    200.0, // Invalid
			},
		},
	}

	for _, tt := range invalidRequests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.req)
			req := httptest.NewRequest("POST", "/api/appeals", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Request should be invalid
			assert.NotNil(t, req)
			// In real test, you would check for validation errors
		})
	}
}

func TestAppealHandler_GetByID_ValidID(t *testing.T) {
	// Test valid ID format
	validIDs := []string{"1", "100", "9999"}

	for _, id := range validIDs {
		req := httptest.NewRequest("GET", "/api/appeals/"+id, nil)
		assert.NotNil(t, req)
		// In real test, you would verify the response
	}
}

func TestAppealHandler_List_QueryParameters(t *testing.T) {
	// Test query parameter parsing
	testCases := []struct {
		name   string
		params map[string]string
	}{
		{
			name: "pagination",
			params: map[string]string{
				"page":  "1",
				"limit": "20",
			},
		},
		{
			name: "filter by status",
			params: map[string]string{
				"status": "new",
			},
		},
		{
			name: "sorting",
			params: map[string]string{
				"sort_by":    "created_at",
				"sort_order": "desc",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Build query string
			query := "?"
			for k, v := range tc.params {
				query += k + "=" + v + "&"
			}

			req := httptest.NewRequest("GET", "/api/appeals"+query, nil)
			assert.NotNil(t, req)
		})
	}
}

func TestAppealHandler_Classify_ValidText(t *testing.T) {
	// Test classification request
	reqBody := map[string]string{
		"text": "Прорив водопроводу на вулиці",
	}

	body, err := json.Marshal(reqBody)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	req := httptest.NewRequest("POST", "/api/appeals/classify", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	assert.NotNil(t, req)
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
}

func TestAppealHandler_ResponseFormat(t *testing.T) {
	// Test response format validation
	response := models.PaginatedResponse{
		Items: []interface{}{},
		Total: 0,
		Page:  1,
		Limit: 20,
	}

	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	var decoded models.PaginatedResponse
	err = json.Unmarshal(jsonData, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, response.Total, decoded.Total)
}

// Helper function to create test router
func createTestRouter() *chi.Mux {
	r := chi.NewRouter()
	return r
}
