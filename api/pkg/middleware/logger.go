package middleware

import (
	"time"

	"github.com/chattycathy/api/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Logger is a Gin middleware that logs requests using zerolog
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		event := logger.Info()
		if c.Writer.Status() >= 500 {
			event = logger.Error()
		} else if c.Writer.Status() >= 400 {
			event = logger.Warn()
		}

		event.
			Str("request_id", GetRequestID(c)).
			Str("method", c.Request.Method).
			Str("path", path).
			Str("query", query).
			Str("ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent()).
			Int("status", c.Writer.Status()).
			Dur("latency", latency).
			Int("body_size", c.Writer.Size()).
			Msg("request")
	}
}
