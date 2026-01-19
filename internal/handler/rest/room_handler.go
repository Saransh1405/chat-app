package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"chat-app/internal/database"
)

type RoomHandler struct {
	db *database.DB
}

func NewRoomHandler(db *database.DB) *RoomHandler {
	return &RoomHandler{db: db}
}

func (h *RoomHandler) Create(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "Create room endpoint not yet implemented",
		},
	})
}

func (h *RoomHandler) Get(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "Get room endpoint not yet implemented",
		},
	})
}

func (h *RoomHandler) List(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "List rooms endpoint not yet implemented",
		},
	})
}

func (h *RoomHandler) Update(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "Update room endpoint not yet implemented",
		},
	})
}

func (h *RoomHandler) Delete(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "Delete room endpoint not yet implemented",
		},
	})
}

func (h *RoomHandler) AddMember(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "Add member endpoint not yet implemented",
		},
	})
}

func (h *RoomHandler) RemoveMember(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "Remove member endpoint not yet implemented",
		},
	})
}

func (h *RoomHandler) ListMembers(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "List members endpoint not yet implemented",
		},
	})
}

