package rest

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"chat-app/internal/database"
	"chat-app/internal/models"
	"chat-app/internal/utils/helperfunctions"
	"chat-app/internal/utils/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RoomHandler struct {
	db *database.DB
}

func NewRoomHandler(db *database.DB) *RoomHandler {
	return &RoomHandler{db: db}
}

func (h *RoomHandler) Create(c *gin.Context) {
	var req models.RoomCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	if !validator.ValidateNotEmpty(req.Name) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Room name is required",
			},
		})
		return
	}

	if !validator.ValidateLength(req.Name, 1, 255) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Room name must be between 1 and 255 characters",
			},
		})
		return
	}

	appID := ""
	if req.ApplicationID != nil {
		appID = req.ApplicationID.String()
		if !validator.ValidateUUID(appID) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "Invalid application ID",
				},
			})
			return
		}

		exists, err := helperfunctions.CheckIfApplicationExistsWithId(h.db, appID)
		if !exists || err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "APPLICATION_NOT_FOUND",
					"message": "Application not found",
				},
			})
			return
		}
	}

	roomType := req.Type
	if roomType == "" {
		roomType = models.RoomTypeGroup
	} else if roomType != models.RoomTypeGroup && roomType != models.RoomTypeDirect && roomType != models.RoomTypeChannel {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid room type. Must be 'group', 'direct', or 'channel'",
			},
		})
		return
	}

	if req.CreatedBy != nil {
		var checkUserQuery string
		var checkUserArgs []interface{}
		if appID != "" {
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
				c.JSON(http.StatusNotFound, gin.H{
					"error": gin.H{
						"code":    "USER_NOT_FOUND",
						"message": "User not found",
					},
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "Failed to check user existence",
				},
			})
			return
		}
	}

	var metadataJSON []byte
	var err error
	if req.Metadata != nil {
		metadataJSON, err = json.Marshal(req.Metadata)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "Invalid metadata format",
				},
			})
			return
		}
	}

	now := time.Now()
	var insertQuery string
	var insertArgs []interface{}

	if appID != "" {
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to create room",
			},
		})
		return
	}

	var fetchQuery string
	var fetchArgs []interface{}
	if appID != "" {
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to retrieve created room",
			},
		})
		return
	}

	if len(metadataJSONFetch) > 0 {
		if err := json.Unmarshal(metadataJSONFetch, &room.Metadata); err != nil {
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

	if !validator.ValidateUUID(roomID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid room ID format",
			},
		})
		return
	}

	var query string
	var args []interface{}

	if appID != "" {
		if !validator.ValidateUUID(appID) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "Invalid application ID format",
				},
			})
			return
		}
		query = `SELECT id, application_id, name, type, description, created_by, metadata, created_at, updated_at, deleted_at
		         FROM rooms
		         WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
		args = []interface{}{roomID, appID}
	} else {
		query = `SELECT id, application_id, name, type, description, created_by, metadata, created_at, updated_at, deleted_at
		         FROM rooms
		         WHERE id = $1 AND deleted_at IS NULL`
		args = []interface{}{roomID}
	}

	room := models.Room{}
	var metadataJSON []byte

	err := h.db.QueryRow(query, args...).Scan(
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
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ROOM_NOT_FOUND",
					"message": "Room not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to retrieve room",
			},
		})
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
		if !validator.ValidateUUID(appID) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "Invalid application ID format",
				},
			})
			return
		}
		query = `SELECT id, application_id, name, type, description, created_by, metadata, created_at, updated_at, deleted_at
		         FROM rooms
		         WHERE application_id = $1 AND deleted_at IS NULL
		         ORDER BY created_at DESC`
		args = []interface{}{appID}
	} else {
		query = `SELECT id, application_id, name, type, description, created_by, metadata, created_at, updated_at, deleted_at
		         FROM rooms
		         WHERE deleted_at IS NULL
		         ORDER BY created_at DESC`
		args = []interface{}{}
	}

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to retrieve rooms",
			},
		})
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
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "Failed to scan room",
				},
			})
			return
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &room.Metadata); err != nil {
			}
		}

		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to retrieve rooms",
			},
		})
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	appID := ""
	if req.ApplicationID != nil {
		appID = req.ApplicationID.String()
		if !validator.ValidateUUID(appID) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "Invalid application ID",
				},
			})
			return
		}
	}

	if req.Name != "" {
		if !validator.ValidateLength(req.Name, 1, 255) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "Room name must be between 1 and 255 characters",
				},
			})
			return
		}
	}

	var checkQuery string
	var checkArgs []interface{}
	if appID != "" {
		checkQuery = `SELECT id FROM rooms WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
		checkArgs = []interface{}{req.ID, req.ApplicationID}
	} else {
		checkQuery = `SELECT id FROM rooms WHERE id = $1 AND deleted_at IS NULL`
		checkArgs = []interface{}{req.ID}
	}

	var existingID uuid.UUID
	err := h.db.QueryRow(checkQuery, checkArgs...).Scan(&existingID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ROOM_NOT_FOUND",
					"message": "Room not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to check room existence",
			},
		})
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

	updateFields = append(updateFields, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	var whereClause string
	if appID != "" {
		whereClause = fmt.Sprintf("WHERE id = $%d AND application_id = $%d AND deleted_at IS NULL", argIndex, argIndex+1)
		args = append(args, req.ID, req.ApplicationID)
	} else {
		whereClause = fmt.Sprintf("WHERE id = $%d AND deleted_at IS NULL", argIndex)
		args = append(args, req.ID)
	}

	updateQuery := fmt.Sprintf("UPDATE rooms SET %s %s", strings.Join(updateFields, ", "), whereClause)

	result, err := h.db.Exec(updateQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to update room",
			},
		})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to update room",
			},
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "ROOM_NOT_FOUND",
				"message": "Room not found",
			},
		})
		return
	}

	// Fetch updated room
	var fetchQuery string
	var fetchArgs []interface{}
	if appID != "" {
		fetchQuery = `SELECT id, application_id, name, type, description, created_by, metadata, created_at, updated_at, deleted_at
		              FROM rooms
		              WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
		fetchArgs = []interface{}{req.ID, req.ApplicationID}
	} else {
		fetchQuery = `SELECT id, application_id, name, type, description, created_by, metadata, created_at, updated_at, deleted_at
		              FROM rooms
		              WHERE id = $1 AND deleted_at IS NULL`
		fetchArgs = []interface{}{req.ID}
	}

	updatedRoom := models.Room{}
	var metadataJSON []byte
	err = h.db.QueryRow(fetchQuery, fetchArgs...).Scan(
		&updatedRoom.ID,
		&updatedRoom.ApplicationID,
		&updatedRoom.Name,
		&updatedRoom.Type,
		&updatedRoom.Description,
		&updatedRoom.CreatedBy,
		&metadataJSON,
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

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &updatedRoom.Metadata); err != nil {
			// Metadata parsing error is not critical
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Room updated successfully",
		"room":    updatedRoom,
	})
}

func (h *RoomHandler) Delete(c *gin.Context) {
	var req models.RoomDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	appID := ""
	if req.ApplicationID != nil {
		appID = req.ApplicationID.String()
		if !validator.ValidateUUID(appID) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "Invalid application ID",
				},
			})
			return
		}
	}

	var checkQuery string
	var checkArgs []interface{}
	if appID != "" {
		checkQuery = `SELECT id FROM rooms WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
		checkArgs = []interface{}{req.ID, req.ApplicationID}
	} else {
		checkQuery = `SELECT id FROM rooms WHERE id = $1 AND deleted_at IS NULL`
		checkArgs = []interface{}{req.ID}
	}

	var existingID uuid.UUID
	err := h.db.QueryRow(checkQuery, checkArgs...).Scan(&existingID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ROOM_NOT_FOUND",
					"message": "Room not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to check room existence",
			},
		})
		return
	}

	now := time.Now()
	var deleteQuery string
	var deleteArgs []interface{}

	if appID != "" {
		deleteQuery = `UPDATE rooms SET deleted_at = $1, updated_at = $2 WHERE id = $3 AND application_id = $4 AND deleted_at IS NULL`
		deleteArgs = []interface{}{now, now, req.ID, req.ApplicationID}
	} else {
		deleteQuery = `UPDATE rooms SET deleted_at = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`
		deleteArgs = []interface{}{now, now, req.ID}
	}

	result, err := h.db.Exec(deleteQuery, deleteArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to delete room",
			},
		})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to delete room",
			},
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "ROOM_NOT_FOUND",
				"message": "Room not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Room deleted successfully",
	})
}

func (h *RoomHandler) AddMember(c *gin.Context) {
	var req models.RoomAddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	appID := ""
	if req.ApplicationID != nil {
		appID = req.ApplicationID.String()
		if !validator.ValidateUUID(appID) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "Invalid application ID",
				},
			})
			return
		}
	}

	var checkRoomQuery string
	var checkRoomArgs []interface{}
	if appID != "" {
		checkRoomQuery = `SELECT id FROM rooms WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
		checkRoomArgs = []interface{}{req.RoomID, req.ApplicationID}
	} else {
		checkRoomQuery = `SELECT id FROM rooms WHERE id = $1 AND deleted_at IS NULL`
		checkRoomArgs = []interface{}{req.RoomID}
	}

	var existingRoomID uuid.UUID
	err := h.db.QueryRow(checkRoomQuery, checkRoomArgs...).Scan(&existingRoomID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ROOM_NOT_FOUND",
					"message": "Room not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to check room existence",
			},
		})
		return
	}

	var checkUserQuery string
	var checkUserArgs []interface{}
	if appID != "" {
		checkUserQuery = `SELECT id FROM users WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
		checkUserArgs = []interface{}{req.UserID, req.ApplicationID}
	} else {
		checkUserQuery = `SELECT id FROM users WHERE id = $1 AND deleted_at IS NULL`
		checkUserArgs = []interface{}{req.UserID}
	}

	var existingUserID uuid.UUID
	err = h.db.QueryRow(checkUserQuery, checkUserArgs...).Scan(&existingUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "USER_NOT_FOUND",
					"message": "User not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to check user existence",
			},
		})
		return
	}

	var existingMemberID uuid.UUID
	checkMemberQuery := `SELECT id FROM room_members WHERE room_id = $1 AND user_id = $2`
	err = h.db.QueryRow(checkMemberQuery, req.RoomID, req.UserID).Scan(&existingMemberID)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{
				"code":    "MEMBER_ALREADY_EXISTS",
				"message": "User is already a member of this room",
			},
		})
		return
	} else if err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to check member existence",
			},
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to add member to room",
			},
		})
		return
	}

	member.RoomID = req.RoomID
	member.UserID = req.UserID
	member.Role = role

	c.JSON(http.StatusCreated, gin.H{
		"message": "Member added to room successfully",
		"member":  member,
	})
}

func (h *RoomHandler) RemoveMember(c *gin.Context) {
	var req models.RoomRemoveMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	appID := ""
	if req.ApplicationID != nil {
		appID = req.ApplicationID.String()
		if !validator.ValidateUUID(appID) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "Invalid application ID",
				},
			})
			return
		}
	}

	var checkRoomQuery string
	var checkRoomArgs []interface{}
	if appID != "" {
		checkRoomQuery = `SELECT id FROM rooms WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
		checkRoomArgs = []interface{}{req.RoomID, req.ApplicationID}
	} else {
		checkRoomQuery = `SELECT id FROM rooms WHERE id = $1 AND deleted_at IS NULL`
		checkRoomArgs = []interface{}{req.RoomID}
	}

	var existingRoomID uuid.UUID
	err := h.db.QueryRow(checkRoomQuery, checkRoomArgs...).Scan(&existingRoomID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ROOM_NOT_FOUND",
					"message": "Room not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to check room existence",
			},
		})
		return
	}

	var existingMemberID uuid.UUID
	checkMemberQuery := `SELECT id FROM room_members WHERE room_id = $1 AND user_id = $2`
	err = h.db.QueryRow(checkMemberQuery, req.RoomID, req.UserID).Scan(&existingMemberID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "MEMBER_NOT_FOUND",
					"message": "User is not a member of this room",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to check member existence",
			},
		})
		return
	}

	deleteQuery := `DELETE FROM room_members WHERE room_id = $1 AND user_id = $2`
	result, err := h.db.Exec(deleteQuery, req.RoomID, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to remove member from room",
			},
		})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to remove member from room",
			},
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "MEMBER_NOT_FOUND",
				"message": "User is not a member of this room",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Member removed from room successfully",
	})
}

func (h *RoomHandler) ListMembers(c *gin.Context) {
	roomID := c.Query("room_id")
	appID := c.Query("application_id")

	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Room ID is required",
			},
		})
		return
	}

	if !validator.ValidateUUID(roomID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid room ID format",
			},
		})
		return
	}

	var checkRoomQuery string
	var checkRoomArgs []interface{}
	if appID != "" {
		if !validator.ValidateUUID(appID) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "Invalid application ID format",
				},
			})
			return
		}
		checkRoomQuery = `SELECT id FROM rooms WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
		checkRoomArgs = []interface{}{roomID, appID}
	} else {
		checkRoomQuery = `SELECT id FROM rooms WHERE id = $1 AND deleted_at IS NULL`
		checkRoomArgs = []interface{}{roomID}
	}

	var existingRoomID uuid.UUID
	err := h.db.QueryRow(checkRoomQuery, checkRoomArgs...).Scan(&existingRoomID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ROOM_NOT_FOUND",
					"message": "Room not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to check room existence",
			},
		})
		return
	}

	query := `SELECT id, room_id, user_id, role, joined_at, last_read_at
	          FROM room_members
	          WHERE room_id = $1
	          ORDER BY joined_at ASC`

	rows, err := h.db.Query(query, roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to retrieve room members",
			},
		})
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
			&member.Role,
			&member.JoinedAt,
			&member.LastReadAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "Failed to scan member",
				},
			})
			return
		}

		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to retrieve room members",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"members": members,
		"count":   len(members),
	})
}
