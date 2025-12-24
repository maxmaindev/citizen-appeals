package models

import (
	"time"
)

// UserService represents the relationship between a user (executor) and a service
type UserService struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	ServiceID int64     `json:"service_id" db:"service_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// Joined fields
	User    *User    `json:"user,omitempty" db:"-"`
	Service *Service `json:"service,omitempty" db:"-"`
}

// UserServiceAssignment represents a request to assign users to a service
type UserServiceAssignment struct {
	ServiceID int64   `json:"service_id" validate:"required"`
	UserIDs   []int64 `json:"user_ids" validate:"required"`
}

// ServiceWithUsers represents a service with its assigned users
type ServiceWithUsers struct {
	Service *Service `json:"service"`
	Users   []*User  `json:"users"`
}
