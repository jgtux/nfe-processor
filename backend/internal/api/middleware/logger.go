package middleware

import (
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

// Logger logs each request with method, path, status and latency.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf("[http] %s %s %d %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(start),
		)
	}
}
