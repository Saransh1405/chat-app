package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"chat-app/internal/utils/errors"
	"chat-app/internal/utils/logger"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()

				fields := logger.Fields{
					"panic":     fmt.Sprintf("%v", err),
					"stack":     string(stack),
					"method":    c.Request.Method,
					"path":      c.Request.URL.Path,
					"client_ip": c.ClientIP(),
				}

				if userID, exists := c.Get("user_id"); exists {
					fields["user_id"] = userID
				}

				if requestID, exists := c.Get("requestId"); exists {
					fields["request_id"] = requestID
				}

				logger.Error("Panic recovered", fmt.Errorf("%v", err), fields)

				errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeInternalError,
					"An unexpected error occurred", nil)

				c.Abort()
			}
		}()

		c.Next()
	}
}
