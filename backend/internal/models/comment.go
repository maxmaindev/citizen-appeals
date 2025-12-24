package models

import (
	"time"
)

type Comment struct {
	ID         int64     `json:"id" db:"id"`
	AppealID   int64     `json:"appeal_id" db:"appeal_id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	Text       string    `json:"text" db:"text"`
	IsInternal bool      `json:"is_internal" db:"is_internal"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`

	// Joined fields
	User   *User   `json:"user,omitempty" db:"-"`
	Photos []Photo `json:"photos,omitempty" db:"-"`
}

type CreateCommentRequest struct {
	Text       string  `json:"text" validate:"required,min=1"`
	IsInternal bool    `json:"is_internal"`
	PhotoIDs   []int64 `json:"photo_ids"`
}
