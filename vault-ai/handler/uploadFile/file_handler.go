package uploadFile

import (
	"fmt"
	"log"
	"net/http"

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

	userId, ok := userIDStr.(uuid.UUID)
	if !ok {
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"Invalid user ID format in token", nil)
		return
	}

	// Parse metadata
	fileType := c.PostForm("type")
	fileTitle := c.PostForm("title")

	// Parse file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "file is required"})
		return
	}

	log.Printf("Processing file: %s (type: %s)", fileTitle, fileType)

	// Generate document ID
	newUUID := uuid.New()

	// Create chunks from file
	chunks, err := helperfunctions.ConvertToChunks(file, fileType, newUUID)
	if err != nil {
		log.Printf("Error creating chunks: %v", err)
		c.JSON(400, gin.H{"error": fmt.Sprintf("Failed to process file: %v", err)})
		return
	}

	// Validate chunks before processing
	if len(chunks) == 0 {
		c.JSON(400, gin.H{"error": "No content extracted from file"})
		return
	}

	// Return immediate success response
	c.JSON(200, gin.H{
		"message":     "File upload initiated",
		"document_id": newUUID.String(),
		"chunks":      len(chunks),
	})

	// Process in background
	log.Printf("Starting background processing for document: %s", newUUID.String())
	go helperfunctions.ProcessDocument(newUUID, userId, fileType, fileTitle, chunks, fh.db)
}

func (f *UploadFileHandler) DeleteFile(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists || userIDStr == "" {
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"User ID not found in token", nil)
		return
	}

	userId, ok := userIDStr.(uuid.UUID)
	if !ok {
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"Invalid user ID format in token", nil)
		return
	}

	// get the document id from the request
	documentId := c.Param("id")
	if documentId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	// delete the document from the database
	query := "delete from documents where user_id = ? AND id = ?"
	err := f.db.QueryRow(query, userId, documentId).Scan(&models.Document{})
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

	userId, ok := userIDStr.(uuid.UUID)
	if !ok {
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"Invalid user ID format in token", nil)
		return
	}

	// get all the document from the database for the user
	var documents []models.Document
	query := "select * from documents where user_id = ?"
	err := f.db.QueryRow(query, userId).Scan(&documents)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting documents from database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"documents": documents})
}
