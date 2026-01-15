package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// JSONB is a helper type for PostgreSQL JSONB fields
type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// User represents a user within an application
type User struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	ApplicationID uuid.UUID  `json:"application_id" db:"application_id"`
	ExternalID    string     `json:"external_id" db:"external_id"`
	Username      string     `json:"username" db:"username"`
	Email         *string    `json:"email,omitempty" db:"email"`
	AvatarURL     *string    `json:"avatar_url,omitempty" db:"avatar_url"`
	Metadata      JSONB      `json:"metadata,omitempty" db:"metadata"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}
