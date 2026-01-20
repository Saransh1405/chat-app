package rest

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"chat-app/internal/database"
	"chat-app/internal/handler/websocket"
	"chat-app/internal/models"
	"chat-app/internal/utils/helperfunctions"
	"chat-app/internal/utils/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MessageHandler struct {
	db    *database.DB
	wsHub *websocket.Hub
}

func NewMessageHandler(db *database.DB, wsHub *websocket.Hub) *MessageHandler {
	return &MessageHandler{
		db:    db,
		wsHub: wsHub,
	}
}

func (h *MessageHandler) Create(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User ID not found in token",
			},
		})
		return
	}

	userID, ok := userIDStr.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Invalid user ID format in token",
			},
		})
		return
	}

	roomID := c.Param("room_id")
	if !validator.ValidateUUID(roomID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid room ID format",
			},
		})
		return
	}

	appID := c.Param("id")
	if !validator.ValidateUUID(appID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid application ID format",
			},
		})
		return
	}

	var messageData models.Message
	err := c.ShouldBindBodyWithJSON(&messageData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	if strings.TrimSpace(messageData.Content) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Message content cannot be empty",
			},
		})
		return
	}

	userIDUUID := uuid.MustParse(userID)
	roomIDUUID := uuid.MustParse(roomID)

	isMember, err := helperfunctions.ValidateUserIsMemberOfRoom(h.db, userID, roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to validate room membership",
			},
		})
		return
	}
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"code":    "FORBIDDEN",
				"message": "User is not a member of this room",
			},
		})
		return
	}

	messageData.UserID = userIDUUID
	messageData.RoomID = roomIDUUID

	if messageData.MessageType == "" {
		messageData.MessageType = models.MessageTypeText
	}

	message, err := helperfunctions.SaveMessageToDB(h.db, &messageData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to save message",
			},
		})
		return
	}

	h.wsHub.Broadcast <- struct {
		RoomID  string
		Message websocket.MessageStruct
	}{
		RoomID: message.RoomID.String(),
		Message: websocket.MessageStruct{
			Type:    "message",
			Payload: message,
		},
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": message,
		"status":  "sent",
	})
}

func (h *MessageHandler) Get(c *gin.Context) {
	messageID := c.Param("message_id")
	if messageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Message ID is required",
			},
		})
		return
	}

	if !validator.ValidateUUID(messageID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid message ID format",
			},
		})
		return
	}

	roomID := c.Param("room_id")
	appID := c.Param("id")

	if !validator.ValidateUUID(roomID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid room ID format",
			},
		})
		return
	}

	if !validator.ValidateUUID(appID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid application ID format",
			},
		})
		return
	}

	query := `SELECT m.id, m.room_id, m.user_id, m.content, m.message_type, m.reply_to, 
	         m.edited_at, m.deleted_at, m.metadata, m.created_at, m.updated_at
	         FROM messages m
	         INNER JOIN rooms r ON m.room_id = r.id
	         WHERE m.id = $1 AND m.room_id = $2 AND r.application_id = $3 AND m.deleted_at IS NULL`

	message := models.Message{}
	var metadataJSON []byte
	var replyToUUID *uuid.UUID

	err := h.db.QueryRow(query, messageID, roomID, appID).Scan(
		&message.ID,
		&message.RoomID,
		&message.UserID,
		&message.Content,
		&message.MessageType,
		&replyToUUID,
		&message.EditedAt,
		&message.DeletedAt,
		&metadataJSON,
		&message.CreatedAt,
		&message.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "MESSAGE_NOT_FOUND",
					"message": "Message not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to retrieve message",
			},
		})
		return
	}

	message.ReplyTo = replyToUUID

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &message.Metadata); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "Failed to parse message metadata",
				},
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
	})
}

func (h *MessageHandler) List(c *gin.Context) {
	roomID := c.Param("room_id")
	appID := c.Param("id")

	if !validator.ValidateUUID(roomID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid room ID format",
			},
		})
		return
	}

	if !validator.ValidateUUID(appID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid application ID format",
			},
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	limit := 50
	offset := 0

	if parsedLimit, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || parsedLimit != 1 || limit < 1 || limit > 100 {
		limit = 50
	}
	if parsedOffset, err := fmt.Sscanf(offsetStr, "%d", &offset); err != nil || parsedOffset != 1 || offset < 0 {
		offset = 0
	}

	query := `SELECT m.id, m.room_id, m.user_id, m.content, m.message_type, m.reply_to, 
	         m.edited_at, m.deleted_at, m.metadata, m.created_at, m.updated_at
	         FROM messages m
	         INNER JOIN rooms r ON m.room_id = r.id
	         WHERE m.room_id = $1 AND r.application_id = $2 AND m.deleted_at IS NULL
	         ORDER BY m.created_at DESC
	         LIMIT $3 OFFSET $4`

	rows, err := h.db.Query(query, roomID, appID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to retrieve messages",
			},
		})
		return
	}
	defer rows.Close()

	messages := []models.Message{}
	for rows.Next() {
		message := models.Message{}
		var metadataJSON []byte
		var replyToUUID *uuid.UUID

		err := rows.Scan(
			&message.ID,
			&message.RoomID,
			&message.UserID,
			&message.Content,
			&message.MessageType,
			&replyToUUID,
			&message.EditedAt,
			&message.DeletedAt,
			&metadataJSON,
			&message.CreatedAt,
			&message.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "Failed to parse message",
				},
			})
			return
		}

		message.ReplyTo = replyToUUID

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &message.Metadata); err != nil {
				continue
			}
		}

		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to retrieve messages",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"limit":    limit,
		"offset":   offset,
		"count":    len(messages),
	})
}

func (h *MessageHandler) Update(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User ID not found in token",
			},
		})
		return
	}

	userID, ok := userIDStr.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Invalid user ID format in token",
			},
		})
		return
	}

	messageID := c.Param("message_id")
	if messageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Message ID is required",
			},
		})
		return
	}

	if !validator.ValidateUUID(messageID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid message ID format",
			},
		})
		return
	}

	roomID := c.Param("room_id")
	appID := c.Param("id")

	if !validator.ValidateUUID(roomID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid room ID format",
			},
		})
		return
	}

	if !validator.ValidateUUID(appID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid application ID format",
			},
		})
		return
	}

	var updateData models.Message
	err := c.ShouldBindBodyWithJSON(&updateData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	if updateData.Content != "" && strings.TrimSpace(updateData.Content) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Message content cannot be empty",
			},
		})
		return
	}

	checkQuery := `SELECT m.user_id FROM messages m
	               INNER JOIN rooms r ON m.room_id = r.id
	               WHERE m.id = $1 AND m.room_id = $2 AND r.application_id = $3 AND m.deleted_at IS NULL`

	var existingUserID uuid.UUID
	err = h.db.QueryRow(checkQuery, messageID, roomID, appID).Scan(&existingUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "MESSAGE_NOT_FOUND",
					"message": "Message not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to check message existence",
			},
		})
		return
	}

	userIDUUID := uuid.MustParse(userID)
	if existingUserID != userIDUUID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"code":    "FORBIDDEN",
				"message": "You can only update your own messages",
			},
		})
		return
	}

	updateFields := []string{}
	args := []interface{}{}
	argIndex := 1

	if updateData.Content != "" {
		updateFields = append(updateFields, fmt.Sprintf("content = $%d", argIndex))
		args = append(args, updateData.Content)
		argIndex++
	}

	if updateData.MessageType != "" {
		updateFields = append(updateFields, fmt.Sprintf("message_type = $%d", argIndex))
		args = append(args, updateData.MessageType)
		argIndex++
	}

	if updateData.Metadata != nil {
		metadataJSON, err := json.Marshal(updateData.Metadata)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "Invalid metadata format",
				},
			})
			return
		}
		updateFields = append(updateFields, fmt.Sprintf("metadata = $%d", argIndex))
		args = append(args, metadataJSON)
		argIndex++
	}

	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "No fields to update",
			},
		})
		return
	}

	now := time.Now()
	updateFields = append(updateFields, fmt.Sprintf("edited_at = $%d", argIndex))
	args = append(args, now)
	argIndex++
	updateFields = append(updateFields, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, now)
	argIndex++

	whereClause := fmt.Sprintf("WHERE id = $%d AND room_id = $%d AND deleted_at IS NULL", argIndex, argIndex+1)
	args = append(args, messageID, roomID)

	updateQuery := fmt.Sprintf("UPDATE messages SET %s %s", strings.Join(updateFields, ", "), whereClause)

	result, err := h.db.Exec(updateQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to update message",
			},
		})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to update message",
			},
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "MESSAGE_NOT_FOUND",
				"message": "Message not found",
			},
		})
		return
	}

	fetchQuery := `SELECT m.id, m.room_id, m.user_id, m.content, m.message_type, m.reply_to, 
	               m.edited_at, m.deleted_at, m.metadata, m.created_at, m.updated_at
	               FROM messages m
	               WHERE m.id = $1 AND m.room_id = $2 AND m.deleted_at IS NULL`

	updatedMessage := models.Message{}
	var metadataJSON []byte
	var replyToUUID *uuid.UUID

	err = h.db.QueryRow(fetchQuery, messageID, roomID).Scan(
		&updatedMessage.ID,
		&updatedMessage.RoomID,
		&updatedMessage.UserID,
		&updatedMessage.Content,
		&updatedMessage.MessageType,
		&replyToUUID,
		&updatedMessage.EditedAt,
		&updatedMessage.DeletedAt,
		&metadataJSON,
		&updatedMessage.CreatedAt,
		&updatedMessage.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Message updated successfully",
		})
		return
	}

	updatedMessage.ReplyTo = replyToUUID

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &updatedMessage.Metadata); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"message": "Message updated successfully",
				"data":    updatedMessage,
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Message updated successfully",
		"data":    updatedMessage,
	})
}

func (h *MessageHandler) Delete(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User ID not found in token",
			},
		})
		return
	}

	userID, ok := userIDStr.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Invalid user ID format in token",
			},
		})
		return
	}

	messageID := c.Param("message_id")
	if messageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Message ID is required",
			},
		})
		return
	}

	if !validator.ValidateUUID(messageID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid message ID format",
			},
		})
		return
	}

	roomID := c.Param("room_id")
	appID := c.Param("id")

	if !validator.ValidateUUID(roomID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid room ID format",
			},
		})
		return
	}

	if !validator.ValidateUUID(appID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid application ID format",
			},
		})
		return
	}

	checkQuery := `SELECT m.user_id FROM messages m
	               INNER JOIN rooms r ON m.room_id = r.id
	               WHERE m.id = $1 AND m.room_id = $2 AND r.application_id = $3 AND m.deleted_at IS NULL`

	var existingUserID uuid.UUID
	err := h.db.QueryRow(checkQuery, messageID, roomID, appID).Scan(&existingUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "MESSAGE_NOT_FOUND",
					"message": "Message not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to check message existence",
			},
		})
		return
	}

	userIDUUID := uuid.MustParse(userID)
	if existingUserID != userIDUUID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"code":    "FORBIDDEN",
				"message": "You can only delete your own messages",
			},
		})
		return
	}

	now := time.Now()
	deleteQuery := `UPDATE messages SET deleted_at = $1, updated_at = $2 
	                WHERE id = $3 AND room_id = $4 AND deleted_at IS NULL`

	result, err := h.db.Exec(deleteQuery, now, now, messageID, roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to delete message",
			},
		})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to delete message",
			},
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "MESSAGE_NOT_FOUND",
				"message": "Message not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Message deleted successfully",
	})
}
