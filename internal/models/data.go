package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

type Document struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserId    uuid.UUID `db:"user_id" json:"user_id"`
	FileName  string    `db:"file_name" json:"file_name"`
	FileType  string    `db:"file_type" json:"file_type"`
	FileSize  int64     `db:"file_size" json:"file_size"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	Chunks    []Chunk   `db:"chunks" json:"chunks"`
}

type Chunk struct {
	ID         uuid.UUID       `db:"id" json:"id"`
	DocumentID uuid.UUID       `db:"document_id" json:"document_id"`
	Content    string          `db:"content" json:"content"`
	ChunkIndex int             `db:"chunk_index" json:"chunk_index"`
	Embedding  pgvector.Vector `db:"embedding" json:"embedding"`
	Source     string          `db:"-" json:"source,omitempty"`
}

type Uploadfile struct {
	Title *string `json:"title"`
	Type  *string `json:"type"`
}
