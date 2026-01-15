package models

import (
	"time"

	"github.com/google/uuid"
)

// TypingIndicator represents a typing indicator
type TypingIndicator struct {
	ID        uuid.UUID `json:"id" db:"id"`
	RoomID    uuid.UUID `json:"room_id" db:"room_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

