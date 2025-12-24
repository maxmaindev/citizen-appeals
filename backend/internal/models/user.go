package models

import (
	"time"
)

type UserRole string

const (
	RoleCitizen    UserRole = "citizen"
	RoleDispatcher UserRole = "dispatcher"
	RoleExecutor   UserRole = "executor"
	RoleAdmin      UserRole = "admin"
)

type User struct {
	ID           int64     `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	FirstName    string    `json:"first_name" db:"first_name"`
	LastName     string    `json:"last_name" db:"last_name"`
	Phone        string    `json:"phone" db:"phone"`
	Role         UserRole  `json:"role" db:"role"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required,min=2"`
	LastName  string `json:"last_name" validate:"required,min=2"`
	Phone     string `json:"phone" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

type UpdateUserRequest struct {
	FirstName *string  `json:"first_name" validate:"omitempty,min=2"`
	LastName  *string  `json:"last_name" validate:"omitempty,min=2"`
	Phone     *string  `json:"phone"`
	Role      *UserRole `json:"role" validate:"omitempty,oneof=citizen dispatcher executor admin"`
	IsActive  *bool    `json:"is_active"`
}

type UpdateProfileRequest struct {
	FirstName *string `json:"first_name" validate:"omitempty,min=2"`
	LastName  *string `json:"last_name" validate:"omitempty,min=2"`
	Phone     *string `json:"phone"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}
