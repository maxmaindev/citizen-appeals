package models

import (
	"time"
)

type AppealStatus string

const (
	StatusNew        AppealStatus = "new"
	StatusAssigned   AppealStatus = "assigned"
	StatusInProgress AppealStatus = "in_progress"
	StatusCompleted  AppealStatus = "completed"
	StatusClosed     AppealStatus = "closed"
	StatusRejected   AppealStatus = "rejected"
)

type Appeal struct {
	ID          int64        `json:"id" db:"id"`
	UserID      int64        `json:"user_id" db:"user_id"`
	CategoryID  *int64       `json:"category_id" db:"category_id"`
	ServiceID   *int64       `json:"service_id" db:"service_id"`
	Status      AppealStatus `json:"status" db:"status"`
	Title       string       `json:"title" db:"title"`
	Description string       `json:"description" db:"description"`
	Address     string       `json:"address" db:"address"`
	Latitude    float64      `json:"latitude" db:"latitude"`
	Longitude   float64      `json:"longitude" db:"longitude"`
	Priority    int          `json:"priority" db:"priority"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
	ClosedAt    *time.Time   `json:"closed_at" db:"closed_at"`

	// Joined fields
	User     *User     `json:"user,omitempty" db:"-"`
	Category *Category `json:"category,omitempty" db:"-"`
	Service  *Service  `json:"service,omitempty" db:"-"`
	Photos   []Photo   `json:"photos,omitempty" db:"-"`
	Comments []Comment `json:"comments,omitempty" db:"-"`
}

type CreateAppealRequest struct {
	Title       string  `json:"title" validate:"required,min=5,max=200"`
	Description string  `json:"description" validate:"required,min=10"`
	CategoryID  int64   `json:"category_id" validate:"required"`
	Address     string  `json:"address" validate:"required"`
	Latitude    float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude   float64 `json:"longitude" validate:"required,min=-180,max=180"`
	Priority    *int    `json:"priority" validate:"omitempty,min=1,max=3"`
	PhotoIDs    []int64 `json:"photo_ids"`
}

type UpdateAppealRequest struct {
	Title       *string  `json:"title" validate:"omitempty,min=5,max=200"`
	Description *string  `json:"description" validate:"omitempty,min=10"`
	CategoryID  *int64   `json:"category_id"`
	Address     *string  `json:"address"`
	Latitude    *float64 `json:"latitude" validate:"omitempty,min=-90,max=90"`
	Longitude   *float64 `json:"longitude" validate:"omitempty,min=-180,max=180"`
}

type AssignAppealRequest struct {
	ServiceID int64 `json:"service_id" validate:"required"`
	Priority  *int  `json:"priority" validate:"omitempty,min=1,max=3"`
}

type UpdateStatusRequest struct {
	Status  AppealStatus `json:"status" validate:"required"`
	Comment *string      `json:"comment"`
}

type UpdatePriorityRequest struct {
	Priority int `json:"priority" validate:"required,min=1,max=3"`
}

type AppealFilters struct {
	Status     *AppealStatus `json:"status"`
	CategoryID *int64        `json:"category_id"`
	ServiceID  *int64        `json:"service_id"`
	UserID     *int64        `json:"user_id"`
	FromDate   *time.Time    `json:"from_date"`
	ToDate     *time.Time    `json:"to_date"`
	Search     *string       `json:"search"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	SortBy     string        `json:"sort_by"`
	SortOrder  string        `json:"sort_order"`
}
