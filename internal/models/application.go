package models

import (
	"time"

	"github.com/google/uuid"
)

// Application represents a multi-tenant application
type Application struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	APIKey    string    `json:"api_key" db:"api_key"`
	SecretKey string    `json:"secret_key" db:"secret_key"` // Should not be exposed in responses
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

