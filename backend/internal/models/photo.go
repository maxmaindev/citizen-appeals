package models

import (
	"time"
)

type Photo struct {
	ID            int64     `json:"id" db:"id"`
	AppealID      *int64    `json:"appeal_id" db:"appeal_id"`
	CommentID     *int64    `json:"comment_id" db:"comment_id"`
	FilePath      string    `json:"file_path" db:"file_path"`
	FileName      string    `json:"file_name" db:"file_name"`
	FileSize      int64     `json:"file_size" db:"file_size"`
	MimeType      string    `json:"mime_type" db:"mime_type"`
	IsResultPhoto bool      `json:"is_result_photo" db:"is_result_photo"`
	UploadedAt    time.Time `json:"uploaded_at" db:"uploaded_at"`
}

type UploadPhotoResponse struct {
	ID       int64  `json:"id"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	URL      string `json:"url"`
}
