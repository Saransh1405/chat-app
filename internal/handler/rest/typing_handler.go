package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"chat-app/internal/database"
	"chat-app/internal/handler/websocket"
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
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "Create typing indicator endpoint not yet implemented",
		},
	})
}

