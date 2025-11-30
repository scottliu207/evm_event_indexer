package middleware

import (
	"bytes"

	"github.com/gin-gonic/gin"
)

const INTERCEPTOR_KEY = "interceptor"

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w bodyLogWriter) Body() *bytes.Buffer {
	return w.body
}

func Interceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Set(INTERCEPTOR_KEY, blw)
		c.Next()
	}
}
