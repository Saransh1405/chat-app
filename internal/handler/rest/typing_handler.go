package rest

import (
	"net/http"
	"time"

	"chat-app/internal/database"
	"chat-app/internal/handler/websocket"
	"chat-app/internal/models"
	"chat-app/internal/utils/helperfunctions"
	"chat-app/internal/utils/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TypingHandler struct {
	db    *database.DB
	wsHub *websocket.Hub
}

func NewTypingHandler(db *database.DB, wsHub *websocket.Hub) *TypingHandler {
	return &TypingHandler{
		db:    db,
		wsHub: wsHub,
	}
}

func (h *TypingHandler) Create(c *gin.Context) {
	userID := c.Param("user_id")
	if !validator.ValidateUUID(userID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid room ID format",
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

	if appID != "" {
		app := models.Application{}
		err := h.db.QueryRow("SELECT id FROM applications WHERE id = $1 AND deleted_at IS NULL", appID).Scan(&app.ID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "APPLICATION_NOT_FOUND",
					"message": "Application not found",
				},
			})
			return
		}

		room := models.Room{}
		err = h.db.QueryRow("SELECT id FROM rooms WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL", roomID, appID).Scan(&room.ID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ROOM_NOT_FOUND",
					"message": "Room not found",
				},
			})
			return
		}

		user := models.User{}
		err = h.db.QueryRow("SELECT id FROM users WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL", userID, appID).Scan(&user.ID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "USER_NOT_FOUND",
					"message": "User not found",
				},
			})
			return
		}
	} else {
		user := models.User{}
		err := h.db.QueryRow("SELECT id FROM users WHERE id = $1 AND deleted_at IS NULL", userID).Scan(&user.ID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "USER_NOT_FOUND",
					"message": "User not found",
				},
			})
			return
		}

		room := models.Room{}
		err = h.db.QueryRow("SELECT id FROM rooms WHERE id = $1 AND deleted_at IS NULL", roomID).Scan(&room.ID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ROOM_NOT_FOUND",
					"message": "Room not found",
				},
			})
			return
		}

		member := models.RoomMember{}
		err = h.db.QueryRow("SELECT id FROM room_members WHERE user_id = $1 AND room_id = $2 AND deleted_at IS NULL", userID, roomID).Scan(&member.ID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "MEMBER_NOT_FOUND",
					"message": "User is not a member of the room",
				},
			})
			return
		}
	}

	typing := models.TypingIndicator{
		RoomID:    uuid.MustParse(roomID),
		UserID:    uuid.MustParse(userID),
		ExpiresAt: time.Now().Add(10 * time.Second),
	}

	typingResult, err := helperfunctions.SaveTypingIndicatorToDB(h.db, &typing)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to save typing indicator",
			},
		})
		return
	}

	h.wsHub.Broadcast <- struct {
		RoomID  string
		Message websocket.MessageStruct
	}{
		RoomID: typingResult.RoomID.String(),
		Message: websocket.MessageStruct{
			Type:    "typing",
			Payload: typingResult,
		},
	}

	c.JSON(http.StatusCreated, gin.H{
		"typing": typingResult,
	})
}
