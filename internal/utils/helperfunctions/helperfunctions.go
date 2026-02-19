package helperfunctions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"chat-app/internal/database"
	"chat-app/internal/models"
	"chat-app/vault-ai/library/tika"
	"crypto/rand"
	"encoding/base64"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"github.com/sashabaranov/go-openai"
)

func CreateUser(ctx *gin.Context, db *database.DB, user *models.User, appID string) error {
	var appIDUUID *uuid.UUID
	var err error

	if appID != "" {
		parsedUUID, err := uuid.Parse(appID)
		if err != nil {
			return errors.New("invalid application ID format")
		}
		appIDUUID = &parsedUUID
	} else {
		appIDUUID = nil
	}

	query := `INSERT INTO users (application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	          RETURNING id, created_at, updated_at`

	now := time.Now()
	var metadataJSON []byte
	if user.Metadata != nil {
		metadataJSON, err = json.Marshal(user.Metadata)
		if err != nil {
			return errors.New("failed to marshal metadata")
		}
	}

	err = db.QueryRow(query,
		appIDUUID,
		user.ExternalID,
		user.Username,
		user.Email,
		user.AvatarURL,
		metadataJSON,
		now,
		now,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return err
	}

	if appIDUUID != nil {
		user.ApplicationID = *appIDUUID
	} else {
		user.ApplicationID = uuid.Nil
	}

	return nil
}

func CheckIfUserExistsWithEmail(ctx *gin.Context, db *database.DB, user *models.User, appID string) error {
	var query string
	var args []interface{}

	if appID != "" {
		appIDUUID, err := uuid.Parse(appID)
		if err != nil {
			return errors.New("invalid application ID format")
		}
		query = `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at
		         FROM users
		         WHERE email = $1 AND application_id = $2 AND deleted_at IS NULL`
		args = []interface{}{user.Email, appIDUUID}
	} else {
		query = `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at
		         FROM users
		         WHERE email = $1 AND application_id IS NULL AND deleted_at IS NULL`
		args = []interface{}{user.Email}
	}

	var metadataJSON []byte
	var appIDPtr *uuid.UUID
	err := db.QueryRow(query, args...).Scan(
		&user.ID,
		&appIDPtr,
		&user.ExternalID,
		&user.Username,
		&user.Email,
		&user.AvatarURL,
		&metadataJSON,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return errors.New("user not found")
	}
	if err != nil {
		return err
	}

	if appIDPtr != nil {
		user.ApplicationID = *appIDPtr
	} else {
		user.ApplicationID = uuid.Nil
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &user.Metadata); err != nil {
		}
	}

	return nil
}

func CheckIfUserExistsWithUsername(ctx *gin.Context, db *database.DB, user *models.User, appID string) error {
	var query string
	var args []interface{}

	if appID != "" {
		appIDUUID, err := uuid.Parse(appID)
		if err != nil {
			return errors.New("invalid application ID format")
		}
		query = `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at
		         FROM users
		         WHERE username = $1 AND application_id = $2 AND deleted_at IS NULL`
		args = []interface{}{user.Username, appIDUUID}
	} else {
		query = `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at
		         FROM users
		         WHERE username = $1 AND application_id IS NULL AND deleted_at IS NULL`
		args = []interface{}{user.Username}
	}

	var metadataJSON []byte
	var appIDPtr *uuid.UUID
	err := db.QueryRow(query, args...).Scan(
		&user.ID,
		&appIDPtr,
		&user.ExternalID,
		&user.Username,
		&user.Email,
		&user.AvatarURL,
		&metadataJSON,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return errors.New("user not found")
	}
	if err != nil {
		return err
	}

	if appIDPtr != nil {
		user.ApplicationID = *appIDPtr
	} else {
		user.ApplicationID = uuid.Nil
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &user.Metadata); err != nil {
		}
	}

	return nil
}

func GetUserByEmail(ctx *gin.Context, db *database.DB, email string) (*models.User, error) {
	query := `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at
	          FROM users
	          WHERE email = $1 AND deleted_at IS NULL`

	user := &models.User{}
	var metadataJSON []byte

	err := db.QueryRow(query, email).Scan(
		&user.ID,
		&user.ApplicationID,
		&user.ExternalID,
		&user.Username,
		&user.Email,
		&user.AvatarURL,
		&metadataJSON,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &user.Metadata); err != nil {
		}
	}

	return user, nil
}

func ValidateUserIsMemberOfRoom(db *database.DB, userID string, roomID string) (bool, error) {
	userIDUUID, err := uuid.Parse(userID)
	if err != nil {
		return false, errors.New("invalid user ID format")
	}

	roomIDUUID, err := uuid.Parse(roomID)
	if err != nil {
		return false, errors.New("invalid room ID format")
	}

	query := `SELECT COUNT(*) FROM room_members
	          WHERE user_id = $1 AND room_id = $2`

	var count int
	err = db.QueryRow(query, userIDUUID, roomIDUUID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func GenerateAPIKey(ctx *gin.Context) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func GenerateSecretKey(ctx *gin.Context) (string, error) {
	bytes := make([]byte, 64)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func SaveMessageToDB(ctx *gin.Context, db *database.DB, message *models.Message) (*models.Message, error) {
	query := `INSERT INTO messages (room_id, user_id, content, message_type, reply_to, metadata, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	          RETURNING id, created_at, updated_at`

	now := time.Now()
	var metadataJSON []byte
	var err error

	if message.Metadata != nil {
		metadataJSON, err = json.Marshal(message.Metadata)
		if err != nil {
			return nil, errors.New("failed to marshal metadata")
		}
	}

	err = db.QueryRow(query,
		message.RoomID,
		message.UserID,
		message.Content,
		message.MessageType,
		message.ReplyTo,
		metadataJSON,
		now,
		now,
	).Scan(&message.ID, &message.CreatedAt, &message.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return message, nil
}

func CheckIfUserExistsWithId(ctx *gin.Context, db *database.DB, userId string) bool {
	userIDUUID, err := uuid.Parse(userId)
	if err != nil {
		return false
	}

	query := `SELECT COUNT(*) FROM users WHERE id = $1 AND deleted_at IS NULL`

	var count int
	err = db.QueryRow(query, userIDUUID).Scan(&count)
	if err != nil {
		return false
	}

	return count > 0
}

func CheckIfApplicationExistsWithId(ctx *gin.Context, db *database.DB, applicationId string) (bool, *models.Application) {
	appIDUUID, err := uuid.Parse(applicationId)
	if err != nil {
		return false, nil
	}

	query := `SELECT id, name, api_key, secret_key, created_at, updated_at, deleted_at FROM applications WHERE id = $1 AND deleted_at IS NULL`

	app := &models.Application{}
	err = db.QueryRow(query, appIDUUID).Scan(
		&app.ID,
		&app.Name,
		&app.APIKey,
		&app.SecretKey,
		&app.CreatedAt,
		&app.UpdatedAt,
		&app.DeletedAt,
	)
	if err != nil {
		return false, nil
	}

	return true, app
}

func SaveTypingIndicatorToDB(ctx *gin.Context, db *database.DB, typing *models.TypingIndicator) (*models.TypingIndicator, error) {
	query := `INSERT INTO typing_indicators (room_id, user_id, expires_at, created_at)
	          VALUES ($1, $2, $3, $4)
	          RETURNING id, room_id, user_id, created_at, expires_at`

	now := time.Now()
	var err error

	err = db.QueryRow(query,
		typing.RoomID,
		typing.UserID,
		typing.ExpiresAt,
		now,
	).Scan(&typing.ID, &typing.RoomID, &typing.UserID, &typing.CreatedAt, &typing.ExpiresAt)

	if err != nil {
		return nil, err
	}

	return typing, nil
}

func ExpireTypingIndicator(ctx *gin.Context, db *database.DB, typing *models.TypingIndicator) error {
	return nil
}

func GetRoomById(ctx *gin.Context, db *database.DB, roomID string) (*models.Room, error) {
	roomIDUUID, err := uuid.Parse(roomID)
	if err != nil {
		return nil, errors.New("invalid room ID format")
	}

	query := `SELECT id, name, description, created_at, updated_at, deleted_at FROM rooms WHERE id = $1 AND deleted_at IS NULL`

	room := &models.Room{}
	err = db.QueryRow(query, roomIDUUID).Scan(
		&room.ID,
		&room.Name,
		&room.Description,
		&room.CreatedAt,
		&room.UpdatedAt,
		&room.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	return room, nil
}

func GetUserById(ctx *gin.Context, db *database.DB, userID string) (*models.User, error) {
	userIDUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	query := `SELECT id, application_id, external_id, username, email, avatar_url, created_at, updated_at, deleted_at FROM users WHERE id = $1 AND deleted_at IS NULL`

	user := &models.User{}
	err = db.QueryRow(query, userIDUUID).Scan(
		&user.ID,
		&user.ApplicationID,
		&user.ExternalID,
		&user.Username,
		&user.Email,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func GetUserRoomsByUserID(ctx *gin.Context, db *database.DB, userID string) ([]*models.Room, error) {
	userIDUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	roomIDsQuery := `SELECT room_id FROM room_members WHERE user_id = $1`
	rows, err := db.Query(roomIDsQuery, userIDUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roomIDs []uuid.UUID
	for rows.Next() {
		var roomID uuid.UUID
		if err := rows.Scan(&roomID); err != nil {
			return nil, err
		}
		roomIDs = append(roomIDs, roomID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(roomIDs) == 0 {
		return []*models.Room{}, nil
	}

	rooms := []*models.Room{}
	for _, roomID := range roomIDs {
		room := &models.Room{}
		roomQuery := `SELECT id, name, description, created_at, updated_at, deleted_at FROM rooms WHERE id = $1 AND deleted_at IS NULL`
		err := db.QueryRow(roomQuery, roomID).Scan(
			&room.ID,
			&room.Name,
			&room.Description,
			&room.CreatedAt,
			&room.UpdatedAt,
			&room.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}

func SaveFileToDB(ctx *gin.Context, db *database.DB, file *models.File) (*models.File, error) {
	query := `INSERT INTO files (application_id, message_id, user_id, filename, file_path, file_size, mime_type, uploaded_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	          RETURNING id, uploaded_at`

	now := time.Now()

	err := db.QueryRow(query,
		file.ApplicationID,
		file.MessageID,
		file.UserID,
		file.Filename,
		file.FilePath,
		file.FileSize,
		file.MimeType,
		now,
	).Scan(&file.ID, &file.UploadedAt)

	if err != nil {
		return nil, err
	}

	return file, nil
}

func ProcessDocument(documentID uuid.UUID, userId uuid.UUID, fileType, fileTitle string, fileSize int64, chunks []models.Chunk, db *database.DB) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Goroutine panicked while processing document %s: %v", documentID.String(), r)
		}
	}()

	_, err := GenerateEmbeddingOfChunks(&chunks)
	if err != nil {
		log.Printf("Error generating embeddings for document %s: %v", documentID.String(), err)
		return
	}

	document := models.Document{
		ID:        documentID,
		UserId:    userId,
		FileType:  fileType,
		FileName:  fileTitle,
		FileSize:  fileSize,
		CreatedAt: time.Now(),
		Chunks:    chunks,
	}

	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction for document %s: %v", documentID.String(), err)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Transaction panicked for document %s: %v", documentID.String(), r)
		}
	}()

	var count int
	checkQuery := "select count(*) from documents where id = $1"
	err = tx.QueryRow(checkQuery, documentID).Scan(&count)
	if err == nil && count > 0 {
		tx.Rollback()
		log.Printf("Document %s already exists, skipping upload", documentID.String())
		return
	}

	insertDocQuery := "insert into documents (id, user_id, file_type, file_name, file_size, created_at) values ($1, $2, $3, $4, $5, $6)"
	_, err = tx.Exec(insertDocQuery, document.ID, document.UserId, document.FileType, document.FileName, document.FileSize, document.CreatedAt)
	if err != nil {
		tx.Rollback()
		log.Printf("Error inserting document %s: %v", documentID.String(), err)
		return
	}

	insertChunkQuery := "insert into chunks (id, document_id, content, chunk_index, embedding) values ($1, $2, $3, $4, $5)"
	for _, chunk := range document.Chunks {
		_, err = tx.Exec(insertChunkQuery, chunk.ID, chunk.DocumentID, chunk.Content, chunk.ChunkIndex, chunk.Embedding)
		if err != nil {
			tx.Rollback()
			log.Printf("Error inserting chunk for document %s: %v", documentID.String(), err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction for document %s: %v", documentID.String(), err)
		return
	}

	log.Printf("Successfully processed document and %d chunks: %s", len(document.Chunks), documentID.String())
}

func ConvertToChunks(fileHeader *multipart.FileHeader, fileType string, documentID uuid.UUID) ([]models.Chunk, error) {
	log.Printf("Started extracting text from file")

	fullText, err := ExtractTextFromFile(fileHeader, fileType)
	if err != nil {
		return nil, fmt.Errorf("text extraction failed: %w", err)
	}

	if strings.TrimSpace(fullText) == "" {
		return nil, fmt.Errorf("no text content found in file")
	}

	log.Printf("Extracted %d characters from file", len(fullText))

	textChunks := SplitIntoTextChunks(fullText, 8000)

	var chunks []models.Chunk
	for i, textChunk := range textChunks {
		chunks = append(chunks, models.Chunk{
			ID:         uuid.New(),
			DocumentID: documentID,
			Content:    textChunk,
			ChunkIndex: i + 1,
		})
	}

	log.Printf("Created %d chunks from file", len(chunks))
	return chunks, nil
}

func GenerateEmbeddingOfChunks(chunks *[]models.Chunk) (*[]models.Chunk, error) {
	log.Printf("Started generating embeddings for %d chunks", len(*chunks))

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	client := openai.NewClient(apiKey)

	for i, chunk := range *chunks {
		if strings.TrimSpace(chunk.Content) == "" {
			log.Printf("Skipping empty chunk %d", i+1)
			continue
		}

		resp, err := client.CreateEmbeddings(
			context.Background(),
			openai.EmbeddingRequest{
				Input: []string{chunk.Content},
				Model: openai.SmallEmbedding3,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create embedding for chunk %d: %w", i+1, err)
		}

		if len(resp.Data) == 0 {
			return nil, fmt.Errorf("no embedding data returned for chunk %d", i+1)
		}

		embedding := pgvector.NewVector(resp.Data[0].Embedding)
		(*chunks)[i].Embedding = embedding

		log.Printf("Generated embedding for chunk %d/%d", i+1, len(*chunks))
	}

	log.Printf("Successfully generated embeddings for all chunks")
	return chunks, nil
}

func ExtractTextFromFile(fileHeader *multipart.FileHeader, fileType string) (string, error) {
	tikaClient := tika.NewClient()

	if err := tikaClient.HealthCheck(); err != nil {
		log.Printf("Warning: Tika server health check failed: %v", err)
		return "", fmt.Errorf("tika server is not available: %w", err)
	}

	extractedText, err := tikaClient.ExtractText(fileHeader)
	if err != nil {
		return "", fmt.Errorf("tika text extraction failed: %w", err)
	}

	extractedText = strings.TrimSpace(extractedText)

	if extractedText == "" {
		return "", fmt.Errorf("no text could be extracted from file")
	}

	log.Printf("Successfully extracted %d characters from file using Tika", len(extractedText))
	return extractedText, nil
}

func SplitIntoTextChunks(text string, maxCharsPerChunk int) []string {
	sentences := SplitIntoSentences(text)
	var chunks []string
	currentChunk := ""

	for _, sentence := range sentences {
		if len(currentChunk)+len(sentence)+1 > maxCharsPerChunk && currentChunk != "" {
			chunks = append(chunks, strings.TrimSpace(currentChunk))
			currentChunk = sentence
		} else {
			if currentChunk == "" {
				currentChunk = sentence
			} else {
				currentChunk += " " + sentence
			}
		}
	}

	if strings.TrimSpace(currentChunk) != "" {
		chunks = append(chunks, strings.TrimSpace(currentChunk))
	}

	return chunks
}

func SplitIntoSentences(text string) []string {
	text = strings.ReplaceAll(text, "! ", "!|")
	text = strings.ReplaceAll(text, "? ", "?|")
	text = strings.ReplaceAll(text, ". ", ".|")

	sentences := strings.Split(text, "|")
	var result []string

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence != "" && len(sentence) > 3 {
			result = append(result, sentence)
		}
	}

	return result
}
