package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/onedotnet/platform/internal/model"
	"github.com/onedotnet/platform/internal/service"
)

// UserHandler handles HTTP requests related to users
type UserHandler struct {
	repo service.Repository
}

// NewUserHandler creates a new user handler
func NewUserHandler(repo service.Repository) *UserHandler {
	return &UserHandler{
		repo: repo,
	}
}

// Register registers user routes with the provided router
func (h *UserHandler) Register(router *gin.RouterGroup) {
	users := router.Group("/users")
	{
		users.GET("", h.List)
		users.POST("", h.Create)
		users.GET("/:id", h.Get)
		users.PUT("/:id", h.Update)
		users.DELETE("/:id", h.Delete)
		users.GET("/username/:username", h.GetByUsername)
		users.GET("/email/:email", h.GetByEmail)
	}
}

// Create handles user creation
func (h *UserHandler) Create(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.CreateUser(c.Request.Context(), &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Don't return the password
	user.Password = ""

	c.JSON(http.StatusCreated, user)
}

// Get handles retrieving a user by ID
func (h *UserHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.repo.GetUser(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Don't return the password
	user.Password = ""

	c.JSON(http.StatusOK, user)
}

// GetByUsername handles retrieving a user by username
func (h *UserHandler) GetByUsername(c *gin.Context) {
	username := c.Param("username")

	user, err := h.repo.GetUserByUsername(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Don't return the password
	user.Password = ""

	c.JSON(http.StatusOK, user)
}

// GetByEmail handles retrieving a user by email
func (h *UserHandler) GetByEmail(c *gin.Context) {
	email := c.Param("email")

	user, err := h.repo.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Don't return the password
	user.Password = ""

	c.JSON(http.StatusOK, user)
}

// Update handles updating a user
func (h *UserHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get existing user
	existingUser, err := h.repo.GetUser(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Bind updated fields
	var updatedUser model.User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Preserve existing data for fields that aren't provided in the update
	if updatedUser.Username == "" {
		updatedUser.Username = existingUser.Username
	}
	if updatedUser.Email == "" {
		updatedUser.Email = existingUser.Email
	}
	if updatedUser.Password == "" {
		updatedUser.Password = existingUser.Password
	}
	if updatedUser.FirstName == "" {
		updatedUser.FirstName = existingUser.FirstName
	}
	if updatedUser.LastName == "" {
		updatedUser.LastName = existingUser.LastName
	}
	if updatedUser.Phone == "" {
		updatedUser.Phone = existingUser.Phone
	}

	// Maintain LastLoginAt if not provided
	if updatedUser.LastLoginAt == nil {
		updatedUser.LastLoginAt = existingUser.LastLoginAt
	}

	// Ensure ID is preserved
	updatedUser.ID = uint(id)

	// Preserve timestamps
	updatedUser.CreatedAt = existingUser.CreatedAt

	// Update user
	if err := h.repo.UpdateUser(c.Request.Context(), &updatedUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Don't return the password
	updatedUser.Password = ""

	c.JSON(http.StatusOK, updatedUser)
}

// Delete handles deleting a user
func (h *UserHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.repo.DeleteUser(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// List handles listing users with pagination
func (h *UserHandler) List(c *gin.Context) {
	// Parse pagination parameters
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "10")

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	users, total, err := h.repo.ListUsers(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Don't return passwords
	for i := range users {
		users[i].Password = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   users,
		"total":  total,
		"offset": offset,
		"limit":  limit,
	})
}
