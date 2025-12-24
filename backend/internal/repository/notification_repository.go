package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"citizen-appeals/internal/models"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
)

type NotificationRepository struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create creates a new notification
func (r *NotificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	query := `
		INSERT INTO notifications (user_id, appeal_id, type, title, message)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, sent_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		notification.UserID,
		notification.AppealID,
		notification.Type,
		notification.Title,
		notification.Message,
	).Scan(&notification.ID, &notification.SentAt)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

// GetByID retrieves a notification by ID
func (r *NotificationRepository) GetByID(ctx context.Context, id int64) (*models.Notification, error) {
	query := `
		SELECT id, user_id, appeal_id, type, title, message, is_read, sent_at
		FROM notifications
		WHERE id = $1
	`

	var notification models.Notification
	err := r.db.QueryRow(ctx, query, id).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.AppealID,
		&notification.Type,
		&notification.Title,
		&notification.Message,
		&notification.IsRead,
		&notification.SentAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotificationNotFound
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	return &notification, nil
}

// GetByUserID retrieves all notifications for a user
func (r *NotificationRepository) GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*models.Notification, error) {
	query := `
		SELECT id, user_id, appeal_id, type, title, message, is_read, sent_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY sent_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}
	defer rows.Close()

	notifications := make([]*models.Notification, 0)
	for rows.Next() {
		var notification models.Notification
		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.AppealID,
			&notification.Type,
			&notification.Title,
			&notification.Message,
			&notification.IsRead,
			&notification.SentAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, &notification)
	}

	return notifications, nil
}

// GetUnreadCount returns the count of unread notifications for a user
func (r *NotificationRepository) GetUnreadCount(ctx context.Context, userID int64) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = $1 AND is_read = false
	`

	var count int64
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}

// MarkAsRead marks a notification as read
func (r *NotificationRepository) MarkAsRead(ctx context.Context, id int64, userID int64) error {
	query := `
		UPDATE notifications
		SET is_read = true
		WHERE id = $1 AND user_id = $2
	`

	result, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotificationNotFound
	}

	return nil
}

// MarkAllAsRead marks all notifications for a user as read
func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID int64) error {
	query := `
		UPDATE notifications
		SET is_read = true
		WHERE user_id = $1 AND is_read = false
	`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

// Delete deletes a notification
func (r *NotificationRepository) Delete(ctx context.Context, id int64, userID int64) error {
	query := `
		DELETE FROM notifications
		WHERE id = $1 AND user_id = $2
	`

	result, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotificationNotFound
	}

	return nil
}

