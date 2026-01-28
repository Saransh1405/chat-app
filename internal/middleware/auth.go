package middleware

import (
	"net/http"
	"strings"

	"chat-app/internal/utils/errors"
	"chat-app/internal/utils/jwt"
	"chat-app/internal/utils/logger"

	"github.com/gin-gonic/gin"
)

func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("Authentication failed: missing authorization header", logger.Fields{
				"client_ip": c.ClientIP(),
				"method":    c.Request.Method,
				"path":      c.Request.URL.Path,
			})
			errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
				"Authorization header is required", nil)
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warn("Authentication failed: invalid authorization header format", logger.Fields{
				"client_ip":  c.ClientIP(),
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"header_len": len(parts),
			})
			errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
				"Invalid authorization header format. Expected: Bearer <token>", nil)
			c.Abort()
			return
		}

		token := parts[1]

		claims, err := jwt.ValidateToken(token, jwtSecret)
		if err != nil {
			logger.Warn("Authentication failed: invalid or expired token", logger.Fields{
				"client_ip": c.ClientIP(),
				"method":    c.Request.Method,
				"path":      c.Request.URL.Path,
				"error":     err.Error(),
			})
			errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
				"Invalid or expired token", nil)
			c.Abort()
			return
		}

		logger.Debug("Authentication successful", logger.Fields{
			"user_id": claims.UserID,
			"app_id":  claims.AppID,
			"method":  c.Request.Method,
			"path":    c.Request.URL.Path,
		})

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
