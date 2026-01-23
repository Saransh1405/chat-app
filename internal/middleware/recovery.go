package middleware

import (
	"log"
	"net/http"

	"chat-app/internal/utils/errors"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)

				errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeInternalError,
					"An unexpected error occurred", nil)

				c.Abort()
			}
		}()

		c.Next()
	}
}
