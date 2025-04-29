package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/onedotnet/platform/internal/model"
	"github.com/onedotnet/platform/internal/service"
	apperrors "github.com/onedotnet/platform/pkg/errors"
	"github.com/onedotnet/platform/pkg/logger"
	"go.uber.org/zap"
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
// @Summary Create a new user
// @Description Create a new user in the system
// @Tags Users
// @Accept json
// @Produce json
// @Param user body model.User true "User information"
// @Success 201 {object} model.User "Created user"
// @Failure 400 {object} middleware.ErrorResponse "Invalid input"
// @Failure 500 {object} middleware.ErrorResponse "Server error"
// @Router /users [post]
// @Security BearerAuth
func (h *UserHandler) Create(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		logger.Log.Debug("Failed to bind user data", zap.Error(err))
		appErr := apperrors.BadRequest(err.Error())
		c.Error(appErr)
		return
	}

	if err := h.repo.CreateUser(c.Request.Context(), &user); err != nil {
		logger.Log.Error("Failed to create user",
			zap.String("username", user.Username),
			zap.String("email", user.Email),
			zap.Error(err))
		appErr := apperrors.InternalServer("Failed to create user")
		c.Error(appErr.WithError(err))
		return
	}

	// Don't return the password
	user.Password = ""

	logger.Log.Info("User created successfully",
		zap.String("username", user.Username),
		zap.String("email", user.Email),
		zap.Uint("id", user.ID))

	c.JSON(http.StatusCreated, user)
}

// Get handles retrieving a user by ID
// @Summary Get a user by ID
// @Description Get a user by their ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} model.User "User details"
// @Failure 400 {object} middleware.ErrorResponse "Invalid user ID"
// @Failure 404 {object} middleware.ErrorResponse "User not found"
// @Failure 500 {object} middleware.ErrorResponse "Server error"
// @Router /users/{id} [get]
// @Security BearerAuth
func (h *UserHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		appErr := apperrors.BadRequest("Invalid user ID")
		c.Error(appErr.WithError(err))
		return
	}

	user, err := h.repo.GetUser(c.Request.Context(), uint(id))
	if err != nil {
		appErr := apperrors.NotFound("User not found")
		c.Error(appErr.WithError(err))
		return
	}

	// Don't return the password
	user.Password = ""

	c.JSON(http.StatusOK, user)
}

// GetByUsername handles retrieving a user by username
// @Summary Get a user by username
// @Description Get a user by their username
// @Tags Users
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} model.User "User details"
// @Failure 404 {object} middleware.ErrorResponse "User not found"
// @Failure 500 {object} middleware.ErrorResponse "Server error"
// @Router /users/username/{username} [get]
// @Security BearerAuth
func (h *UserHandler) GetByUsername(c *gin.Context) {
	username := c.Param("username")

	user, err := h.repo.GetUserByUsername(c.Request.Context(), username)
	if err != nil {
		appErr := apperrors.NotFound("User not found")
		c.Error(appErr.WithError(err))
		return
	}

	// Don't return the password
	user.Password = ""

	c.JSON(http.StatusOK, user)
}

// GetByEmail handles retrieving a user by email
// @Summary Get a user by email
// @Description Get a user by their email address
// @Tags Users
// @Accept json
// @Produce json
// @Param email path string true "Email address"
// @Success 200 {object} model.User "User details"
// @Failure 404 {object} middleware.ErrorResponse "User not found"
// @Failure 500 {object} middleware.ErrorResponse "Server error"
// @Router /users/email/{email} [get]
// @Security BearerAuth
func (h *UserHandler) GetByEmail(c *gin.Context) {
	email := c.Param("email")

	user, err := h.repo.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		appErr := apperrors.NotFound("User not found")
		c.Error(appErr.WithError(err))
		return
	}

	// Don't return the password
	user.Password = ""

	c.JSON(http.StatusOK, user)
}

// Update handles updating a user
// @Summary Update a user
// @Description Update an existing user's information
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body model.User true "Updated user information"
// @Success 200 {object} model.User "Updated user details"
// @Failure 400 {object} middleware.ErrorResponse "Invalid user ID or data"
// @Failure 404 {object} middleware.ErrorResponse "User not found"
// @Failure 500 {object} middleware.ErrorResponse "Server error"
// @Router /users/{id} [put]
// @Security BearerAuth
func (h *UserHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		appErr := apperrors.BadRequest("Invalid user ID")
		c.Error(appErr.WithError(err))
		return
	}

	// Get existing user
	existingUser, err := h.repo.GetUser(c.Request.Context(), uint(id))
	if err != nil {
		appErr := apperrors.NotFound("User not found")
		c.Error(appErr.WithError(err))
		return
	}

	// Bind updated fields
	var updatedUser model.User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		appErr := apperrors.BadRequest(err.Error())
		c.Error(appErr.WithError(err))
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
		logger.Log.Error("Failed to update user",
			zap.Uint("id", updatedUser.ID),
			zap.Error(err))
		appErr := apperrors.InternalServer("Failed to update user")
		c.Error(appErr.WithError(err))
		return
	}

	// Don't return the password
	updatedUser.Password = ""

	logger.Log.Info("User updated successfully",
		zap.Uint("id", updatedUser.ID),
		zap.String("username", updatedUser.Username))

	c.JSON(http.StatusOK, updatedUser)
}

// Delete handles deleting a user
// @Summary Delete a user
// @Description Delete an existing user
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} middleware.ErrorResponse "Invalid user ID"
// @Failure 500 {object} middleware.ErrorResponse "Server error"
// @Router /users/{id} [delete]
// @Security BearerAuth
func (h *UserHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		appErr := apperrors.BadRequest("Invalid user ID")
		c.Error(appErr.WithError(err))
		return
	}

	if err := h.repo.DeleteUser(c.Request.Context(), uint(id)); err != nil {
		logger.Log.Error("Failed to delete user",
			zap.Uint("id", uint(id)),
			zap.Error(err))
		appErr := apperrors.InternalServer("Failed to delete user")
		c.Error(appErr.WithError(err))
		return
	}

	logger.Log.Info("User deleted successfully", zap.Uint("id", uint(id)))

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// List handles listing users with pagination
// @Summary List all users
// @Description Get a paginated list of users
// @Tags Users
// @Accept json
// @Produce json
// @Param offset query int false "Pagination offset" default(0)
// @Param limit query int false "Pagination limit" default(10)
// @Success 200 {object} map[string]interface{} "List of users with pagination metadata"
// @Failure 500 {object} middleware.ErrorResponse "Server error"
// @Router /users [get]
// @Security BearerAuth
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
		logger.Log.Error("Failed to list users", zap.Error(err))
		appErr := apperrors.InternalServer("Failed to list users")
		c.Error(appErr.WithError(err))
		return
	}

	// Don't return passwords
	for i := range users {
		users[i].Password = ""
	}

	logger.Log.Info("Listed users successfully",
		zap.Int("count", len(users)),
		zap.Int64("total", total))

	c.JSON(http.StatusOK, gin.H{
		"data":   users,
		"total":  total,
		"offset": offset,
		"limit":  limit,
	})
}
