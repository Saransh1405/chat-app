package rest

import (
	"net/http"

	"chat-app/internal/config"
	"chat-app/internal/database"
	"chat-app/internal/utils/helperfunctions"
	"chat-app/internal/utils/jwt"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	config *config.Config
	db     *database.DB
}

func NewAuthHandler(cfg *config.Config, db *database.DB) *AuthHandler {
	return &AuthHandler{
		config: cfg,
		db:     db,
	}
}

// Register handles application registration
func (h *AuthHandler) Register(c *gin.Context) {
	// TODO: Implement registration logic
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "Registration endpoint not yet implemented",
		},
	})
}

// Login handles user authentication
func (h *AuthHandler) Login(c *gin.Context) {
	email := c.Query("email")

	err := c.ShouldBindQuery(&email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid email or password",
			},
		})
		return
	}

	// get the user from the database
	user, err := helperfunctions.GetUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "USER_NOT_FOUND",
				"message": "User not found",
			},
		})
		return
	}

	if err := helperfunctions.CheckIfUserExistsWithEmail(user, ""); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "USER_NOT_FOUND",
				"message": "User not found",
			},
		})
		return
	}

	token, err := jwt.GenerateToken(user.ID.String(), user.ApplicationID.String(), user.ExternalID, h.config.JWT.Secret, h.config.JWT.AccessExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to generate token",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error": gin.H{
			"code":    "OK",
			"message": "Login successful",
			"token":   token,
			"user":    user,
		},
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// TODO: Implement token refresh logic
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "Token refresh endpoint not yet implemented",
		},
	})
}
