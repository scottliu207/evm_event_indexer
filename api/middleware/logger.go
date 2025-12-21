package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// logging api request and response
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		buf, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(buf))

		// Process request
		c.Next()

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)

		method := c.Request.Method
		statusCode := c.Writer.Status()

		comment := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if raw != "" {
			path = path + "?" + raw
		}

		// get error from gin context, handle only the first error
		if lastError := c.Errors.Last(); lastError != nil {
			slog.Error(
				"API response error",
				slog.Any("statusCode", statusCode),
				slog.Any("latency", latency/time.Millisecond),
				slog.Any("method", method),
				slog.Any("path", path),
				slog.Any("comment", comment),
				slog.Any("Stack", lastError.Err),
			)
			return
		}

		slog.Info(
			"API response success",
			slog.Any("statusCode", statusCode),
			slog.Any("latency", latency/time.Millisecond),
			slog.Any("method", method),
			slog.Any("path", path),
			slog.Any("comment", comment))
	}
}
