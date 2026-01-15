package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"chat-app/internal/database"
)

type ApplicationHandler struct {
	db *database.DB
}

func NewApplicationHandler(db *database.DB) *ApplicationHandler {
	return &ApplicationHandler{db: db}
}

func (h *ApplicationHandler) Create(c *gin.Context) {
	// TODO: Implement create logic
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "Create application endpoint not yet implemented",
		},
	})
}

func (h *ApplicationHandler) Get(c *gin.Context) {
	// TODO: Implement get logic
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "Get application endpoint not yet implemented",
		},
	})
}

func (h *ApplicationHandler) List(c *gin.Context) {
	// TODO: Implement list logic
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "List applications endpoint not yet implemented",
		},
	})
}

func (h *ApplicationHandler) Update(c *gin.Context) {
	// TODO: Implement update logic
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "Update application endpoint not yet implemented",
		},
	})
}

func (h *ApplicationHandler) Delete(c *gin.Context) {
	// TODO: Implement delete logic
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "Delete application endpoint not yet implemented",
		},
	})
}

