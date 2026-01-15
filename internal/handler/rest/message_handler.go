package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"chat-app/internal/database"
	"chat-app/internal/handler/websocket"
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
	// TODO: Implement create logic
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "Create message endpoint not yet implemented",
		},
	})
}

func (h *MessageHandler) Get(c *gin.Context) {
	// TODO: Implement get logic
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "Get message endpoint not yet implemented",
		},
	})
}

func (h *MessageHandler) List(c *gin.Context) {
	// TODO: Implement list logic
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "List messages endpoint not yet implemented",
		},
	})
}

func (h *MessageHandler) Update(c *gin.Context) {
	// TODO: Implement update logic
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "Update message endpoint not yet implemented",
		},
	})
}

func (h *MessageHandler) Delete(c *gin.Context) {
	// TODO: Implement delete logic
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "Delete message endpoint not yet implemented",
		},
	})
}

