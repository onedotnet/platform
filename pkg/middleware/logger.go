package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/onedotnet/platform/pkg/logger"
	"go.uber.org/zap"
)

// Logger is a middleware that logs each API request using Zap
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Create request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Set request ID in context
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		// Process request
		c.Next()

		// Log information after request is processed
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// Create logger entry with request information
		logEntry := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("client_ip", clientIP),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.Int("size", c.Writer.Size()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		// Log errors separately if present
		if errorMessage != "" {
			logEntry = append(logEntry, zap.String("error", errorMessage))
		}

		// Log at appropriate level based on status code
		if statusCode >= 500 {
			logger.Log.Error("Server error", logEntry...)
		} else if statusCode >= 400 {
			logger.Log.Warn("Client error", logEntry...)
		} else {
			logger.Log.Info("Request processed", logEntry...)
		}
	}
}

// generateRequestID creates a unique ID for each request
func generateRequestID() string {
	// Create a timestamp prefix for better sorting and readability
	timestamp := time.Now().Format("20060102150405")

	// Generate a random suffix
	randBytes := make([]byte, 4)
	rand.Read(randBytes)
	randomSuffix := hex.EncodeToString(randBytes)

	return timestamp + "-" + randomSuffix
}

// randomString generates a random string of the specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)

	// Generate random indices for the charset
	randBytes := make([]byte, length)
	rand.Read(randBytes)

	// Map random bytes to the charset
	for i := range b {
		b[i] = charset[int(randBytes[i])%len(charset)]
	}

	return string(b)
}
