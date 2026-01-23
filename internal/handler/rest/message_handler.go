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
	"chat-app/internal/utils/errors"
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
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"User ID not found in token", nil)
		return
	}

	userID, ok := userIDStr.(string)
	if !ok {
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"Invalid user ID format in token", nil)
		return
	}

	var req models.MessageCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		details := []errors.ErrorDetail{
			{
				Field: "request_body",
				Issue: "Invalid JSON format or missing required fields",
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid request body", details)
		return
	}

	if strings.TrimSpace(req.Content) == "" {
		details := []errors.ErrorDetail{
			{
				Field: "content",
				Issue: "Message content cannot be empty",
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Message content cannot be empty", details)
		return
	}

	appID := ""
	if req.ApplicationID != nil {
		appID = req.ApplicationID.String()
		if !validator.ValidateUUID(appID) {
			details := []errors.ErrorDetail{
				{
					Field: "application_id",
					Issue: "Invalid UUID format",
					Value: appID,
				},
			}
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Invalid application ID format", details)
			return
		}
	}

	userIDUUID := uuid.MustParse(userID)
	roomIDUUID := req.RoomID

	isMember, err := helperfunctions.ValidateUserIsMemberOfRoom(h.db, userID, req.RoomID.String())
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to validate room membership", nil)
		return
	}
	if !isMember {
		errors.RespondWithError(c, http.StatusForbidden, errors.ErrCodeForbidden,
			"User is not a member of this room", nil)
		return
	}

	messageData := models.Message{
		UserID:      userIDUUID,
		RoomID:      roomIDUUID,
		Content:     req.Content,
		MessageType: req.MessageType,
		ReplyTo:     req.ReplyTo,
		Metadata:    req.Metadata,
	}

	if messageData.MessageType == "" {
		messageData.MessageType = models.MessageTypeText
	}

	message, err := helperfunctions.SaveMessageToDB(c, h.db, &messageData)
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to save message", nil)
		return
	}

	go func() {
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
	}()

	c.JSON(http.StatusCreated, gin.H{
		"message": message,
		"status":  "sent",
	})
}

func (h *MessageHandler) Get(c *gin.Context) {
	messageID := c.Query("id")
	roomID := c.Query("room_id")
	appID := c.Query("application_id")

	if messageID == "" {
		details := []errors.ErrorDetail{
			{
				Field: "id",
				Issue: "Message ID is required",
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Message ID is required", details)
		return
	}

	if !validator.ValidateUUID(messageID) {
		details := []errors.ErrorDetail{
			{
				Field: "id",
				Issue: "Invalid UUID format",
				Value: messageID,
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid message ID format", details)
		return
	}

	if roomID == "" {
		details := []errors.ErrorDetail{
			{
				Field: "room_id",
				Issue: "Room ID is required",
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Room ID is required", details)
		return
	}

	if !validator.ValidateUUID(roomID) {
		details := []errors.ErrorDetail{
			{
				Field: "room_id",
				Issue: "Invalid UUID format",
				Value: roomID,
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid room ID format", details)
		return
	}

	var query string
	var args []interface{}

	if appID != "" {
		if !validator.ValidateUUID(appID) {
			details := []errors.ErrorDetail{
				{
					Field: "application_id",
					Issue: "Invalid UUID format",
					Value: appID,
				},
			}
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Invalid application ID format", details)
			return
		}
		query = `SELECT m.id, m.room_id, m.user_id, m.content, m.message_type, m.reply_to, 
		         m.edited_at, m.deleted_at, m.metadata, m.created_at, m.updated_at
		         FROM messages m
		         INNER JOIN rooms r ON m.room_id = r.id
		         WHERE m.id = $1 AND m.room_id = $2 AND r.application_id = $3 AND m.deleted_at IS NULL`
		args = []interface{}{messageID, roomID, appID}
	} else {
		query = `SELECT m.id, m.room_id, m.user_id, m.content, m.message_type, m.reply_to, 
		         m.edited_at, m.deleted_at, m.metadata, m.created_at, m.updated_at
		         FROM messages m
		         WHERE m.id = $1 AND m.room_id = $2 AND m.deleted_at IS NULL`
		args = []interface{}{messageID, roomID}
	}

	message := models.Message{}
	var metadataJSON []byte
	var replyToUUID *uuid.UUID

	err := h.db.QueryRow(query, args...).Scan(
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
			errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
				"Message not found", nil)
			return
		}
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to retrieve message", nil)
		return
	}

	message.ReplyTo = replyToUUID

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &message.Metadata); err != nil {
			errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeInternalError,
				"Failed to parse message metadata", nil)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
	})
}

func (h *MessageHandler) List(c *gin.Context) {
	roomID := c.Query("room_id")
	appID := c.Query("application_id")

	if roomID == "" {
		details := []errors.ErrorDetail{
			{
				Field: "room_id",
				Issue: "Room ID is required",
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Room ID is required", details)
		return
	}

	if !validator.ValidateUUID(roomID) {
		details := []errors.ErrorDetail{
			{
				Field: "room_id",
				Issue: "Invalid UUID format",
				Value: roomID,
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid room ID format", details)
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

	var rows *sql.Rows
	var err error

	if appID != "" {
		if !validator.ValidateUUID(appID) {
			details := []errors.ErrorDetail{
				{
					Field: "application_id",
					Issue: "Invalid UUID format",
					Value: appID,
				},
			}
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Invalid application ID format", details)
			return
		}
		query := `SELECT m.id, m.room_id, m.user_id, m.content, m.message_type, m.reply_to, 
		         m.edited_at, m.deleted_at, m.metadata, m.created_at, m.updated_at
		         FROM messages m
		         INNER JOIN rooms r ON m.room_id = r.id
		         WHERE m.room_id = $1 AND r.application_id = $2 AND m.deleted_at IS NULL
		         ORDER BY m.created_at DESC
		         LIMIT $3 OFFSET $4`

		rows, err = h.db.Query(query, roomID, appID, limit, offset)
	} else {
		query := `SELECT m.id, m.room_id, m.user_id, m.content, m.message_type, m.reply_to, 
		         m.edited_at, m.deleted_at, m.metadata, m.created_at, m.updated_at
		         FROM messages m
		         WHERE m.room_id = $1 AND m.deleted_at IS NULL
		         ORDER BY m.created_at DESC
		         LIMIT $2 OFFSET $3`

		rows, err = h.db.Query(query, roomID, limit, offset)
	}
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to retrieve messages", nil)
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
			errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
				"Failed to parse message", nil)
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
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to retrieve messages", nil)
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
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"User ID not found in token", nil)
		return
	}

	userID, ok := userIDStr.(string)
	if !ok {
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"Invalid user ID format in token", nil)
		return
	}

	var req models.MessageUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		details := []errors.ErrorDetail{
			{
				Field: "request_body",
				Issue: "Invalid JSON format or missing required fields",
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid request body", details)
		return
	}

	if req.Content != "" && strings.TrimSpace(req.Content) == "" {
		details := []errors.ErrorDetail{
			{
				Field: "content",
				Issue: "Message content cannot be empty",
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Message content cannot be empty", details)
		return
	}

	appID := req.ApplicationID
	var checkQuery string
	var existingUserID uuid.UUID
	var err error

	if appID == nil {
		checkQuery = `SELECT m.user_id FROM messages m
		               WHERE m.id = $1 AND m.room_id = $2 AND m.deleted_at IS NULL`
		err = h.db.QueryRow(checkQuery, req.ID, req.RoomID).Scan(&existingUserID)
	} else {
		checkQuery = `SELECT m.user_id FROM messages m
		               INNER JOIN rooms r ON m.room_id = r.id
		               WHERE m.id = $1 AND m.room_id = $2 AND r.application_id = $3 AND m.deleted_at IS NULL`
		err = h.db.QueryRow(checkQuery, req.ID, req.RoomID, appID).Scan(&existingUserID)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
				"Message not found", nil)
			return
		}
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to check message existence", nil)
		return
	}

	userIDUUID := uuid.MustParse(userID)
	if existingUserID != userIDUUID {
		errors.RespondWithError(c, http.StatusForbidden, errors.ErrCodeForbidden,
			"You can only update your own messages", nil)
		return
	}

	updateFields := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Content != "" {
		updateFields = append(updateFields, fmt.Sprintf("content = $%d", argIndex))
		args = append(args, req.Content)
		argIndex++
	}

	if req.MessageType != "" {
		updateFields = append(updateFields, fmt.Sprintf("message_type = $%d", argIndex))
		args = append(args, req.MessageType)
		argIndex++
	}

	if req.Metadata != nil {
		metadataJSON, err := json.Marshal(req.Metadata)
		if err != nil {
			details := []errors.ErrorDetail{
				{
					Field: "metadata",
					Issue: "Invalid JSON format",
				},
			}
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Invalid metadata format", details)
			return
		}
		updateFields = append(updateFields, fmt.Sprintf("metadata = $%d", argIndex))
		args = append(args, metadataJSON)
		argIndex++
	}

	if len(updateFields) == 0 {
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"No fields to update", nil)
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
	args = append(args, req.ID, req.RoomID)

	updateQuery := fmt.Sprintf("UPDATE messages SET %s %s", strings.Join(updateFields, ", "), whereClause)

	result, err := h.db.Exec(updateQuery, args...)
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to update message", nil)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to update message", nil)
		return
	}

	if rowsAffected == 0 {
		errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
			"Message not found", nil)
		return
	}

	fetchQuery := `SELECT m.id, m.room_id, m.user_id, m.content, m.message_type, m.reply_to, 
	               m.edited_at, m.deleted_at, m.metadata, m.created_at, m.updated_at
	               FROM messages m
	               WHERE m.id = $1 AND m.room_id = $2 AND m.deleted_at IS NULL`

	updatedMessage := models.Message{}
	var metadataJSON []byte
	var replyToUUID *uuid.UUID

	err = h.db.QueryRow(fetchQuery, req.ID, req.RoomID).Scan(
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

	go func() {
		h.wsHub.Broadcast <- struct {
			RoomID  string
			Message websocket.MessageStruct
		}{
			RoomID: updatedMessage.RoomID.String(),
			Message: websocket.MessageStruct{
				Type:    "message_updated",
				Payload: updatedMessage,
			},
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Message updated successfully",
		"data":    updatedMessage,
	})
}

func (h *MessageHandler) Delete(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists || userIDStr == "" {
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"User ID not found in token", nil)
		return
	}

	userID, ok := userIDStr.(string)
	if !ok {
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"Invalid user ID format in token", nil)
		return
	}

	var req models.MessageDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		details := []errors.ErrorDetail{
			{
				Field: "request_body",
				Issue: "Invalid JSON format or missing required fields",
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid request body", details)
		return
	}

	appID := req.ApplicationID
	var checkQuery string
	var existingUserID uuid.UUID
	var err error

	if appID == nil {
		checkQuery = `SELECT m.user_id FROM messages m
		               WHERE m.id = $1 AND m.room_id = $2 AND m.deleted_at IS NULL`
		err = h.db.QueryRow(checkQuery, req.ID, req.RoomID).Scan(&existingUserID)
	} else {
		checkQuery = `SELECT m.user_id FROM messages m
		               INNER JOIN rooms r ON m.room_id = r.id
		               WHERE m.id = $1 AND m.room_id = $2 AND r.application_id = $3 AND m.deleted_at IS NULL`
		err = h.db.QueryRow(checkQuery, req.ID, req.RoomID, appID).Scan(&existingUserID)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
				"Message not found", nil)
			return
		}
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to check message existence", nil)
		return
	}

	userIDUUID := uuid.MustParse(userID)
	if existingUserID != userIDUUID {
		errors.RespondWithError(c, http.StatusForbidden, errors.ErrCodeForbidden,
			"You can only delete your own messages", nil)
		return
	}

	now := time.Now()
	deleteQuery := `UPDATE messages SET deleted_at = $1, updated_at = $2 
	                WHERE id = $3 AND room_id = $4 AND deleted_at IS NULL`

	result, err := h.db.Exec(deleteQuery, now, now, req.ID, req.RoomID)
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to delete message", nil)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to delete message", nil)
		return
	}

	if rowsAffected == 0 {
		errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
			"Message not found", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Message deleted successfully",
	})
}
