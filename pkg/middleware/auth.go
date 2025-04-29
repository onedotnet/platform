package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/onedotnet/platform/internal/service"
)

// AuthMiddleware returns a middleware that handles authentication
func AuthMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the token from the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		// The Authorization header should be in the format "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer <token>"})
			return
		}

		// Extract the token
		tokenString := parts[1]

		// Verify the token
		claims, err := authService.VerifyToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Set the user ID and other claims in the context
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("uuid", claims.UUID)

		c.Next()
	}
}

// OptionalAuthMiddleware tries to authenticate the user but doesn't abort if token is missing
func OptionalAuthMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the token from the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Skip authentication if no header is provided
			c.Next()
			return
		}

		// The Authorization header should be in the format "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Skip authentication if header format is invalid
			c.Next()
			return
		}

		// Extract the token
		tokenString := parts[1]

		// Verify the token
		claims, err := authService.VerifyToken(tokenString)
		if err != nil {
			// Skip authentication if token is invalid
			c.Next()
			return
		}

		// Set the user ID and other claims in the context
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("uuid", claims.UUID)

		c.Next()
	}
}

// RoleAuthMiddleware checks if the authenticated user has the required role
func RoleAuthMiddleware(authService service.AuthService, requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from the context (set by AuthMiddleware)
		userID, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Get user from database
		user, err := authService.GetUserByID(c.Request.Context(), userID.(uint))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		// Check if user has one of the required roles
		hasRole := false
		for _, role := range user.Roles {
			for _, requiredRole := range requiredRoles {
				if role.Name == requiredRole {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}

		c.Next()
	}
}
