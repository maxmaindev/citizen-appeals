package service

import (
	"context"
	"log"

	"citizen-appeals/internal/models"
	"citizen-appeals/internal/repository"
	"citizen-appeals/pkg/classification"
)

// AppealService contains business logic related to appeals.
// It sits between HTTP handlers and the repository layer.
type AppealService struct {
	repo                 *repository.AppealRepository
	serviceRepo          *repository.ServiceRepository
	classifier           *classification.Classifier
	systemSettingsLoader func(context.Context) (*models.SystemSettings, error)
}

// NewAppealService creates a new AppealService instance.
func NewAppealService(
	repo *repository.AppealRepository,
	serviceRepo *repository.ServiceRepository,
	classifier *classification.Classifier,
	systemSettingsLoader func(context.Context) (*models.SystemSettings, error),
) *AppealService {
	return &AppealService{
		repo:                 repo,
		serviceRepo:          serviceRepo,
		classifier:           classifier,
		systemSettingsLoader: systemSettingsLoader,
	}
}

// CreateAppeal encapsulates the logic of creating a new appeal:
// - applies default / provided priority
// - automatically assigns service ONLY through classification service
// - persists the appeal via repository.
func (s *AppealService) CreateAppeal(
	ctx context.Context,
	userID int64,
	req models.CreateAppealRequest,
) (*models.Appeal, error) {
	priority := 2 // Default priority (середній)
	if req.Priority != nil {
		priority = *req.Priority
	}

	categoryID := req.CategoryID
	appeal := &models.Appeal{
		UserID:      userID,
		CategoryID:  &categoryID,
		Title:       req.Title,
		Description: req.Description,
		Address:     req.Address,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Status:      models.StatusNew,
		Priority:    priority,
	}

	if err := s.repo.Create(ctx, appeal); err != nil {
		return nil, err
	}

	// Assign service ONLY through classification service
	if s.classifier != nil {
		// Load confidence threshold from system settings
		if s.systemSettingsLoader != nil {
			settings, err := s.systemSettingsLoader(ctx)
			if err == nil && settings != nil {
				s.classifier.SetConfidenceThreshold(settings.ConfidenceThreshold)
			}
		}

		// Use only description for classification (title is often too short/generic)
		serviceName, confidence, err := s.classifier.ClassifyAppeal(ctx, req.Description)
		if err != nil {
			log.Printf("Classification error: %v", err)
		} else if serviceName != "" {
			// Try to find service by name
			service, err := s.serviceRepo.GetByName(ctx, serviceName)
			if err == nil && service != nil {
				appeal.ServiceID = &service.ID
				// Update appeal with service_id from classification
				if err := s.repo.Update(ctx, appeal); err != nil {
					log.Printf("Failed to assign service from classification: %v", err)
				} else {
					log.Printf("Assigned service '%s' from classification (confidence: %.2f)", serviceName, confidence)
				}
			} else {
				log.Printf("Service '%s' from classification not found in database", serviceName)
			}
		} else {
			log.Printf("Classification service did not return a service (confidence: %.2f)", confidence)
		}
	} else {
		log.Printf("Classification service is disabled, service will not be assigned automatically")
	}

	return appeal, nil
}

// ClassifyText classifies text and returns suggested service name and confidence
func (s *AppealService) ClassifyText(ctx context.Context, text string) (string, float64, error) {
	if s.classifier == nil {
		return "", 0, nil
	}

	// Load confidence threshold from system settings
	if s.systemSettingsLoader != nil {
		settings, err := s.systemSettingsLoader(ctx)
		if err == nil && settings != nil {
			s.classifier.SetConfidenceThreshold(settings.ConfidenceThreshold)
		}
	}

	return s.classifier.ClassifyAppeal(ctx, text)
}
