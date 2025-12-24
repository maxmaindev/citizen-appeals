package classification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Classifier struct {
	serviceURL          string
	enabled             bool
	httpClient          *http.Client
	confidenceThreshold float64
}

type ClassificationRequest struct {
	Text string `json:"text"`
}

type ClassificationResponse struct {
	Service         string  `json:"service"`
	Confidence      float64 `json:"confidence"`
	NeedsModeration bool    `json:"needs_moderation"`
	TopAlternatives []struct {
		Service    string  `json:"service"`
		Confidence float64 `json:"confidence"`
	} `json:"top_alternatives"`
}

func NewClassifier(serviceURL string, enabled bool) *Classifier {
	return &Classifier{
		serviceURL:          serviceURL,
		enabled:             enabled,
		confidenceThreshold: 0.8, // Default threshold
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// SetConfidenceThreshold sets the confidence threshold for classification
func (c *Classifier) SetConfidenceThreshold(threshold float64) {
	if threshold < 0.0 {
		threshold = 0.0
	}
	if threshold > 1.0 {
		threshold = 1.0
	}
	c.confidenceThreshold = threshold
}

// ClassifyAppeal classifies an appeal text and returns the suggested service name
// Returns empty string if classification is disabled or fails
func (c *Classifier) ClassifyAppeal(ctx context.Context, text string) (string, float64, error) {
	if !c.enabled || c.serviceURL == "" {
		return "", 0, nil
	}

	reqBody := ClassificationRequest{
		Text: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.serviceURL+"/classify", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to call classification service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("classification service returned status %d: %s", resp.StatusCode, string(body))
	}

	var result ClassificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, fmt.Errorf("failed to decode response: %w", err)
	}

	// Only return service if confidence is above threshold
	if result.Confidence < c.confidenceThreshold {
		return "", result.Confidence, nil
	}

	return result.Service, result.Confidence, nil
}
