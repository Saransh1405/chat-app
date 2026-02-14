package models

import (
	"github.com/google/uuid"
)

// models/chat.go
type ChatSession struct {
	ID        uuid.UUID     `db:"id" json:"id"`
	UserID    uuid.UUID     `db:"user_id" json:"user_id"`
	Title     string        `db:"title" json:"title"`
	CreatedAt int64         `db:"created_at" json:"created_at"`
	UpdatedAt int64         `db:"updated_at" json:"updated_at"`
	Messages  []ChatMessage `db:"messages" json:"messages"`
}

type ChatMessage struct {
	ID        uuid.UUID `db:"id" json:"id"`
	SessionID uuid.UUID `db:"session_id" json:"session_id"`
	Role      string    `db:"role" json:"role"`
	Content   string    `db:"content" json:"content"`
	Timestamp int64     `db:"timestamp" json:"timestamp"`
}

type AskRequest struct {
	Question  string     `json:"question" binding:"required"`
	SessionID *uuid.UUID `json:"sessionId,omitempty"`
}
