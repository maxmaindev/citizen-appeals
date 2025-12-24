package service

import (
	"context"
	"fmt"

	"citizen-appeals/internal/models"
	"citizen-appeals/internal/repository"
)

type NotificationService struct {
	repo        *repository.NotificationRepository
	userRepo    *repository.UserRepository
	appealRepo  *repository.AppealRepository
	serviceRepo *repository.ServiceRepository
}

func NewNotificationService(
	repo *repository.NotificationRepository,
	userRepo *repository.UserRepository,
	appealRepo *repository.AppealRepository,
	serviceRepo *repository.ServiceRepository,
) *NotificationService {
	return &NotificationService{
		repo:        repo,
		userRepo:    userRepo,
		appealRepo:  appealRepo,
		serviceRepo: serviceRepo,
	}
}

// SendAppealCreated sends notification to dispatchers when a new appeal is created
func (s *NotificationService) SendAppealCreated(ctx context.Context, appeal *models.Appeal) error {
	// Get all dispatchers and admins
	users, _, err := s.userRepo.List(ctx, 1, 1000) // Get all dispatchers/admins
	if err != nil {
		return fmt.Errorf("failed to get dispatchers: %w", err)
	}

	appealID := appeal.ID
	for _, user := range users {
		// Only send to dispatchers and admins
		if user.Role != models.RoleDispatcher && user.Role != models.RoleAdmin {
			continue
		}

		notification := &models.Notification{
			UserID:   user.ID,
			AppealID: &appealID,
			Type:     models.NotificationAppealCreated,
			Title:    "Нове звернення",
			Message:  fmt.Sprintf("Створено нове звернення: %s", appeal.Title),
		}

		if err := s.repo.Create(ctx, notification); err != nil {
			// Log error but continue with other notifications
			continue
		}
	}

	return nil
}

// SendAppealAssigned sends notification to service executors when an appeal is assigned
func (s *NotificationService) SendAppealAssigned(ctx context.Context, appeal *models.Appeal) error {
	if appeal.ServiceID == nil {
		return nil // No service assigned, no notifications needed
	}

	// Get all executors for this service
	executors, err := s.userRepo.GetExecutorsByService(ctx, *appeal.ServiceID)
	if err != nil {
		return fmt.Errorf("failed to get executors: %w", err)
	}

	appealID := appeal.ID
	for _, executor := range executors {
		notification := &models.Notification{
			UserID:   executor.ID,
			AppealID: &appealID,
			Type:     models.NotificationAppealAssigned,
			Title:    "Звернення призначено",
			Message:  fmt.Sprintf("Звернення '%s' призначено до вашої служби", appeal.Title),
		}

		if err := s.repo.Create(ctx, notification); err != nil {
			// Log error but continue with other notifications
			continue
		}
	}

	return nil
}

// SendStatusChanged sends notification to appeal creator when status changes
func (s *NotificationService) SendStatusChanged(ctx context.Context, appeal *models.Appeal, oldStatus models.AppealStatus) error {
	if appeal.Status == oldStatus {
		return nil // No status change
	}

	appealID := appeal.ID
	statusLabels := map[models.AppealStatus]string{
		models.StatusNew:        "Нове",
		models.StatusAssigned:    "Призначене",
		models.StatusInProgress:  "В роботі",
		models.StatusCompleted:   "Виконане",
		models.StatusClosed:      "Закрите",
		models.StatusRejected:   "Відхилене",
	}

	newStatusLabel := statusLabels[appeal.Status]
	if newStatusLabel == "" {
		newStatusLabel = string(appeal.Status)
	}

	notification := &models.Notification{
		UserID:   appeal.UserID,
		AppealID: &appealID,
		Type:     models.NotificationStatusChanged,
		Title:    "Статус звернення змінено",
		Message:  fmt.Sprintf("Статус звернення '%s' змінено на: %s", appeal.Title, newStatusLabel),
	}

	if appeal.Status == models.StatusCompleted || appeal.Status == models.StatusClosed {
		notification.Type = models.NotificationAppealCompleted
		notification.Title = "Звернення виконано"
		notification.Message = fmt.Sprintf("Ваше звернення '%s' виконано", appeal.Title)
	}

	return s.repo.Create(ctx, notification)
}

// SendCommentAdded sends notification when a comment is added to an appeal
func (s *NotificationService) SendCommentAdded(ctx context.Context, appealID int64, commentUserID int64, commentText string) error {
	// Get the appeal to find the creator
	appeal, err := s.appealRepo.GetByID(context.Background(), appealID)
	if err != nil {
		return fmt.Errorf("failed to get appeal: %w", err)
	}

	// Don't notify the comment author
	if appeal.UserID == commentUserID {
		return nil
	}

	// Get comment author name
	commentUser, err := s.userRepo.GetByID(ctx, commentUserID)
	if err != nil {
		return fmt.Errorf("failed to get comment user: %w", err)
	}

	appealIDPtr := &appealID
	notification := &models.Notification{
		UserID:   appeal.UserID,
		AppealID: appealIDPtr,
		Type:     models.NotificationCommentAdded,
		Title:    "Новий коментар",
		Message:  fmt.Sprintf("%s %s додав коментар до звернення '%s'", commentUser.FirstName, commentUser.LastName, appeal.Title),
	}

	return s.repo.Create(ctx, notification)
}

