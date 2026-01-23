package middleware

import (
	"net/http"
	"strings"

	"chat-app/internal/utils/errors"
	"chat-app/internal/utils/jwt"

	"github.com/gin-gonic/gin"
)

func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
				"Authorization header is required", nil)
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
				"Invalid authorization header format. Expected: Bearer <token>", nil)
			c.Abort()
			return
		}

		token := parts[1]

		claims, err := jwt.ValidateToken(token, jwtSecret)
		if err != nil {
			errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
				"Invalid or expired token", nil)
			c.Abort()
			return
		}

		// Note: Database access in middleware requires refactoring to pass db instance
		// For now, we'll skip these checks in middleware and rely on handler-level validation
		// TODO: Refactor middleware to accept database connection or move these checks to handlers

		// Store user information in context
		c.Set("user_id", claims.UserID)
		c.Set("app_id", claims.AppID)
		c.Set("room_id", claims.RoomID)

		c.Next()
	}
}
