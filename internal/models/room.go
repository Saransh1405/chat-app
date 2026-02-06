package models

import (
	"time"

	"github.com/google/uuid"
)

type RoomType string

const (
	RoomTypeGroup   RoomType = "group"
	RoomTypeDirect  RoomType = "direct"
	RoomTypeChannel RoomType = "channel"
)

type Room struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	ApplicationID uuid.UUID  `json:"application_id" db:"application_id"`
	Name          string     `json:"name" db:"name"`
	Type          RoomType   `json:"type" db:"type"`
	Description   *string    `json:"description,omitempty" db:"description"`
	CreatedBy     *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	Metadata      JSONB      `json:"metadata,omitempty" db:"metadata"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type RoomMember struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	RoomID     uuid.UUID  `json:"room_id" db:"room_id"`
	UserID     uuid.UUID  `json:"user_id" db:"user_id"`
	Username   string     `json:"username" db:"user_name"`
	Role       string     `json:"role" db:"role"`
	JoinedAt   time.Time  `json:"joined_at" db:"joined_at"`
	LastReadAt *time.Time `json:"last_read_at,omitempty" db:"last_read_at"`
}
