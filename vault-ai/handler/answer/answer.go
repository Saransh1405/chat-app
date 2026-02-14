package answer

import (
	"chat-app/internal/database"
	"chat-app/internal/models"
	"chat-app/internal/utils/errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AnswerHandler struct {
	db *database.DB
}

func NewAnswerHandler(db *database.DB) *AnswerHandler {
	return &AnswerHandler{db: db}
}

func (ah *AnswerHandler) Answer(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists || userIDStr == "" {
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"User ID not found in token", nil)
		return
	}

	userID, ok := userIDStr.(uuid.UUID)
	if !ok {
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"Invalid user ID format in token", nil)
		return
	}

	var req models.AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessionManager := NewSessionManager(ah.db)
	var session *models.ChatSession
	var err error
	if req.SessionID != nil {
		session, err = sessionManager.GetSession(userID, *req.SessionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting session"})
			return
		}

		// add user message to session
		err = sessionManager.AddMessage(session.ID, "user", req.Question)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error adding message to session"})
			return
		}
	} else {
		// Create new session
		session, err = sessionManager.CreateSession(userID, req.Question)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
			return
		}
	}

	// Start AI processing in background
	go AIProcessing(c, session.ID, req.Question, userID, ah.db)

	c.JSON(http.StatusOK, gin.H{
		"sessionId": session.ID,
		"message":   "Question received, processing...",
	})
}
