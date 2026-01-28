package rest

import (
	"database/sql"
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
	"chat-app/internal/utils/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReactionHandler struct {
	db    *database.DB
	wsHub *websocket.Hub
}

func NewReactionHandler(db *database.DB, wsHub *websocket.Hub) *ReactionHandler {
	return &ReactionHandler{
		db:    db,
		wsHub: wsHub,
	}
}

func (h *ReactionHandler) Create(c *gin.Context) {
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

	var req models.ReactionCreateRequest
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

	// Validate reaction string
	reaction := strings.TrimSpace(req.Reaction)
	if reaction == "" {
		details := []errors.ErrorDetail{
			{
				Field: "reaction",
				Issue: "Reaction cannot be empty",
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Reaction cannot be empty", details)
		return
	}

	if len(reaction) > 10 {
		details := []errors.ErrorDetail{
			{
				Field: "reaction",
				Issue: "Reaction must be 10 characters or less",
				Value: reaction,
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Reaction must be 10 characters or less", details)
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

	// Validate message exists and get room_id
	var messageQuery string
	var messageArgs []interface{}
	if appID != "" {
		messageQuery = `SELECT m.id, m.room_id, m.user_id 
		                FROM messages m
		                INNER JOIN rooms r ON m.room_id = r.id
		                WHERE m.id = $1 AND r.application_id = $2 AND m.deleted_at IS NULL`
		messageArgs = []interface{}{req.MessageID, req.ApplicationID}
	} else {
		messageQuery = `SELECT id, room_id, user_id 
		                FROM messages 
		                WHERE id = $1 AND deleted_at IS NULL`
		messageArgs = []interface{}{req.MessageID}
	}

	var messageID, roomID, messageUserID uuid.UUID
	err := h.db.QueryRow(messageQuery, messageArgs...).Scan(&messageID, &roomID, &messageUserID)
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

	// Validate user is member of room
	userIDUUID := uuid.MustParse(userID)
	isMember, err := helperfunctions.ValidateUserIsMemberOfRoom(h.db, userID, roomID.String())
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

	// Check if reaction already exists (unique constraint: message_id, user_id, reaction)
	var existingReactionID uuid.UUID
	checkQuery := `SELECT id FROM message_reactions 
	               WHERE message_id = $1 AND user_id = $2 AND reaction = $3`
	err = h.db.QueryRow(checkQuery, req.MessageID, userIDUUID, reaction).Scan(&existingReactionID)
	if err == nil {
		// Reaction already exists, return it
		fetchQuery := `SELECT id, message_id, user_id, reaction, created_at 
		               FROM message_reactions 
		               WHERE id = $1`
		reactionModel := models.MessageReaction{}
		err = h.db.QueryRow(fetchQuery, existingReactionID).Scan(
			&reactionModel.ID,
			&reactionModel.MessageID,
			&reactionModel.UserID,
			&reactionModel.Reaction,
			&reactionModel.CreatedAt,
		)
		if err != nil {
			errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
				"Failed to retrieve reaction", nil)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "Reaction already exists",
			"reaction": reactionModel,
		})
		return
	} else if err != sql.ErrNoRows {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to check reaction existence", nil)
		return
	}

	// Create reaction
	insertQuery := `INSERT INTO message_reactions (message_id, user_id, reaction, created_at)
	                VALUES ($1, $2, $3, $4)
	                RETURNING id, created_at`

	now := time.Now()
	reactionModel := models.MessageReaction{
		MessageID: req.MessageID,
		UserID:    userIDUUID,
		Reaction:  reaction,
	}

	err = h.db.QueryRow(insertQuery, req.MessageID, userIDUUID, reaction, now).Scan(
		&reactionModel.ID,
		&reactionModel.CreatedAt,
	)
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to create reaction", nil)
		return
	}

	// Broadcast reaction via websocket
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Panic in WebSocket reaction broadcast", fmt.Errorf("%v", r), logger.Fields{
					"room_id":    roomID.String(),
					"message_id": req.MessageID.String(),
					"reaction":   reaction,
				})
			}
		}()

		select {
		case h.wsHub.Broadcast <- struct {
			RoomID  string
			Message websocket.MessageStruct
		}{
			RoomID: roomID.String(),
			Message: websocket.MessageStruct{
				Type:    "reaction_added",
				Payload: reactionModel,
			},
		}:
			logger.Debug("Reaction broadcasted via WebSocket", logger.Fields{
				"room_id":    roomID.String(),
				"message_id": req.MessageID.String(),
				"reaction":   reaction,
			})
		case <-time.After(5 * time.Second):
			logger.Warn("WebSocket reaction broadcast timeout", logger.Fields{
				"room_id":    roomID.String(),
				"message_id": req.MessageID.String(),
			})
		}
	}()

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Reaction created successfully",
		"reaction": reactionModel,
	})
}

func (h *ReactionHandler) Delete(c *gin.Context) {
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

	var req models.ReactionDeleteRequest
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

	// Validate reaction string
	reaction := strings.TrimSpace(req.Reaction)
	if reaction == "" {
		details := []errors.ErrorDetail{
			{
				Field: "reaction",
				Issue: "Reaction cannot be empty",
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Reaction cannot be empty", details)
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

	// Validate message exists and get room_id
	var messageQuery string
	var messageArgs []interface{}
	if appID != "" {
		messageQuery = `SELECT m.id, m.room_id 
		                FROM messages m
		                INNER JOIN rooms r ON m.room_id = r.id
		                WHERE m.id = $1 AND r.application_id = $2 AND m.deleted_at IS NULL`
		messageArgs = []interface{}{req.MessageID, req.ApplicationID}
	} else {
		messageQuery = `SELECT id, room_id 
		                FROM messages 
		                WHERE id = $1 AND deleted_at IS NULL`
		messageArgs = []interface{}{req.MessageID}
	}

	var messageID, roomID uuid.UUID
	err := h.db.QueryRow(messageQuery, messageArgs...).Scan(&messageID, &roomID)
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

	// Validate user is member of room
	userIDUUID := uuid.MustParse(userID)
	isMember, err := helperfunctions.ValidateUserIsMemberOfRoom(h.db, userID, roomID.String())
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

	// Check if reaction exists and belongs to user
	var existingReactionID uuid.UUID
	checkQuery := `SELECT id FROM message_reactions 
	               WHERE message_id = $1 AND user_id = $2 AND reaction = $3`
	err = h.db.QueryRow(checkQuery, req.MessageID, userIDUUID, reaction).Scan(&existingReactionID)
	if err != nil {
		if err == sql.ErrNoRows {
			errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
				"Reaction not found", nil)
			return
		}
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to check reaction existence", nil)
		return
	}

	// Delete reaction
	deleteQuery := `DELETE FROM message_reactions 
	                WHERE message_id = $1 AND user_id = $2 AND reaction = $3`
	result, err := h.db.Exec(deleteQuery, req.MessageID, userIDUUID, reaction)
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to delete reaction", nil)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to delete reaction", nil)
		return
	}

	if rowsAffected == 0 {
		errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
			"Reaction not found", nil)
		return
	}

	// Broadcast reaction deletion via websocket
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Panic in WebSocket reaction deletion broadcast", fmt.Errorf("%v", r), logger.Fields{
					"room_id":    roomID.String(),
					"message_id": req.MessageID.String(),
					"reaction":   reaction,
				})
			}
		}()

		select {
		case h.wsHub.Broadcast <- struct {
			RoomID  string
			Message websocket.MessageStruct
		}{
			RoomID: roomID.String(),
			Message: websocket.MessageStruct{
				Type: "reaction_removed",
				Payload: map[string]interface{}{
					"message_id": req.MessageID,
					"user_id":    userIDUUID,
					"reaction":   reaction,
				},
			},
		}:
			logger.Debug("Reaction deletion broadcasted via WebSocket", logger.Fields{
				"room_id":    roomID.String(),
				"message_id": req.MessageID.String(),
				"reaction":   reaction,
			})
		case <-time.After(5 * time.Second):
			logger.Warn("WebSocket reaction deletion broadcast timeout", logger.Fields{
				"room_id":    roomID.String(),
				"message_id": req.MessageID.String(),
			})
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Reaction deleted successfully",
	})
}

func (h *ReactionHandler) List(c *gin.Context) {
	messageID := c.Query("message_id")
	appID := c.Query("application_id")

	if messageID == "" {
		details := []errors.ErrorDetail{
			{
				Field: "message_id",
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
				Field: "message_id",
				Issue: "Invalid UUID format",
				Value: messageID,
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid message ID format", details)
		return
	}

	messageIDUUID, err := uuid.Parse(messageID)
	if err != nil {
		details := []errors.ErrorDetail{
			{
				Field: "message_id",
				Issue: "Invalid UUID format",
				Value: messageID,
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid message ID format", details)
		return
	}

	// Validate message exists
	var messageQuery string
	var messageArgs []interface{}
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
		messageQuery = `SELECT m.id 
		                FROM messages m
		                INNER JOIN rooms r ON m.room_id = r.id
		                WHERE m.id = $1 AND r.application_id = $2 AND m.deleted_at IS NULL`
		messageArgs = []interface{}{messageIDUUID, appID}
	} else {
		messageQuery = `SELECT id 
		                FROM messages 
		                WHERE id = $1 AND deleted_at IS NULL`
		messageArgs = []interface{}{messageIDUUID}
	}

	var existingMessageID uuid.UUID
	err = h.db.QueryRow(messageQuery, messageArgs...).Scan(&existingMessageID)
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

	// Fetch all reactions for the message
	query := `SELECT id, message_id, user_id, reaction, created_at 
	          FROM message_reactions 
	          WHERE message_id = $1 
	          ORDER BY created_at ASC`

	rows, err := h.db.Query(query, messageIDUUID)
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to retrieve reactions", nil)
		return
	}
	defer rows.Close()

	var reactions []models.MessageReaction
	for rows.Next() {
		reaction := models.MessageReaction{}
		err := rows.Scan(
			&reaction.ID,
			&reaction.MessageID,
			&reaction.UserID,
			&reaction.Reaction,
			&reaction.CreatedAt,
		)
		if err != nil {
			errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
				"Failed to scan reaction", nil)
			return
		}

		reactions = append(reactions, reaction)
	}

	if err = rows.Err(); err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to retrieve reactions", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reactions": reactions,
		"count":     len(reactions),
	})
}
