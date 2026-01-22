package rest

import (
	"net/http"

	"chat-app/internal/database"
	"chat-app/internal/handler/websocket"
	"chat-app/internal/utils/errors"

	"github.com/gin-gonic/gin"
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
	errors.RespondWithError(c, http.StatusNotImplemented, errors.ErrCodeInternalError,
		"Create reaction endpoint not yet implemented", nil)
}

func (h *ReactionHandler) Delete(c *gin.Context) {
	errors.RespondWithError(c, http.StatusNotImplemented, errors.ErrCodeInternalError,
		"Delete reaction endpoint not yet implemented", nil)
}

func (h *ReactionHandler) List(c *gin.Context) {
	errors.RespondWithError(c, http.StatusNotImplemented, errors.ErrCodeInternalError,
		"List reactions endpoint not yet implemented", nil)
}
