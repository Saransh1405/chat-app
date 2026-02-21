package answer

import (
	"chat-app/internal/utils/errors"
	"chat-app/internal/utils/jwt"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (ah *AnswerHandler) StreamAnswer(c *gin.Context) {
	jwtSecret := os.Getenv("JWT_SECRET")
	token := c.Query("token")
	if token == "" {
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"Token is required", nil)
		return
	}

	claims, err := jwt.ValidateToken(token, jwtSecret)
	if err != nil {
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"Invalid token", nil)
		return
	}

	userId := claims.UserID
	userID, err := uuid.Parse(userId)
	if err != nil {
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

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	responseChan := make(chan string, 100)
	GlobalBroadcaster.RegisterClient(sessionID, responseChan)
	defer GlobalBroadcaster.UnRegisterClient(sessionID, responseChan)
	for {
		select {
		case chunk := <-responseChan:
			if chunk == "[DONE]" {
				return
			}

			var data map[string]interface{}
			if chunk == "[TYPING]" {
				data = map[string]interface{}{
					"type":      "typing",
					"sessionId": sessionID,
					"timestamp": time.Now(),
				}
			} else {
				data = map[string]interface{}{
					"type":      "content",
					"content":   chunk,
					"sessionId": sessionID,
					"timestamp": time.Now(),
				}
			}

			fmt.Println("data: ", data)

			jsonData, _ := json.Marshal(data)
			fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
			c.Writer.Flush()

		case <-c.Request.Context().Done():
			return
		}
	}
}
