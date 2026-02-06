package rest

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
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

type RoomHandler struct {
	ws *websocket.Hub
	db *database.DB
}

func NewRoomHandler(db *database.DB, ws *websocket.Hub) *RoomHandler {
	return &RoomHandler{
		db: db,
		ws: ws,
	}
}

func (h *RoomHandler) Create(c *gin.Context) {
	var req models.RoomCreateRequest
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

	if !validator.ValidateNotEmpty(req.Name) {
		details := []errors.ErrorDetail{
			{
				Field: "name",
				Issue: "Room name is required",
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Room name is required", details)
		return
	}

	if !validator.ValidateLength(req.Name, 1, 255) {
		details := []errors.ErrorDetail{
			{
				Field: "name",
				Issue: "Room name must be between 1 and 255 characters",
				Value: req.Name,
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Room name must be between 1 and 255 characters", details)
		return
	}

	if req.ApplicationID != nil {
		exists, err := helperfunctions.CheckIfApplicationExistsWithId(c, h.db, req.ApplicationID.String())
		if !exists || err != nil {
			errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
				"Application not found", nil)
			return
		}
	}

	roomType := req.Type
	if roomType == "" {
		roomType = models.RoomTypeGroup
	} else if roomType != models.RoomTypeGroup && roomType != models.RoomTypeDirect && roomType != models.RoomTypeChannel {
		details := []errors.ErrorDetail{
			{
				Field: "type",
				Issue: "Invalid room type. Must be 'group', 'direct', or 'channel'",
				Value: string(roomType),
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid room type. Must be 'group', 'direct', or 'channel'", details)
		return
	}

	if req.CreatedBy != nil {
		var checkUserQuery string
		var checkUserArgs []interface{}
		if req.ApplicationID != nil {
			checkUserQuery = `SELECT id FROM users WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
			checkUserArgs = []interface{}{req.CreatedBy, req.ApplicationID}
		} else {
			checkUserQuery = `SELECT id FROM users WHERE id = $1 AND deleted_at IS NULL`
			checkUserArgs = []interface{}{req.CreatedBy}
		}

		var existingUserID uuid.UUID
		err := h.db.QueryRow(checkUserQuery, checkUserArgs...).Scan(&existingUserID)
		if err != nil {
			if err == sql.ErrNoRows {
				errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
					"User not found", nil)
				return
			}
			errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
				"Failed to check user existence", nil)
			return
		}
	}

	var metadataJSON []byte
	var err error
	if req.Metadata != nil {
		metadataJSON, err = json.Marshal(req.Metadata)
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
	}

	now := time.Now()
	var insertQuery string
	var insertArgs []interface{}

	if req.ApplicationID != nil {
		insertQuery = `INSERT INTO rooms (application_id, name, type, description, created_by, metadata, created_at, updated_at)
		               VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		               RETURNING id, created_at, updated_at`
		insertArgs = []interface{}{req.ApplicationID, req.Name, roomType, req.Description, req.CreatedBy, metadataJSON, now, now}
	} else {
		insertQuery = `INSERT INTO rooms (name, type, description, created_by, metadata, created_at, updated_at)
		               VALUES ($1, $2, $3, $4, $5, $6, $7)
		               RETURNING id, created_at, updated_at`
		insertArgs = []interface{}{req.Name, roomType, req.Description, req.CreatedBy, metadataJSON, now, now}
	}

	room := models.Room{}
	err = h.db.QueryRow(insertQuery, insertArgs...).Scan(&room.ID, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to create room", nil)
		return
	}

	var fetchQuery string
	var fetchArgs []interface{}
	if req.ApplicationID != nil {
		fetchQuery = `SELECT id, application_id, name, type, description, created_by, metadata, created_at, updated_at, deleted_at
		              FROM rooms
		              WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
		fetchArgs = []interface{}{room.ID, req.ApplicationID}
	} else {
		fetchQuery = `SELECT id, application_id, name, type, description, created_by, metadata, created_at, updated_at, deleted_at
		              FROM rooms
		              WHERE id = $1 AND deleted_at IS NULL`
		fetchArgs = []interface{}{room.ID}
	}

	var metadataJSONFetch []byte
	err = h.db.QueryRow(fetchQuery, fetchArgs...).Scan(
		&room.ID,
		&room.ApplicationID,
		&room.Name,
		&room.Type,
		&room.Description,
		&room.CreatedBy,
		&metadataJSONFetch,
		&room.CreatedAt,
		&room.UpdatedAt,
		&room.DeletedAt,
	)

	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to retrieve created room", nil)
		return
	}

	if len(metadataJSONFetch) > 0 {
		if err := json.Unmarshal(metadataJSONFetch, &room.Metadata); err != nil {
		}
	}

	// Broadcast room creation event via WebSocket if hub is available
	if h.ws != nil {
		select {
		case h.ws.Broadcast <- struct {
			RoomID  string
			Message websocket.MessageStruct
		}{
			RoomID: room.ID.String(),
			Message: websocket.MessageStruct{
				Type:    "room_created",
				Payload: room,
			},
		}:
		default:
			// Channel is full or closed, log but don't fail the request
			log.Printf("Failed to broadcast room_created event: channel unavailable")
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Room created successfully",
		"room":    room,
	})
}

func (h *RoomHandler) Get(c *gin.Context) {
	roomID := c.Query("id")
	appID := c.Query("application_id")

	if roomID == "" {
		h.List(c)
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
		query = `SELECT id, application_id, name, type, description, created_by, metadata, created_at, updated_at, deleted_at
		         FROM rooms
		         WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
		args = []interface{}{roomIDUUID, appIDUUID}
	} else {
		query = `SELECT id, application_id, name, type, description, created_by, metadata, created_at, updated_at, deleted_at
		         FROM rooms
		         WHERE id = $1 AND deleted_at IS NULL`
		args = []interface{}{roomIDUUID}
	}

	room := models.Room{}
	var metadataJSON []byte

	err = h.db.QueryRow(query, args...).Scan(
		&room.ID,
		&room.ApplicationID,
		&room.Name,
		&room.Type,
		&room.Description,
		&room.CreatedBy,
		&metadataJSON,
		&room.CreatedAt,
		&room.UpdatedAt,
		&room.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
				"Room not found", nil)
			return
		}
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to retrieve room", nil)
		return
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &room.Metadata); err != nil {
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"room": room,
	})
}

func (h *RoomHandler) List(c *gin.Context) {
	appID := c.Query("application_id")

	var query string
	var args []interface{}

	if appID != "" {
		appIDUUID, err := uuid.Parse(appID)
		if err != nil {
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Invalid application ID format", nil)
			return
		}
		query = `SELECT id, application_id, name, type, description, created_by, metadata, created_at, updated_at, deleted_at
		         FROM rooms
		         WHERE application_id = $1 AND deleted_at IS NULL
		         ORDER BY created_at DESC`
		args = []interface{}{appIDUUID}
	} else {
		query = `SELECT id, application_id, name, type, description, created_by, metadata, created_at, updated_at, deleted_at
		         FROM rooms
		         WHERE deleted_at IS NULL
		         ORDER BY created_at DESC`
		args = []interface{}{}
	}

	rows, err := h.db.Query(query, args...)
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to retrieve rooms", nil)
		return
	}
	defer rows.Close()

	var rooms []models.Room
	for rows.Next() {
		room := models.Room{}
		var metadataJSON []byte

		err := rows.Scan(
			&room.ID,
			&room.ApplicationID,
			&room.Name,
			&room.Type,
			&room.Description,
			&room.CreatedBy,
			&metadataJSON,
			&room.CreatedAt,
			&room.UpdatedAt,
			&room.DeletedAt,
		)
		if err != nil {
			errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
				"Failed to scan room", nil)
			return
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &room.Metadata); err != nil {
			}
		}

		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to retrieve rooms", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rooms": rooms,
		"count": len(rooms),
	})
}

func (h *RoomHandler) Update(c *gin.Context) {
	var req models.RoomUpdateRequest
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

	var appIDUUID *uuid.UUID
	if req.ApplicationID != nil {
		appIDUUID = req.ApplicationID
	}

	if req.Name != "" {
		if !validator.ValidateLength(req.Name, 1, 255) {
			details := []errors.ErrorDetail{
				{
					Field: "name",
					Issue: "Room name must be between 1 and 255 characters",
					Value: req.Name,
				},
			}
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Room name must be between 1 and 255 characters", details)
			return
		}
	}

	var checkQuery string
	var checkArgs []interface{}
	if appIDUUID != nil {
		checkQuery = `SELECT id FROM rooms WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
		checkArgs = []interface{}{req.ID, *appIDUUID}
	} else {
		checkQuery = `SELECT id FROM rooms WHERE id = $1 AND deleted_at IS NULL`
		checkArgs = []interface{}{req.ID}
	}

	var existingID uuid.UUID
	err := h.db.QueryRow(checkQuery, checkArgs...).Scan(&existingID)
	if err != nil {
		if err == sql.ErrNoRows {
			errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
				"Room not found", nil)
			return
		}
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to check room existence", nil)
		return
	}

	updateFields := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Name != "" {
		updateFields = append(updateFields, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, req.Name)
		argIndex++
	}

	if req.Description != nil {
		updateFields = append(updateFields, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *req.Description)
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

	updateFields = append(updateFields, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	var whereClause string
	if appIDUUID != nil {
		whereClause = fmt.Sprintf("WHERE id = $%d AND application_id = $%d AND deleted_at IS NULL", argIndex, argIndex+1)
		args = append(args, req.ID, *appIDUUID)
	} else {
		whereClause = fmt.Sprintf("WHERE id = $%d AND deleted_at IS NULL", argIndex)
		args = append(args, req.ID)
	}

	updateQuery := fmt.Sprintf("UPDATE rooms SET %s %s", strings.Join(updateFields, ", "), whereClause)

	result, err := h.db.Exec(updateQuery, args...)
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to update room", nil)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to update room", nil)
		return
	}

	if rowsAffected == 0 {
		errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
			"Room not found", nil)
		return
	}

	// Fetch updated room
	var fetchQuery string
	var fetchArgs []interface{}
	if appIDUUID != nil {
		fetchQuery = `SELECT id, application_id, name, type, description, created_by, metadata, created_at, updated_at, deleted_at
		              FROM rooms
		              WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
		fetchArgs = []interface{}{req.ID, *appIDUUID}
	} else {
		fetchQuery = `SELECT id, application_id, name, type, description, created_by, metadata, created_at, updated_at, deleted_at
		              FROM rooms
		              WHERE id = $1 AND deleted_at IS NULL`
		fetchArgs = []interface{}{req.ID}
	}

	updatedRoom := models.Room{}
	var metadataJSONScan []byte
	err = h.db.QueryRow(fetchQuery, fetchArgs...).Scan(
		&updatedRoom.ID,
		&updatedRoom.ApplicationID,
		&updatedRoom.Name,
		&updatedRoom.Type,
		&updatedRoom.Description,
		&updatedRoom.CreatedBy,
		&metadataJSONScan,
		&updatedRoom.CreatedAt,
		&updatedRoom.UpdatedAt,
		&updatedRoom.DeletedAt,
	)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Room updated successfully",
		})
		return
	}

	if len(metadataJSONScan) > 0 {
		if err := json.Unmarshal(metadataJSONScan, &updatedRoom.Metadata); err != nil {
		}
	}

	go func() {
		if h.ws != nil {
			select {
			case h.ws.Broadcast <- struct {
				RoomID  string
				Message websocket.MessageStruct
			}{
				RoomID: updatedRoom.ID.String(),
				Message: websocket.MessageStruct{
					Type:    "room_updated",
					Payload: updatedRoom,
				},
			}:
			default:
				log.Printf("Failed to broadcast room_updated event: channel unavailable")
			}
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Room updated successfully",
		"room":    updatedRoom,
	})
}

func (h *RoomHandler) Delete(c *gin.Context) {
	var req models.RoomDeleteRequest
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

	var appIDUUID *uuid.UUID
	if req.ApplicationID != nil {
		appIDUUID = req.ApplicationID
	}

	var checkQuery string
	var checkArgs []interface{}
	if appIDUUID != nil {
		checkQuery = `SELECT id FROM rooms WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
		checkArgs = []interface{}{req.ID, *appIDUUID}
	} else {
		checkQuery = `SELECT id FROM rooms WHERE id = $1 AND deleted_at IS NULL`
		checkArgs = []interface{}{req.ID}
	}

	var existingID uuid.UUID
	err := h.db.QueryRow(checkQuery, checkArgs...).Scan(&existingID)
	if err != nil {
		if err == sql.ErrNoRows {
			errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
				"Room not found", nil)
			return
		}
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to check room existence", nil)
		return
	}

	now := time.Now()
	var deleteQuery string
	var deleteArgs []interface{}

	if appIDUUID != nil {
		deleteQuery = `UPDATE rooms SET deleted_at = $1, updated_at = $2 WHERE id = $3 AND application_id = $4 AND deleted_at IS NULL`
		deleteArgs = []interface{}{now, now, req.ID, *appIDUUID}
	} else {
		deleteQuery = `UPDATE rooms SET deleted_at = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`
		deleteArgs = []interface{}{now, now, req.ID}
	}

	result, err := h.db.Exec(deleteQuery, deleteArgs...)
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to delete room", nil)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to delete room", nil)
		return
	}

	if rowsAffected == 0 {
		errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
			"Room not found", nil)
		return
	}

	go func() {
		if h.ws != nil {
			select {
			case h.ws.Broadcast <- struct {
				RoomID  string
				Message websocket.MessageStruct
			}{
				RoomID: req.ID.String(),
				Message: websocket.MessageStruct{
					Type:    "room_updated",
					Payload: req,
				},
			}:
			default:
				log.Printf("Failed to broadcast room_updated event: channel unavailable")
			}
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Room deleted successfully",
	})
}

func (h *RoomHandler) AddMember(c *gin.Context) {
	var req models.RoomAddMemberRequest
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

	var appIDUUID *uuid.UUID
	if req.ApplicationID != nil {
		appIDUUID = req.ApplicationID
	}

	var checkRoomQuery string
	var checkRoomArgs []interface{}
	if appIDUUID != nil {
		checkRoomQuery = `SELECT id FROM rooms WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
		checkRoomArgs = []interface{}{req.RoomID, *appIDUUID}
	} else {
		checkRoomQuery = `SELECT id FROM rooms WHERE id = $1 AND deleted_at IS NULL`
		checkRoomArgs = []interface{}{req.RoomID}
	}

	var existingRoomID uuid.UUID
	err := h.db.QueryRow(checkRoomQuery, checkRoomArgs...).Scan(&existingRoomID)
	if err != nil {
		if err == sql.ErrNoRows {
			errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
				"Room not found", nil)
			return
		}
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to check room existence", nil)
		return
	}

	var checkUserQuery string
	var checkUserArgs []interface{}
	var userName string
	if appIDUUID != nil {
		checkUserQuery = `SELECT id,username FROM users WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
		checkUserArgs = []interface{}{req.UserID, *appIDUUID}
	} else {
		checkUserQuery = `SELECT id,username FROM users WHERE id = $1 AND deleted_at IS NULL`
		checkUserArgs = []interface{}{req.UserID}
	}

	var existingUserID uuid.UUID
	err = h.db.QueryRow(checkUserQuery, checkUserArgs...).Scan(&existingUserID, &userName)
	if err != nil {
		if err == sql.ErrNoRows {
			errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
				"User not found", nil)
			return
		}
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to check user existence", nil)
		return
	}

	var existingMemberID uuid.UUID
	checkMemberQuery := `SELECT id FROM room_members WHERE room_id = $1 AND user_id = $2`
	err = h.db.QueryRow(checkMemberQuery, req.RoomID, req.UserID).Scan(&existingMemberID)
	if err == nil {
		errors.RespondWithError(c, http.StatusConflict, errors.ErrCodeConflict,
			"User is already a member of this room", nil)
		return
	} else if err != sql.ErrNoRows {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to check member existence", nil)
		return
	}

	role := req.Role
	if role == "" {
		role = "member"
	}

	now := time.Now()

	insertQuery := `INSERT INTO room_members (room_id, user_id, role, joined_at)
	                VALUES ($1, $2, $3, $4)
	                RETURNING id, joined_at`

	member := models.RoomMember{}
	err = h.db.QueryRow(insertQuery, req.RoomID, req.UserID, role, now).Scan(&member.ID, &member.JoinedAt)
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to add member to room", nil)
		return
	}

	member.RoomID = req.RoomID
	member.UserID = req.UserID
	member.Role = role

	go func() {
		room, err := helperfunctions.GetRoomById(c, h.db, req.RoomID.String())
		if err != nil {
			log.Printf("Error getting room by ID: %v", err)
			return
		}

		roomWithMembers := map[string]interface{}{
			"room":   room,
			"member": member,
		}

		if h.ws != nil {
			select {
			case h.ws.Broadcast <- struct {
				RoomID  string
				Message websocket.MessageStruct
			}{
				RoomID: req.RoomID.String(),
				Message: websocket.MessageStruct{
					Type:    "room_member_added",
					Payload: roomWithMembers,
				},
			}:
			default:
				log.Printf("Failed to broadcast room_member_added event: channel unavailable")
			}
		}
	}()

	c.JSON(http.StatusCreated, gin.H{
		"message": "Member added to room successfully",
		"member":  member,
	})
}

func (h *RoomHandler) RemoveMember(c *gin.Context) {
	var req models.RoomRemoveMemberRequest
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

	var appIDUUID *uuid.UUID
	if req.ApplicationID != nil {
		appIDUUID = req.ApplicationID
	}

	var checkRoomQuery string
	var checkRoomArgs []interface{}
	if appIDUUID != nil {
		checkRoomQuery = `SELECT id FROM rooms WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
		checkRoomArgs = []interface{}{req.RoomID, *appIDUUID}
	} else {
		checkRoomQuery = `SELECT id FROM rooms WHERE id = $1 AND deleted_at IS NULL`
		checkRoomArgs = []interface{}{req.RoomID}
	}

	var existingRoomID uuid.UUID
	err := h.db.QueryRow(checkRoomQuery, checkRoomArgs...).Scan(&existingRoomID)
	if err != nil {
		if err == sql.ErrNoRows {
			errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
				"Room not found", nil)
			return
		}
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to check room existence", nil)
		return
	}

	var existingMemberID uuid.UUID
	checkMemberQuery := `SELECT id FROM room_members WHERE room_id = $1 AND user_id = $2`
	err = h.db.QueryRow(checkMemberQuery, req.RoomID, req.UserID).Scan(&existingMemberID)
	if err != nil {
		if err == sql.ErrNoRows {
			errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
				"User is not a member of this room", nil)
			return
		}
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to check member existence", nil)
		return
	}

	deleteQuery := `DELETE FROM room_members WHERE room_id = $1 AND user_id = $2`
	result, err := h.db.Exec(deleteQuery, req.RoomID, req.UserID)
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to remove member from room", nil)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to remove member from room", nil)
		return
	}

	if rowsAffected == 0 {
		errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
			"User is not a member of this room", nil)
		return
	}

	go func() {
		room, err := helperfunctions.GetRoomById(c, h.db, req.RoomID.String())
		if err != nil {
			log.Printf("Error getting room by ID: %v", err)
			return
		}

		user, err := helperfunctions.GetUserById(c, h.db, req.UserID.String())
		if err != nil {
			log.Printf("Error getting user by ID: %v", err)
			return
		}

		roomWithMembers := map[string]interface{}{
			"room":          room,
			"memberRemoved": user,
		}

		if h.ws != nil {
			select {
			case h.ws.Broadcast <- struct {
				RoomID  string
				Message websocket.MessageStruct
			}{
				RoomID: req.RoomID.String(),
				Message: websocket.MessageStruct{
					Type:    "room_member_removed",
					Payload: roomWithMembers,
				},
			}:
			default:
				log.Printf("Failed to broadcast room_member_removed event: channel unavailable")
			}
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Member removed from room successfully",
	})
}

func (h *RoomHandler) ListMembers(c *gin.Context) {
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

	var checkRoomQuery string
	var checkRoomArgs []interface{}
	if appID != "" {
		appIDUUID, err := uuid.Parse(appID)
		if err != nil {
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Invalid application ID format", nil)
			return
		}
		checkRoomQuery = `SELECT id FROM rooms WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
		checkRoomArgs = []interface{}{roomIDUUID, appIDUUID}
	} else {
		checkRoomQuery = `SELECT id FROM rooms WHERE id = $1 AND deleted_at IS NULL`
		checkRoomArgs = []interface{}{roomIDUUID}
	}

	var existingRoomID uuid.UUID
	err = h.db.QueryRow(checkRoomQuery, checkRoomArgs...).Scan(&existingRoomID)
	if err != nil {
		if err == sql.ErrNoRows {
			errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
				"Room not found", nil)
			return
		}
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to check room existence", nil)
		return
	}

	query := `SELECT rm.id, rm.room_id, rm.user_id, u.username, rm.role, rm.joined_at, rm.last_read_at
	          FROM room_members rm
	          JOIN users u ON rm.user_id = u.id
	          WHERE rm.room_id = $1
	          ORDER BY rm.joined_at ASC`

	rows, err := h.db.Query(query, roomIDUUID)
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to retrieve room members", nil)
		return
	}
	defer rows.Close()

	var members []models.RoomMember
	for rows.Next() {
		member := models.RoomMember{}
		err := rows.Scan(
			&member.ID,
			&member.RoomID,
			&member.UserID,
			&member.Username,
			&member.Role,
			&member.JoinedAt,
			&member.LastReadAt,
		)
		if err != nil {
			errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
				"Failed to scan member", nil)
			return
		}

		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to retrieve room members", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"members": members,
		"count":   len(members),
	})
}
