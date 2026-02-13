package models

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	ApplicationID *uuid.UUID `json:"application_id,omitempty" db:"application_id"`
	MessageID     uuid.UUID  `json:"message_id" db:"message_id"`
	UserID        uuid.UUID  `json:"user_id" db:"user_id"`
	Filename      string     `json:"filename" db:"filename"`
	FilePath      string     `json:"file_path" db:"file_path"`
	FileSize      int64      `json:"file_size" db:"file_size"`
	MimeType      string     `json:"mime_type" db:"mime_type"`
	UploadedAt    time.Time  `json:"uploaded_at" db:"uploaded_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}
