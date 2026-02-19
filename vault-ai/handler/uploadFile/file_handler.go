package uploadFile

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"chat-app/internal/database"
	"chat-app/internal/models"
	"chat-app/internal/utils/errors"
	"chat-app/internal/utils/helperfunctions"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UploadFileHandler struct {
	db *database.DB
}

func NewUploadFileHandler(db *database.DB) *UploadFileHandler {
	return &UploadFileHandler{db: db}
}

func (fh *UploadFileHandler) UploadFile(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists || userIDStr == "" {
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"User ID not found in token", nil)
		return
	}

	userId, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"Invalid user ID format in token", nil)
		return
	}

	fileType := c.PostForm("type")
	fileTitle := c.PostForm("title")
	fileSizeStr := c.PostForm("file_size")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "file is required"})
		return
	}

	var fileSize int64
	if fileSizeStr != "" {
		if parsedSize, err := strconv.ParseInt(fileSizeStr, 10, 64); err == nil {
			fileSize = parsedSize
		} else {
			fileSize = file.Size
		}
	} else {
		fileSize = file.Size
	}

	log.Printf("Processing file: %s (type: %s, size: %d bytes)", fileTitle, fileType, fileSize)

	newUUID := uuid.New()

	chunks, err := helperfunctions.ConvertToChunks(file, fileType, newUUID)
	if err != nil {
		log.Printf("Error creating chunks: %v", err)
		c.JSON(400, gin.H{"error": fmt.Sprintf("Failed to process file: %v", err)})
		return
	}

	if len(chunks) == 0 {
		c.JSON(400, gin.H{"error": "No content extracted from file"})
		return
	}

	c.JSON(200, gin.H{
		"message":     "File upload initiated",
		"document_id": newUUID.String(),
		"chunks":      len(chunks),
	})

	log.Printf("Starting background processing for document: %s", newUUID.String())
	helperfunctions.ProcessDocument(newUUID, userId, fileType, fileTitle, fileSize, chunks, fh.db)
}

func (f *UploadFileHandler) DeleteFile(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists || userIDStr == "" {
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"User ID not found in token", nil)
		return
	}

	userId, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"Invalid user ID format in token", nil)
		return
	}

	documentId := c.Param("id")
	if documentId == "" {
		documentId = c.Query("id")
		if documentId == "" {
			documentId = c.PostForm("id")
		}
	}
	if documentId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	query := "delete from documents where user_id = $1 AND id = $2"
	_, err = f.db.Exec(query, userId, documentId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting document from database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document deleted successfully"})
}

func (f *UploadFileHandler) GetFile(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists || userIDStr == "" {
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"User ID not found in token", nil)
		return
	}

	userId, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"Invalid user ID format in token", nil)
		return
	}

	type DocumentResponse struct {
		ID         string `json:"id"`
		UserId     string `json:"user_id"`
		FileName   string `json:"file_name"`
		FileType   string `json:"file_type"`
		FileSize   int64  `json:"file_size"`
		UploadedAt string `json:"uploaded_at"`
	}

	var documents []DocumentResponse
	query := "select id, user_id, file_name, file_type, file_size, created_at from documents where user_id = $1 ORDER BY created_at DESC"

	rows, err := f.db.Query(query, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting documents from database"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var doc models.Document
		if err := rows.Scan(&doc.ID, &doc.UserId, &doc.FileName, &doc.FileType, &doc.FileSize, &doc.CreatedAt); err != nil {
			log.Printf("Error scanning document row: %v", err)
			continue
		}
		documents = append(documents, DocumentResponse{
			ID:         doc.ID.String(),
			UserId:     doc.UserId.String(),
			FileName:   doc.FileName,
			FileType:   doc.FileType,
			FileSize:   doc.FileSize,
			UploadedAt: doc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating documents"})
		return
	}

	if documents == nil {
		documents = []DocumentResponse{}
	}

	c.JSON(http.StatusOK, gin.H{"documents": documents})
}
