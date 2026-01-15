package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"chat-app/internal/utils/jwt"
)

// Auth middleware validates JWT tokens and extracts user information
func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authorization header is required",
				},
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Invalid authorization header format. Expected: Bearer <token>",
				},
			})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate and parse token
		claims, err := jwt.ValidateToken(token, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Invalid or expired token",
				},
			})
			c.Abort()
			return
		}

		// Store user information in context
		c.Set("user_id", claims.UserID)
		c.Set("app_id", claims.AppID)
		c.Set("external_user_id", claims.ExternalUserID)

		c.Next()
	}
}

