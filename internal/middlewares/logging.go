package middlewares

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func CustomLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		startTime := time.Now()

		// Process request
		c.Next()

		// Stop timer
		endTime := time.Now()
		latency := endTime.Sub(startTime)

		// Collect request details
		req := c.Request
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := req.Method
		path := req.RequestURI

		// Get user ID from context (assuming it's set elsewhere in your application)
		userID := c.GetString("userId")

		logLine := fmt.Sprintf("%s \"%s %s\" %d %s \"%s\" %s\n",
			clientIP, method, path, status, latency, userID, c.Request.UserAgent())

		logger := log.New(os.Stdout, "", log.Default().Flags())

		logger.Println(logLine)
	}
}
