package answer

import (
	"chat-app/internal/utils/errors"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (ah *AnswerHandler) StreamAnswer(c *gin.Context) {
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
	sessionID, err := uuid.Parse(c.Query("sessionID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}
	if sessionID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
		return
	}

	sessionManager := NewSessionManager(ah.db)
	_, err = sessionManager.GetSession(userID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting session"})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Create channel for this session's AI response
	responseChan := make(chan string, 100)
	GlobalBroadcaster.RegisterClient(sessionID, responseChan)
	defer GlobalBroadcaster.UnRegisterClient(sessionID, responseChan)

	// Stream responses
	for {
		select {
		case chunk := <-responseChan:
			if chunk == "[DONE]" {
				return
			}

			data := map[string]interface{}{
				"content":   chunk,
				"sessionId": sessionID,
				"timestamp": time.Now(),
			}

			jsonData, _ := json.Marshal(data)
			fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
			c.Writer.Flush()

		case <-c.Request.Context().Done():
			return
		}
	}
}
