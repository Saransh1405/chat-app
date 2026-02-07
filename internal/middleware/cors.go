package middleware

import (
	"strings"

	"chat-app/internal/config"

	"github.com/gin-gonic/gin"
)

func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		allowedOrigin := ""
		allowWildcard := false

		for _, allowed := range cfg.AllowedOrigins {
			if allowed == "*" {
				if !cfg.AllowCredentials {
					allowWildcard = true
					break
				}
			} else if allowed == origin {
				allowedOrigin = origin
				break
			}
		}

		if allowWildcard {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else if allowedOrigin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		}

		if len(cfg.AllowedMethods) > 0 {
			c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
		}

		if len(cfg.AllowedHeaders) > 0 {
			c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
		}

		if cfg.AllowCredentials && (allowedOrigin != "" || allowWildcard) {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length,Content-Type")

		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
