package models

import (
	"time"
)

// CategoryService represents the relationship between a category and a service
type CategoryService struct {
	ID         int64     `json:"id" db:"id"`
	CategoryID int64     `json:"category_id" db:"category_id"`
	ServiceID  int64     `json:"service_id" db:"service_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`

	// Joined fields
	Category *Category `json:"category,omitempty" db:"-"`
	Service  *Service  `json:"service,omitempty" db:"-"`
}

// CategoryServiceAssignment represents a request to assign services to a category
type CategoryServiceAssignment struct {
	CategoryID int64   `json:"category_id" validate:"required"`
	ServiceIDs []int64 `json:"service_ids" validate:"required"` // Allow empty array to remove all services
}

// CategoryWithServices represents a category with its assigned services
type CategoryWithServices struct {
	Category *Category        `json:"category"`
	Services []*CategoryService `json:"services"`
}

