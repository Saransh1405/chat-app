package answer

import (
	"chat-app/internal/database"
	"chat-app/internal/models"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	openai "github.com/sashabaranov/go-openai"
)

func SearchRelevantChunks(ctx context.Context, question string, userId uuid.UUID, limit int, db *database.DB) ([]models.Chunk, error) {
	// 1. Generate embedding for the question
	log.Printf("Started generating embedding for the question")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	client := openai.NewClient(apiKey)

	resp, err := client.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			Input: []string{question},
			Model: openai.SmallEmbedding3,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding for the question: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned for the question")
	}

	log.Printf("Generated embedding for the question")

	// 2. Search similar chunks in PostgreSQL
	var chunks []models.Chunk
	query := `SELECT * FROM chunks c
		 INNER JOIN documents d ON c.document_id = d.id
		 WHERE d.user_id = ?
		 ORDER BY c.embedding <=> ?
		 LIMIT ?`

	err = db.QueryRow(query, userId, resp.Data[0].Embedding, limit).Scan(&chunks)
	if err != nil {
		return nil, fmt.Errorf("failed to search relevant chunks: %w", err)
	}
	return chunks, nil
}
