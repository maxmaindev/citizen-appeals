package models

import (
	"time"
)

type NotificationType string

const (
	NotificationAppealCreated   NotificationType = "appeal_created"
	NotificationAppealAssigned  NotificationType = "appeal_assigned"
	NotificationStatusChanged   NotificationType = "status_changed"
	NotificationCommentAdded    NotificationType = "comment_added"
	NotificationAppealCompleted NotificationType = "appeal_completed"
)

type Notification struct {
	ID       int64            `json:"id" db:"id"`
	UserID   int64            `json:"user_id" db:"user_id"`
	AppealID *int64           `json:"appeal_id" db:"appeal_id"`
	Type     NotificationType `json:"type" db:"type"`
	Title    string           `json:"title" db:"title"`
	Message  string           `json:"message" db:"message"`
	IsRead   bool             `json:"is_read" db:"is_read"`
	SentAt   time.Time        `json:"sent_at" db:"sent_at"`

	// Joined fields
	Appeal *Appeal `json:"appeal,omitempty" db:"-"`
}
