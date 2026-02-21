package answer

import (
	"chat-app/internal/database"
	"chat-app/internal/models"
	"chat-app/internal/utils/errors"
	"log"
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

	userID, ok := uuid.Parse(userIDStr.(string))
	if ok != nil {
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

		err = sessionManager.AddMessage(session.ID, "user", req.Question)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error adding message to session"})
			return
		}
	} else {
		session, err = sessionManager.CreateSession(userID, req.Question)
		if err != nil {
			log.Printf("Error creating session: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
			return
		}
	}

	go AIProcessing(c, session.ID, req.Question, userID, ah.db)

	c.JSON(http.StatusOK, gin.H{
		"sessionId": session.ID,
		"message":   "Question received, processing...",
	})
}

func (ah *AnswerHandler) GetHistory(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists || userIDStr == "" {
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"User ID not found in token", nil)
		return
	}

	sessionIDStr := c.Query("sessionID")
	if sessionIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
		return
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	var messages []models.ChatMessage
	query := `SELECT * FROM chat_messages WHERE session_id = $1 ORDER BY timestamp ASC`
	rows, err := ah.db.Query(query, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting messages"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var msg models.ChatMessage
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.Timestamp); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	if messages == nil {
		messages = []models.ChatMessage{}
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}
