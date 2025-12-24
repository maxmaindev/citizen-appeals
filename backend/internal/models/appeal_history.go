package models

import "time"

type AppealHistory struct {
	ID        int64        `json:"id" db:"id"`
	AppealID  int64        `json:"appeal_id" db:"appeal_id"`
	UserID    int64        `json:"user_id" db:"user_id"`
	OldStatus *AppealStatus `json:"old_status" db:"old_status"`
	NewStatus AppealStatus `json:"new_status" db:"new_status"`
	Action    string       `json:"action" db:"action"`
	Comment   *string      `json:"comment" db:"comment"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	
	// Joined fields
	User *User `json:"user,omitempty" db:"-"`
}

