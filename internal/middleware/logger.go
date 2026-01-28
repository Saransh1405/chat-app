package middleware

import (
	"time"

	"chat-app/internal/utils/logger"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Store request ID in context for tracing
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = c.GetHeader("X-Request-Id")
		}
		if requestID != "" {
			c.Set("requestId", requestID)
		}

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		fields := logger.Fields{
			"request_id": requestID,
			"user_agent": c.Request.UserAgent(),
		}

		// Add user ID if available
		if userID, exists := c.Get("user_id"); exists {
			fields["user_id"] = userID
		}

		// Add error details if any
		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
		}

		logger.LogRequest(method, path, clientIP, statusCode, latency, fields)
	}
}
