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
	"chat-app/internal/utils/logger"

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

	messageType := strings.ToUpper(strings.TrimSpace(req.Type))
	if messageType == "" {
		if string(req.MessageType) != "" {
			messageType = strings.ToUpper(string(req.MessageType))
		} else {
			if req.FileRequest != nil {
				messageType = "MEDIA"
			} else {
				messageType = "TEXT"
			}
		}
	}

	if messageType == "TEXT" {
		if req.FileRequest != nil {
			messageType = "MEDIA"
		} else if strings.TrimSpace(req.Content) == "" {
			details := []errors.ErrorDetail{
				{
					Field: "content",
					Issue: "Message content cannot be empty for TEXT messages",
				},
			}
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Message content cannot be empty", details)
			return
		}
	} else if messageType == "MEDIA" {
		if req.FileRequest == nil {
			details := []errors.ErrorDetail{
				{
					Field: "file_request",
					Issue: "File data is required for MEDIA messages",
				},
			}
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"File data is required", details)
			return
		}
	}

	userIDUUID, err := uuid.Parse(userID)
	if err != nil {
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"Invalid user ID format in token", nil)
		return
	}
	roomIDUUID := req.RoomID

	isMember, err := helperfunctions.ValidateUserIsMemberOfRoom(h.db, userID, req.RoomID.String())
	if err != nil {
		logger.Error("Failed to validate room membership", err, logger.Fields{
			"user_id": userID,
			"room_id": req.RoomID.String(),
		})
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to validate room membership", nil)
		return
	}
	if !isMember {
		logger.Warn("User attempted to send message to room they're not a member of", logger.Fields{
			"user_id": userID,
			"room_id": req.RoomID.String(),
		})
		errors.RespondWithError(c, http.StatusForbidden, errors.ErrCodeForbidden,
			"User is not a member of this room", nil)
		return
	}

	var messageTypeEnum models.MessageType
	if messageType == "MEDIA" {
		if req.FileRequest != nil {
			if strings.HasPrefix(req.FileRequest.MimeType, "image/") {
				messageTypeEnum = models.MessageTypeImage
			} else {
				messageTypeEnum = models.MessageTypeFile
			}
		} else {
			messageTypeEnum = models.MessageTypeFile
		}
	} else {
		messageTypeEnum = models.MessageTypeText
	}

	messageData := models.Message{
		UserID:      userIDUUID,
		RoomID:      roomIDUUID,
		Content:     req.Content,
		MessageType: messageTypeEnum,
		ReplyTo:     req.ReplyTo,
		Metadata:    req.Metadata,
	}

	message, err := helperfunctions.SaveMessageToDB(c, h.db, &messageData)
	if err != nil {
		logger.Error("Failed to save message to database", err, logger.Fields{
			"user_id":     userID,
			"room_id":     req.RoomID.String(),
			"content_len": len(req.Content),
		})
		c.Error(err)
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to save message", nil)
		return
	}

	logger.Info("Message created successfully", logger.Fields{
		"message_id": message.ID.String(),
		"user_id":    userID,
		"room_id":    req.RoomID.String(),
		"type":       message.MessageType,
	})

	var file *models.File
	if messageType == "MEDIA" {
		if req.FileRequest == nil {
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"File data is required for media messages", []errors.ErrorDetail{
					{Field: "file_request", Issue: "File request is required when type is MEDIA"},
				})
			return
		}

		if req.FileRequest.FilePath == "" {
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"File path is required", []errors.ErrorDetail{
					{Field: "file_request.file_path", Issue: "File path cannot be empty"},
				})
			return
		}

		if req.FileRequest.Filename == "" {
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Filename is required", []errors.ErrorDetail{
					{Field: "file_request.filename", Issue: "Filename cannot be empty"},
				})
			return
		}

		if req.FileRequest.FileSize <= 0 {
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"File size is required", []errors.ErrorDetail{
					{Field: "file_request.file_size", Issue: "File size must be greater than 0"},
				})
			return
		}

		file = &models.File{
			Filename:   req.FileRequest.Filename,
			FilePath:   req.FileRequest.FilePath,
			FileSize:   req.FileRequest.FileSize,
			MimeType:   req.FileRequest.MimeType,
			MessageID:  message.ID,
			UserID:     userIDUUID,
			UploadedAt: time.Now(),
		}

		file, err := helperfunctions.SaveFileToDB(c, h.db, file)
		if err != nil {
			logger.Error("Failed to save file to database", err, logger.Fields{
				"message_id": message.ID.String(),
				"file_path":  req.FileRequest.FilePath,
			})
			errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
				"Message created but failed to save file record", nil)
			return
		}

		logger.Info("File record saved successfully", logger.Fields{
			"file_id":    file.ID.String(),
			"message_id": message.ID.String(),
		})
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Panic in WebSocket broadcast", fmt.Errorf("%v", r), logger.Fields{
					"room_id":    message.RoomID.String(),
					"message_id": message.ID.String(),
				})
			}
		}()

		var user *models.User
		userQuery := `SELECT id, username, email, avatar_url, application_id FROM users WHERE id = $1 AND deleted_at IS NULL LIMIT 1`
		userRow := h.db.QueryRow(userQuery, message.UserID)
		userData := models.User{}
		var email sql.NullString
		var avatarURL *string
		var appID uuid.UUID
		err := userRow.Scan(&userData.ID, &userData.Username, &email, &avatarURL, &appID)
		if err == nil {
			if email.Valid {
				emailStr := email.String
				userData.Email = &emailStr
			}
			userData.AvatarURL = avatarURL
			userData.ApplicationID = appID
			user = &userData
		} else {
			logger.Warn("Failed to fetch user for message broadcast", logger.Fields{
				"user_id":    message.UserID.String(),
				"message_id": message.ID.String(),
				"error":      err.Error(),
			})
		}

		messageWithUser := message
		messageWithUser.User = user

		messagePayload := map[string]interface{}{
			"message": messageWithUser,
		}
		if file != nil {
			messagePayload["file"] = file
		}

		h.wsHub.Broadcast <- struct {
			RoomID  string
			Message websocket.MessageStruct
		}{
			RoomID: message.RoomID.String(),
			Message: websocket.MessageStruct{
				Type:    "message",
				Payload: messagePayload,
			},
		}
		logger.Debug("Message broadcasted via WebSocket", logger.Fields{
			"room_id":    message.RoomID.String(),
			"message_id": message.ID.String(),
			"has_user":   user != nil,
			"has_file":   file != nil,
		})
	}()

	response := gin.H{
		"status":  "sent",
		"message": message,
	}
	if file != nil {
		response["file"] = file
	}

	c.JSON(http.StatusCreated, response)
}

func (h *MessageHandler) Get(c *gin.Context) {
	messageID := c.Query("id")
	roomID := c.Query("room_id")
	appID := c.Query("application_id")

	if messageID == "" {
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Message ID is required", nil)
		return
	}

	messageIDUUID, err := uuid.Parse(messageID)
	if err != nil {
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid message ID format", nil)
		return
	}

	if roomID == "" {
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Room ID is required", nil)
		return
	}

	roomIDUUID, err := uuid.Parse(roomID)
	if err != nil {
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid room ID format", nil)
		return
	}

	var query string
	var args []interface{}

	if appID != "" {
		appIDUUID, err := uuid.Parse(appID)
		if err != nil {
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Invalid application ID format", nil)
			return
		}
		query = `SELECT m.id, m.room_id, m.user_id, m.content, m.message_type, m.reply_to, 
		         m.edited_at, m.deleted_at, m.metadata, m.created_at, m.updated_at
		         FROM messages m
		         INNER JOIN rooms r ON m.room_id = r.id
		         WHERE m.id = $1 AND m.room_id = $2 AND r.application_id = $3 AND m.deleted_at IS NULL`
		args = []interface{}{messageIDUUID, roomIDUUID, appIDUUID}
	} else {
		query = `SELECT m.id, m.room_id, m.user_id, m.content, m.message_type, m.reply_to, 
		         m.edited_at, m.deleted_at, m.metadata, m.created_at, m.updated_at
		         FROM messages m
		         WHERE m.id = $1 AND m.room_id = $2 AND m.deleted_at IS NULL`
		args = []interface{}{messageIDUUID, roomIDUUID}
	}

	message := models.Message{}
	var metadataJSON []byte
	var replyToUUID *uuid.UUID

	err = h.db.QueryRow(query, args...).Scan(
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
		logger.Error("Failed to retrieve message", err, logger.Fields{
			"message_id": messageID,
			"room_id":    roomID,
		})
		c.Error(err)
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to retrieve message", nil)
		return
	}

	message.ReplyTo = replyToUUID

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &message.Metadata); err != nil {
			logger.Error("Failed to parse message metadata", err, logger.Fields{
				"message_id": message.ID.String(),
			})
			c.Error(err)
			errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeInternalError,
				"Failed to parse message metadata", nil)
			return
		}
	}

	// Fetch file data if message has a file
	var file *models.File
	fileQuery := `SELECT id, application_id, message_id, user_id, filename, file_path, file_size, mime_type, uploaded_at, deleted_at
	              FROM files WHERE message_id = $1 AND deleted_at IS NULL LIMIT 1`
	fileRow := h.db.QueryRow(fileQuery, message.ID)
	fileData := models.File{}
	var fileDeletedAt *time.Time
	err = fileRow.Scan(
		&fileData.ID,
		&fileData.ApplicationID,
		&fileData.MessageID,
		&fileData.UserID,
		&fileData.Filename,
		&fileData.FilePath,
		&fileData.FileSize,
		&fileData.MimeType,
		&fileData.UploadedAt,
		&fileDeletedAt,
	)
	if err == nil {
		file = &fileData
	} else if err != sql.ErrNoRows {
		logger.Warn("Failed to fetch file for message", logger.Fields{
			"message_id": message.ID.String(),
			"error":      err.Error(),
		})
	}

	response := gin.H{
		"message": message,
	}
	if file != nil {
		response["file"] = file
	}

	c.JSON(http.StatusOK, response)
}

func (h *MessageHandler) List(c *gin.Context) {
	roomID := c.Query("room_id")
	appID := c.Query("application_id")

	if roomID == "" {
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Room ID is required", nil)
		return
	}

	roomIDUUID, err := uuid.Parse(roomID)
	if err != nil {
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid room ID format", nil)
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

	if appID != "" {
		appIDUUID, err := uuid.Parse(appID)
		if err != nil {
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Invalid application ID format", nil)
			return
		}
		query := `SELECT m.id, m.room_id, m.user_id, m.content, m.message_type, m.reply_to, 
		         m.edited_at, m.deleted_at, m.metadata, m.created_at, m.updated_at,
		         u.username, u.avatar_url
		         FROM messages m
		         INNER JOIN rooms r ON m.room_id = r.id
		         INNER JOIN users u ON m.user_id = u.id
		         WHERE m.room_id = $1 AND r.application_id = $2 AND m.deleted_at IS NULL
		         ORDER BY m.created_at DESC
		         LIMIT $3 OFFSET $4`

		rows, err = h.db.Query(query, roomIDUUID, appIDUUID, limit, offset)
	} else {
		query := `SELECT m.id, m.room_id, m.user_id, m.content, m.message_type, m.reply_to, 
		         m.edited_at, m.deleted_at, m.metadata, m.created_at, m.updated_at,
		         u.username, u.avatar_url
		         FROM messages m
		         INNER JOIN users u ON m.user_id = u.id
		         WHERE m.room_id = $1 AND m.deleted_at IS NULL
		         ORDER BY m.created_at DESC
		         LIMIT $2 OFFSET $3`

		rows, err = h.db.Query(query, roomIDUUID, limit, offset)
	}
	if err != nil {
		logger.Error("Failed to retrieve messages from database", err, logger.Fields{
			"room_id": roomID,
			"app_id":  appID,
			"limit":   limit,
			"offset":  offset,
		})
		c.Error(err)
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

		var username string
		var avatarURL *string
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
			&username,
			&avatarURL,
		)
		if err != nil {
			errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
				"Failed to parse message", nil)
			return
		}

		message.User = &models.User{
			ID:        message.UserID,
			Username:  username,
			AvatarURL: avatarURL,
		}

		message.ReplyTo = replyToUUID

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &message.Metadata); err != nil {
				continue
			}
		}

		// Fetch file data for this message
		fileQuery := `SELECT id, application_id, message_id, user_id, filename, file_path, file_size, mime_type, uploaded_at, deleted_at
		              FROM files WHERE message_id = $1 AND deleted_at IS NULL LIMIT 1`
		fileRow := h.db.QueryRow(fileQuery, message.ID)
		fileData := models.File{}
		var fileDeletedAt *time.Time
		fileErr := fileRow.Scan(
			&fileData.ID,
			&fileData.ApplicationID,
			&fileData.MessageID,
			&fileData.UserID,
			&fileData.Filename,
			&fileData.FilePath,
			&fileData.FileSize,
			&fileData.MimeType,
			&fileData.UploadedAt,
			&fileDeletedAt,
		)
		if fileErr == nil {
			// File found - we'll include it in the response
			// For now, we'll create a response structure that includes files
		}

		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		logger.Error("Row error while retrieving messages", err, logger.Fields{
			"room_id": roomID,
		})
		c.Error(err)
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to retrieve messages", nil)
		return
	}

	// Fetch all files for these messages in one query
	messageIDs := make([]uuid.UUID, len(messages))
	for i, msg := range messages {
		messageIDs[i] = msg.ID
	}

	// Build a map of message_id -> file
	filesMap := make(map[uuid.UUID]*models.File)
	if len(messageIDs) > 0 {
		// Create placeholders for IN clause
		placeholders := make([]string, len(messageIDs))
		args := make([]interface{}, len(messageIDs))
		for i, id := range messageIDs {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args[i] = id
		}

		filesQuery := fmt.Sprintf(`SELECT id, application_id, message_id, user_id, filename, file_path, file_size, mime_type, uploaded_at, deleted_at
		                         FROM files WHERE message_id IN (%s) AND deleted_at IS NULL`, strings.Join(placeholders, ","))

		fileRows, err := h.db.Query(filesQuery, args...)
		if err == nil {
			defer fileRows.Close()
			for fileRows.Next() {
				fileData := models.File{}
				var fileDeletedAt *time.Time
				if err := fileRows.Scan(
					&fileData.ID,
					&fileData.ApplicationID,
					&fileData.MessageID,
					&fileData.UserID,
					&fileData.Filename,
					&fileData.FilePath,
					&fileData.FileSize,
					&fileData.MimeType,
					&fileData.UploadedAt,
					&fileDeletedAt,
				); err == nil {
					filesMap[fileData.MessageID] = &fileData
				}
			}
		}
	}

	// Create response with messages and their files
	type MessageWithFile struct {
		models.Message
		File *models.File `json:"file,omitempty"`
	}

	messagesWithFiles := make([]MessageWithFile, len(messages))
	for i, msg := range messages {
		messagesWithFiles[i] = MessageWithFile{
			Message: msg,
			File:    filesMap[msg.ID],
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messagesWithFiles,
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

	userIDUUID, err := uuid.Parse(userID)
	if err != nil {
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

	if appID == nil {
		checkQuery = `SELECT m.user_id FROM messages m
		               WHERE m.id = $1 AND m.room_id = $2 AND m.deleted_at IS NULL`
		err = h.db.QueryRow(checkQuery, req.ID, req.RoomID).Scan(&existingUserID)
	} else {
		checkQuery = `SELECT m.user_id FROM messages m
		               INNER JOIN rooms r ON m.room_id = r.id
		               WHERE m.id = $1 AND m.room_id = $2 AND r.application_id = $3 AND m.deleted_at IS NULL`
		err = h.db.QueryRow(checkQuery, req.ID, req.RoomID, *appID).Scan(&existingUserID)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
				"Message not found", nil)
			return
		}
		logger.Error("Failed to check message existence for update", err, logger.Fields{
			"message_id": req.ID,
			"room_id":    req.RoomID,
			"app_id":     appID,
		})
		c.Error(err)
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to check message existence", nil)
		return
	}

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
		logger.Error("Failed to update message", err, logger.Fields{
			"message_id": req.ID,
			"room_id":    req.RoomID,
			"query":      updateQuery,
		})
		c.Error(err)
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to update message", nil)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected after update", err, logger.Fields{
			"message_id": req.ID,
		})
		c.Error(err)
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

	userIDUUID, err := uuid.Parse(userID)
	if err != nil {
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

	if req.ID == uuid.Nil {
		logger.Warn("Delete message request with invalid message ID", logger.Fields{
			"message_id": req.ID,
			"room_id":    req.RoomID,
			"app_id":     req.ApplicationID,
		})
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Message ID is required and must be a valid UUID", nil)
		return
	}

	if req.RoomID == uuid.Nil {
		logger.Warn("Delete message request with invalid room ID", logger.Fields{
			"message_id": req.ID,
			"room_id":    req.RoomID,
			"app_id":     req.ApplicationID,
		})
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Room ID is required and must be a valid UUID", nil)
		return
	}

	appID := req.ApplicationID
	var checkQuery string
	var existingUserID uuid.UUID

	checkQuery = `SELECT m.user_id FROM messages m
	               WHERE m.id = $1 AND m.room_id = $2 AND m.deleted_at IS NULL`
	err = h.db.QueryRow(checkQuery, req.ID, req.RoomID).Scan(&existingUserID)

	if err == nil && appID != nil {
		var roomAppID uuid.UUID
		verifyQuery := `SELECT r.application_id FROM rooms r WHERE r.id = $1`
		if verifyErr := h.db.QueryRow(verifyQuery, req.RoomID).Scan(&roomAppID); verifyErr == nil {
			if roomAppID != *appID {
				errors.RespondWithError(c, http.StatusForbidden, errors.ErrCodeForbidden,
					"Application ID does not match the room's application", nil)
				return
			}
		}
	}

	if err != nil {
		if err == sql.ErrNoRows {
			errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
				"Message not found", nil)
			return
		}
		logger.Error("Failed to check message existence for deletion", err, logger.Fields{
			"message_id": req.ID,
			"room_id":    req.RoomID,
			"app_id":     appID,
		})
		c.Error(err)
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to check message existence", nil)
		return
	}

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
		logger.Error("Failed to delete message", err, logger.Fields{
			"message_id": req.ID,
			"room_id":    req.RoomID,
		})
		c.Error(err)
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to delete message", nil)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected after deletion", err, logger.Fields{
			"message_id": req.ID,
		})
		c.Error(err)
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to delete message", nil)
		return
	}

	if rowsAffected == 0 {
		errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
			"Message not found", nil)
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Panic in WebSocket message deletion broadcast", fmt.Errorf("%v", r), logger.Fields{
					"room_id":    req.RoomID.String(),
					"message_id": req.ID.String(),
				})
			}
		}()

		h.wsHub.Broadcast <- struct {
			RoomID  string
			Message websocket.MessageStruct
		}{
			RoomID: req.RoomID.String(),
			Message: websocket.MessageStruct{
				Type: "message_deleted",
				Payload: map[string]interface{}{
					"message_id": req.ID,
					"room_id":    req.RoomID,
				},
			},
		}
		logger.Debug("Message deletion broadcasted via WebSocket", logger.Fields{
			"room_id":    req.RoomID.String(),
			"message_id": req.ID.String(),
		})
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Message deleted successfully",
	})
}
