package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/onedotnet/platform/internal/model"
	"github.com/onedotnet/platform/internal/service"
)

// RoleHandler handles HTTP requests related to roles
type RoleHandler struct {
	repo service.Repository
}

// NewRoleHandler creates a new role handler
func NewRoleHandler(repo service.Repository) *RoleHandler {
	return &RoleHandler{
		repo: repo,
	}
}

// Register registers role routes with the provided router
func (h *RoleHandler) Register(router *gin.RouterGroup) {
	roles := router.Group("/roles")
	{
		roles.GET("", h.List)
		roles.POST("", h.Create)
		roles.GET("/:id", h.Get)
		roles.PUT("/:id", h.Update)
		roles.DELETE("/:id", h.Delete)
		roles.GET("/name/:name", h.GetByName)
	}
}

// Create handles role creation
func (h *RoleHandler) Create(c *gin.Context) {
	var role model.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.CreateRole(c.Request.Context(), &role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, role)
}

// Get handles retrieving a role by ID
func (h *RoleHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	role, err := h.repo.GetRole(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, role)
}

// GetByName handles retrieving a role by name
func (h *RoleHandler) GetByName(c *gin.Context) {
	name := c.Param("name")

	role, err := h.repo.GetRoleByName(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, role)
}

// Update handles updating a role
func (h *RoleHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	// Get existing role
	existingRole, err := h.repo.GetRole(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Bind updated fields
	var updatedRole model.Role
	if err := c.ShouldBindJSON(&updatedRole); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Preserve existing data for fields that aren't provided in the update
	if updatedRole.Name == "" {
		updatedRole.Name = existingRole.Name
	}
	if updatedRole.Description == "" {
		updatedRole.Description = existingRole.Description
	}

	// Ensure ID is preserved
	updatedRole.ID = uint(id)

	// Preserve timestamps
	updatedRole.CreatedAt = existingRole.CreatedAt

	// Update role
	if err := h.repo.UpdateRole(c.Request.Context(), &updatedRole); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedRole)
}

// Delete handles deleting a role
func (h *RoleHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	if err := h.repo.DeleteRole(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}

// List handles listing roles with pagination
func (h *RoleHandler) List(c *gin.Context) {
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

	roles, total, err := h.repo.ListRoles(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   roles,
		"total":  total,
		"offset": offset,
		"limit":  limit,
	})
}
