package models

import (
	"time"
)

type Category struct {
	ID              int64     `json:"id" db:"id"`
	Name            string    `json:"name" db:"name"`
	Description     string    `json:"description" db:"description"`
	DefaultPriority int       `json:"default_priority" db:"default_priority"`
	IsActive        bool      `json:"is_active" db:"is_active"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type CreateCategoryRequest struct {
	Name            string `json:"name" validate:"required,min=3,max=100"`
	Description     string `json:"description"`
	DefaultPriority int    `json:"default_priority" validate:"required,min=1,max=3"`
}

type UpdateCategoryRequest struct {
	Name            *string `json:"name" validate:"omitempty,min=3,max=100"`
	Description     *string `json:"description"`
	DefaultPriority *int    `json:"default_priority" validate:"omitempty,min=1,max=3"`
	IsActive        *bool   `json:"is_active"`
}
