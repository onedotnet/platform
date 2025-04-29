package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	apperrors "github.com/onedotnet/platform/pkg/errors"
	"github.com/onedotnet/platform/pkg/logger"
	"go.uber.org/zap"
)

// ErrorResponse represents a standardized API error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	// Additional context data (optional)
	Context map[string]interface{} `json:"context,omitempty"`
}

// ErrorHandler is a middleware that catches and formats errors
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// If there are errors, handle them
		if len(c.Errors) > 0 {
			// Get the first error
			err := c.Errors[0].Err
			handleError(c, err)
		}
	}
}

// handleError processes different types of errors and returns appropriate responses
func handleError(c *gin.Context, err error) {
	var appErr *apperrors.AppError

	// Try to convert to AppError
	if errors.As(err, &appErr) {
		// It's one of our application errors
		logAppError(c, appErr)

		// Return the error response
		c.JSON(appErr.StatusCode, ErrorResponse{
			Code:    appErr.Code,
			Message: appErr.Message,
		})
		return
	}

	// Handle validation errors from gin binding
	var validationErrors interface{ Error() string }
	if errors.As(err, &validationErrors) {
		validationErr := apperrors.ValidationError("Validation failed")
		logError(c, validationErr, err)

		c.JSON(validationErr.StatusCode, ErrorResponse{
			Code:    validationErr.Code,
			Message: validationErr.Message,
			Context: map[string]interface{}{
				"errors": validationErrors.Error(),
			},
		})
		return
	}

	// Convert any other error to internal server error
	internalErr := apperrors.InternalServer("")
	logError(c, internalErr, err)

	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Code:    internalErr.Code,
		Message: internalErr.Message,
	})
}

// logAppError logs application errors with context
func logAppError(c *gin.Context, err *apperrors.AppError) {
	// Create fields for logging
	fields := []zap.Field{
		zap.String("error_code", err.Code),
		zap.Int("status_code", err.StatusCode),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
	}

	// Add request ID if available
	if reqID, exists := c.Get("request_id"); exists {
		fields = append(fields, zap.String("request_id", reqID.(string)))
	}

	// Add detail if available
	if err.Detail != "" {
		fields = append(fields, zap.String("detail", err.Detail))
	}

	// Add original error if available
	if err.Err != nil {
		fields = append(fields, zap.Error(err.Err))
	}

	// Log at appropriate level based on status code
	if err.StatusCode >= 500 {
		logger.Log.Error(err.Message, fields...)
	} else if err.StatusCode >= 400 {
		logger.Log.Warn(err.Message, fields...)
	} else {
		logger.Log.Info(err.Message, fields...)
	}
}

// logError logs non-AppError errors with context
func logError(c *gin.Context, appErr *apperrors.AppError, originalErr error) {
	// Create fields for logging
	fields := []zap.Field{
		zap.String("error_code", appErr.Code),
		zap.Int("status_code", appErr.StatusCode),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.Error(originalErr),
	}

	// Add request ID if available
	if reqID, exists := c.Get("request_id"); exists {
		fields = append(fields, zap.String("request_id", reqID.(string)))
	}

	// Always log non-AppErrors as errors
	logger.Log.Error("Unhandled error", fields...)
}
