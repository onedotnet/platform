package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/onedotnet/platform/internal/model"
	"github.com/onedotnet/platform/internal/service"
)

// OrganizationHandler handles HTTP requests related to organizations
type OrganizationHandler struct {
	repo service.Repository
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(repo service.Repository) *OrganizationHandler {
	return &OrganizationHandler{
		repo: repo,
	}
}

// Register registers organization routes with the provided router
func (h *OrganizationHandler) Register(router *gin.RouterGroup) {
	orgs := router.Group("/organizations")
	{
		orgs.GET("", h.List)
		orgs.POST("", h.Create)
		orgs.GET("/:id", h.Get)
		orgs.PUT("/:id", h.Update)
		orgs.DELETE("/:id", h.Delete)
		orgs.GET("/name/:name", h.GetByName)
	}
}

// Create handles organization creation
func (h *OrganizationHandler) Create(c *gin.Context) {
	var org model.Organization
	if err := c.ShouldBindJSON(&org); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.CreateOrganization(c.Request.Context(), &org); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, org)
}

// Get handles retrieving an organization by ID
func (h *OrganizationHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	org, err := h.repo.GetOrganization(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, org)
}

// GetByName handles retrieving an organization by name
func (h *OrganizationHandler) GetByName(c *gin.Context) {
	name := c.Param("name")

	org, err := h.repo.GetOrganizationByName(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, org)
}

// Update handles updating an organization
func (h *OrganizationHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	// Get existing organization
	existingOrg, err := h.repo.GetOrganization(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Bind updated fields
	var updatedOrg model.Organization
	if err := c.ShouldBindJSON(&updatedOrg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Preserve existing data for fields that aren't provided in the update
	// Only update fields with non-zero values from the request
	if updatedOrg.Name == "" {
		updatedOrg.Name = existingOrg.Name
	}
	if updatedOrg.DisplayName == "" {
		updatedOrg.DisplayName = existingOrg.DisplayName
	}
	if updatedOrg.Description == "" {
		updatedOrg.Description = existingOrg.Description
	}
	if updatedOrg.LogoURL == "" {
		updatedOrg.LogoURL = existingOrg.LogoURL
	}
	if updatedOrg.Website == "" {
		updatedOrg.Website = existingOrg.Website
	}

	// Ensure ID is preserved
	updatedOrg.ID = uint(id)

	// Preserve timestamps
	updatedOrg.CreatedAt = existingOrg.CreatedAt

	// Update organization
	if err := h.repo.UpdateOrganization(c.Request.Context(), &updatedOrg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedOrg)
}

// Delete handles deleting an organization
func (h *OrganizationHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	if err := h.repo.DeleteOrganization(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Organization deleted successfully"})
}

// List handles listing organizations with pagination
func (h *OrganizationHandler) List(c *gin.Context) {
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

	orgs, total, err := h.repo.ListOrganizations(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   orgs,
		"total":  total,
		"offset": offset,
		"limit":  limit,
	})
}
