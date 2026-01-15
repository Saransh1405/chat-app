package models

import (
	"time"

	"github.com/google/uuid"
)

// MessageType represents the type of message
type MessageType string

const (
	MessageTypeText   MessageType = "text"
	MessageTypeImage  MessageType = "image"
	MessageTypeFile   MessageType = "file"
	MessageTypeSystem MessageType = "system"
)

// Message represents a chat message
type Message struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	RoomID      uuid.UUID   `json:"room_id" db:"room_id"`
	UserID      uuid.UUID   `json:"user_id" db:"user_id"`
	Content     string      `json:"content" db:"content"`
	MessageType MessageType `json:"message_type" db:"message_type"`
	ReplyTo     *uuid.UUID  `json:"reply_to,omitempty" db:"reply_to"`
	EditedAt    *time.Time  `json:"edited_at,omitempty" db:"edited_at"`
	DeletedAt   *time.Time  `json:"deleted_at,omitempty" db:"deleted_at"`
	Metadata    JSONB       `json:"metadata,omitempty" db:"metadata"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

// MessageReaction represents a reaction on a message
type MessageReaction struct {
	ID        uuid.UUID `json:"id" db:"id"`
	MessageID uuid.UUID `json:"message_id" db:"message_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Reaction  string    `json:"reaction" db:"reaction"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// MessageRead represents a read receipt for a message
type MessageRead struct {
	ID        uuid.UUID `json:"id" db:"id"`
	MessageID uuid.UUID `json:"message_id" db:"message_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	ReadAt    time.Time `json:"read_at" db:"read_at"`
}

