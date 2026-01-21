package models

import (
	"github.com/google/uuid"
)

type ApplicationCreateRequest struct {
	Name string `json:"name" binding:"required"`
}

type ApplicationUpdateRequest struct {
	ID   uuid.UUID `json:"id" binding:"required"`
	Name string    `json:"name"`
}

type ApplicationGetRequest struct {
	ID uuid.UUID `json:"id" binding:"required"`
}

type ApplicationDeleteRequest struct {
	ID uuid.UUID `json:"id" binding:"required"`
}

// User Request Models
type UserCreateRequest struct {
	ApplicationID *uuid.UUID `json:"application_id,omitempty"`
	ExternalID    string     `json:"external_id,omitempty"`
	Username      string     `json:"username" binding:"required"`
	Email         *string    `json:"email,omitempty"`
	AvatarURL     *string    `json:"avatar_url,omitempty"`
	Metadata      JSONB      `json:"metadata,omitempty"`
}

type UserUpdateRequest struct {
	ID            uuid.UUID  `json:"id" binding:"required"`
	ApplicationID *uuid.UUID `json:"application_id,omitempty"`
	Username      string     `json:"username,omitempty"`
	Email         *string    `json:"email,omitempty"`
	AvatarURL     *string    `json:"avatar_url,omitempty"`
	Metadata      JSONB      `json:"metadata,omitempty"`
}

type UserGetRequest struct {
	ID            *uuid.UUID `json:"id,omitempty"`
	ApplicationID *uuid.UUID `json:"application_id,omitempty"`
}

type UserDeleteRequest struct {
	ID            uuid.UUID  `json:"id" binding:"required"`
	ApplicationID *uuid.UUID `json:"application_id,omitempty"`
}

// Message Request Models
type MessageCreateRequest struct {
	ApplicationID *uuid.UUID  `json:"application_id,omitempty"`
	RoomID        uuid.UUID   `json:"room_id" binding:"required"`
	Content       string      `json:"content" binding:"required"`
	MessageType   MessageType `json:"message_type,omitempty"`
	ReplyTo       *uuid.UUID  `json:"reply_to,omitempty"`
	Metadata      JSONB       `json:"metadata,omitempty"`
}

type MessageUpdateRequest struct {
	ApplicationID *uuid.UUID  `json:"application_id,omitempty"`
	ID            uuid.UUID   `json:"id" binding:"required"`
	RoomID        uuid.UUID   `json:"room_id" binding:"required"`
	Content       string      `json:"content,omitempty"`
	MessageType   MessageType `json:"message_type,omitempty"`
	Metadata      JSONB       `json:"metadata,omitempty"`
}

type MessageGetRequest struct {
	ApplicationID *uuid.UUID `json:"application_id,omitempty"`
	ID            uuid.UUID  `json:"id" binding:"required"`
	RoomID        uuid.UUID  `json:"room_id" binding:"required"`
}

type MessageListRequest struct {
	ApplicationID *uuid.UUID `json:"application_id,omitempty"`
	RoomID        uuid.UUID  `json:"room_id" binding:"required"`
	Limit         int        `json:"limit,omitempty"`
	Offset        int        `json:"offset,omitempty"`
}

type MessageDeleteRequest struct {
	ApplicationID *uuid.UUID `json:"application_id,omitempty"`
	ID            uuid.UUID  `json:"id" binding:"required"`
	RoomID        uuid.UUID  `json:"room_id" binding:"required"`
}

type TypingCreateRequest struct {
	ApplicationID *uuid.UUID `json:"application_id,omitempty"`
	RoomID        uuid.UUID  `json:"room_id" binding:"required"`
	UserID        uuid.UUID  `json:"user_id" binding:"required"`
}

type RoomCreateRequest struct {
	ApplicationID *uuid.UUID `json:"application_id,omitempty"`
	Name          string     `json:"name" binding:"required"`
	Type          RoomType   `json:"type,omitempty"`
	Description   *string    `json:"description,omitempty"`
	CreatedBy     *uuid.UUID `json:"created_by,omitempty"`
	Metadata      JSONB      `json:"metadata,omitempty"`
}

type RoomUpdateRequest struct {
	ApplicationID *uuid.UUID `json:"application_id,omitempty"`
	ID            uuid.UUID  `json:"id" binding:"required"`
	Name          string     `json:"name,omitempty"`
	Description   *string    `json:"description,omitempty"`
	Metadata      JSONB      `json:"metadata,omitempty"`
}

type RoomDeleteRequest struct {
	ApplicationID *uuid.UUID `json:"application_id,omitempty"`
	ID            uuid.UUID  `json:"id" binding:"required"`
}

type RoomAddMemberRequest struct {
	ApplicationID *uuid.UUID `json:"application_id,omitempty"`
	RoomID        uuid.UUID  `json:"room_id" binding:"required"`
	UserID        uuid.UUID  `json:"user_id" binding:"required"`
	Role          string     `json:"role,omitempty"`
}

type RoomRemoveMemberRequest struct {
	ApplicationID *uuid.UUID `json:"application_id,omitempty"`
	RoomID        uuid.UUID  `json:"room_id" binding:"required"`
	UserID        uuid.UUID  `json:"user_id" binding:"required"`
}
