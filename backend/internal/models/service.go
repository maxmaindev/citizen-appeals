package models

import (
	"time"
)

type Service struct {
	ID            int64     `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	Description   string    `json:"description" db:"description"`
	ContactPerson string    `json:"contact_person" db:"contact_person"`
	ContactPhone  string    `json:"contact_phone" db:"contact_phone"`
	ContactEmail  string    `json:"contact_email" db:"contact_email"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	// Keywords are stored in MongoDB, not PostgreSQL
	Keywords string `json:"keywords,omitempty" db:"-"`
}

type CreateServiceRequest struct {
	Name          string `json:"name" validate:"required,min=3,max=200"`
	Description   string `json:"description"`
	Keywords      string `json:"keywords"` // Keywords are stored in MongoDB
	ContactPerson string `json:"contact_person" validate:"required"`
	ContactPhone  string `json:"contact_phone" validate:"required"`
	ContactEmail  string `json:"contact_email" validate:"required,email"`
}

type UpdateServiceRequest struct {
	Name          *string `json:"name" validate:"omitempty,min=3,max=200"`
	Description   *string `json:"description"`
	Keywords      *string `json:"keywords"` // Keywords are stored in MongoDB
	ContactPerson *string `json:"contact_person"`
	ContactPhone  *string `json:"contact_phone"`
	ContactEmail  *string `json:"contact_email" validate:"omitempty,email"`
	IsActive      *bool   `json:"is_active"`
}
