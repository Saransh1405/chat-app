package answer

import (
	"chat-app/internal/database"
	"chat-app/internal/models"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"github.com/sashabaranov/go-openai"
)

func SearchRelevantChunks(ctx context.Context, question string, userId uuid.UUID, limit int, db *database.DB) ([]models.Chunk, error) {
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

	emd := resp.Data[0].Embedding

	// emd, err := GetDeepSeekEmbedding(question)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get embedding for the question: %w", err)
	// }

	// log.Printf("Generated embedding for the question")

	// emdFloat32 := make([]float32, len(emd))
	// for i, v := range emd {
	// 	emdFloat32[i] = float32(v)
	// }

	embedding := pgvector.NewVector(emd)

	var chunks []models.Chunk
	query := `SELECT c.id, c.document_id, c.content, c.chunk_index, c.embedding, d.file_name 
	     FROM chunks c
		 INNER JOIN documents d ON c.document_id = d.id
		 WHERE d.user_id = $1
		 ORDER BY c.embedding <=> $2
		 LIMIT $3`

	rows, err := db.Query(query, userId, embedding, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search relevant chunks: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var chunk models.Chunk
		if err := rows.Scan(&chunk.ID, &chunk.DocumentID, &chunk.Content, &chunk.ChunkIndex, &chunk.Embedding, &chunk.Source); err != nil {
			log.Printf("Error scanning chunk row: %v", err)
			continue
		}
		chunks = append(chunks, chunk)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating chunks: %w", err)
	}

	if chunks == nil {
		chunks = []models.Chunk{}
	}

	return chunks, nil
}
