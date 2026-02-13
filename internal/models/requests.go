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
	Email         *string    `json:"email" binding:"required"`
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
	ID            *string `json:"id,omitempty" form:"id"`
	ApplicationID *string `json:"application_id,omitempty" form:"application_id"`
}

type UserDeleteRequest struct {
	ID            uuid.UUID  `json:"id" binding:"required"`
	ApplicationID *uuid.UUID `json:"application_id,omitempty"`
}

// Message Request Models
type MessageCreateRequest struct {
	ApplicationID *uuid.UUID                `json:"application_id,omitempty"`
	RoomID        uuid.UUID                 `json:"room_id" binding:"required"`
	Content       string                    `json:"content,omitempty"`
	Type          string                    `json:"type,omitempty"` // "TEXT" or "MEDIA"
	MessageType   MessageType               `json:"message_type,omitempty"`
	ReplyTo       *uuid.UUID                `json:"reply_to,omitempty"`
	Metadata      JSONB                     `json:"metadata,omitempty"`
	FileRequest   *FileMessageCreateRequest `json:"file_request,omitempty"`
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

// Auth Request Models
type RegisterRequest struct {
	ApplicationID *uuid.UUID `json:"application_id,omitempty"`
	ExternalID    string     `json:"external_id,omitempty"`
	Username      string     `json:"username" binding:"required"`
	Email         *string    `json:"email" binding:"required"`
	AvatarURL     *string    `json:"avatar_url,omitempty"`
	Metadata      JSONB      `json:"metadata,omitempty"`
}

type LoginRequest struct {
	Email string `json:"email" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Reaction Request Models
type ReactionCreateRequest struct {
	ApplicationID *uuid.UUID `json:"application_id,omitempty"`
	MessageID     uuid.UUID  `json:"message_id" binding:"required"`
	Reaction      string     `json:"reaction" binding:"required"`
}

type ReactionDeleteRequest struct {
	ApplicationID *uuid.UUID `json:"application_id,omitempty"`
	MessageID     uuid.UUID  `json:"message_id" binding:"required"`
	Reaction      string     `json:"reaction" binding:"required"`
}

type ReactionListRequest struct {
	ApplicationID *uuid.UUID `json:"application_id,omitempty"`
	MessageID     uuid.UUID  `json:"message_id" binding:"required"`
}

type FileUploadRequest struct {
	Folder         string                 `form:"folder,omitempty"`
	FileName       string                 `form:"file_name,omitempty"`
	Tags           []string               `form:"tags,omitempty"`
	CustomMetadata map[string]interface{} `form:"custom_metadata,omitempty"`
}

type FileMessageCreateRequest struct {
	Filename string `json:"filename" binding:"omitempty"`
	FilePath string `json:"file_path" binding:"omitempty"`
	FileSize int64  `json:"file_size" binding:"omitempty"`
	MimeType string `json:"mime_type" binding:"omitempty"`
}
