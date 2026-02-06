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

	userIDUUID, err := uuid.Parse(userID)
	if err != nil {
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

	reaction := strings.TrimSpace(req.Reaction)
	if reaction == "" {
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Reaction cannot be empty", nil)
		return
	}

	if len(reaction) > 10 {
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Reaction must be 10 characters or less", nil)
		return
	}

	var appIDUUID *uuid.UUID
	if req.ApplicationID != nil {
		appIDUUID = req.ApplicationID
	}

	var messageQuery string
	var messageArgs []interface{}
	if appIDUUID != nil {
		messageQuery = `SELECT m.id, m.room_id, m.user_id 
		                FROM messages m
		                INNER JOIN rooms r ON m.room_id = r.id
		                WHERE m.id = $1 AND r.application_id = $2 AND m.deleted_at IS NULL`
		messageArgs = []interface{}{req.MessageID, *appIDUUID}
	} else {
		messageQuery = `SELECT id, room_id, user_id 
		                FROM messages 
		                WHERE id = $1 AND deleted_at IS NULL`
		messageArgs = []interface{}{req.MessageID}
	}

	var messageID, roomID, messageUserID uuid.UUID
	err = h.db.QueryRow(messageQuery, messageArgs...).Scan(&messageID, &roomID, &messageUserID)
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

	isMember, err := helperfunctions.ValidateUserIsMemberOfRoom(h.db, userIDUUID.String(), roomID.String())
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

	var existingReactionID uuid.UUID
	checkQuery := `SELECT id FROM message_reactions 
	               WHERE message_id = $1 AND user_id = $2 AND reaction = $3`
	err = h.db.QueryRow(checkQuery, req.MessageID, userIDUUID, reaction).Scan(&existingReactionID)
	if err == nil {
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

		if h.wsHub != nil {
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
			case <-time.After(5 * time.Second):
			}
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

	userIDUUID, err := uuid.Parse(userID)
	if err != nil {
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

	reaction := strings.TrimSpace(req.Reaction)
	if reaction == "" {
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Reaction cannot be empty", nil)
		return
	}

	var appIDUUID *uuid.UUID
	if req.ApplicationID != nil {
		appIDUUID = req.ApplicationID
	}

	var messageQuery string
	var messageArgs []interface{}
	if appIDUUID != nil {
		messageQuery = `SELECT m.id, m.room_id 
		                FROM messages m
		                INNER JOIN rooms r ON m.room_id = r.id
		                WHERE m.id = $1 AND r.application_id = $2 AND m.deleted_at IS NULL`
		messageArgs = []interface{}{req.MessageID, *appIDUUID}
	} else {
		messageQuery = `SELECT id, room_id 
		                FROM messages 
		                WHERE id = $1 AND deleted_at IS NULL`
		messageArgs = []interface{}{req.MessageID}
	}

	var messageID, roomID uuid.UUID
	err = h.db.QueryRow(messageQuery, messageArgs...).Scan(&messageID, &roomID)
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

	isMember, err := helperfunctions.ValidateUserIsMemberOfRoom(h.db, userIDUUID.String(), roomID.String())
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

		if h.wsHub != nil {
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
			case <-time.After(5 * time.Second):
			}
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

	var messageQuery string
	var messageArgs []interface{}
	if appID != "" {
		appIDUUID, err := uuid.Parse(appID)
		if err != nil {
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Invalid application ID format", nil)
			return
		}
		messageQuery = `SELECT m.id 
		                FROM messages m
		                INNER JOIN rooms r ON m.room_id = r.id
		                WHERE m.id = $1 AND r.application_id = $2 AND m.deleted_at IS NULL`
		messageArgs = []interface{}{messageIDUUID, appIDUUID}
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
