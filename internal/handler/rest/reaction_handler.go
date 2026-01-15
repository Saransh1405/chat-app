package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"chat-app/internal/database"
	"chat-app/internal/handler/websocket"
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
	// TODO: Implement create logic
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "Create reaction endpoint not yet implemented",
		},
	})
}

func (h *ReactionHandler) Delete(c *gin.Context) {
	// TODO: Implement delete logic
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "Delete reaction endpoint not yet implemented",
		},
	})
}

func (h *ReactionHandler) List(c *gin.Context) {
	// TODO: Implement list logic
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "List reactions endpoint not yet implemented",
		},
	})
}

