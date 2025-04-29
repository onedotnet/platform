package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/onedotnet/platform/internal/cache"
	"github.com/onedotnet/platform/internal/model"
	"gorm.io/gorm"
)

const (
	// Cache expiration time
	defaultCacheExpiration = 15 * time.Minute
)

// GormRepository implements the Repository interface using GORM
type GormRepository struct {
	db    *gorm.DB
	cache cache.Cache
}

// NewGormRepository creates a new GORM repository with cache
func NewGormRepository(db *gorm.DB, cache cache.Cache) *GormRepository {
	return &GormRepository{
		db:    db,
		cache: cache,
	}
}

// User repository methods

// CreateUser validates the user, saves to cache, and then to the database
func (r *GormRepository) CreateUser(ctx context.Context, user *model.User) error {
	// Validate first
	if err := user.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Save to database
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Save to cache
	if err := r.cache.Set(ctx, user.CacheKey(), user, defaultCacheExpiration); err != nil {
		// Just log the error but don't return it
		fmt.Printf("cache error: %v\n", err)
	}

	return nil
}

// GetUser retrieves a user by ID, first checking cache then database
func (r *GormRepository) GetUser(ctx context.Context, id uint) (*model.User, error) {
	user := &model.User{}

	// Try to get from cache
	cacheKey := fmt.Sprintf("user:%d", id)
	err := r.cache.Get(ctx, cacheKey, user)
	if err == nil {
		return user, nil
	}

	// Get from database
	if err := r.db.First(user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found: %d", id)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Save to cache
	if err := r.cache.Set(ctx, cacheKey, user, defaultCacheExpiration); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}

	return user, nil
}

// GetUserByUUID retrieves a user by UUID, first checking cache then database
func (r *GormRepository) GetUserByUUID(ctx context.Context, uuid uuid.UUID) (*model.User, error) {
	user := &model.User{}

	// Try to get from cache
	cacheKey := fmt.Sprintf("user:%s", uuid.String())
	err := r.cache.Get(ctx, cacheKey, user)
	if err == nil {
		return user, nil
	}

	// Get from database
	if err := r.db.Where("uuid = ?", uuid).First(user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found with UUID: %s", uuid)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Save to cache
	if err := r.cache.Set(ctx, cacheKey, user, defaultCacheExpiration); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (r *GormRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	user := &model.User{}

	// Try to get from cache
	cacheKey := fmt.Sprintf("user:username:%s", username)
	err := r.cache.Get(ctx, cacheKey, user)
	if err == nil {
		return user, nil
	}

	// Get from database
	if err := r.db.Where("username = ?", username).First(user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found with username: %s", username)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Save to cache with UUID as the primary key
	if err := r.cache.Set(ctx, user.CacheKey(), user, defaultCacheExpiration); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}
	if err := r.cache.Set(ctx, cacheKey, user, defaultCacheExpiration); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (r *GormRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}

	// Try to get from cache
	cacheKey := fmt.Sprintf("user:email:%s", email)
	err := r.cache.Get(ctx, cacheKey, user)
	if err == nil {
		return user, nil
	}

	// Get from database
	if err := r.db.Where("email = ?", email).First(user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found with email: %s", email)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Save to cache with UUID as the primary key
	if err := r.cache.Set(ctx, user.CacheKey(), user, defaultCacheExpiration); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}
	if err := r.cache.Set(ctx, cacheKey, user, defaultCacheExpiration); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}

	return user, nil
}

// UpdateUser validates the user, updates cache, and then updates the database
func (r *GormRepository) UpdateUser(ctx context.Context, user *model.User) error {
	// Validate first
	if err := user.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Update database
	if err := r.db.Save(user).Error; err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Update cache
	if err := r.cache.Set(ctx, user.CacheKey(), user, defaultCacheExpiration); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}

	return nil
}

// DeleteUser removes a user from cache and database
func (r *GormRepository) DeleteUser(ctx context.Context, id uint) error {
	// Check if user exists
	user, err := r.GetUser(ctx, id)
	if err != nil {
		return err
	}

	// Delete from database
	if err := r.db.Delete(user).Error; err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Delete from cache
	if err := r.cache.Delete(ctx, user.CacheKey()); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}

	// Also delete by username and email cache
	r.cache.Delete(ctx, fmt.Sprintf("user:username:%s", user.Username))
	r.cache.Delete(ctx, fmt.Sprintf("user:email:%s", user.Email))

	return nil
}

// DeleteUserByUUID removes a user from cache and database using UUID
func (r *GormRepository) DeleteUserByUUID(ctx context.Context, uuid uuid.UUID) error {
	// Check if user exists
	user, err := r.GetUserByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	// Delete from database
	if err := r.db.Delete(user).Error; err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Delete from cache
	if err := r.cache.Delete(ctx, user.CacheKey()); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}

	// Also delete by username and email cache
	r.cache.Delete(ctx, fmt.Sprintf("user:username:%s", user.Username))
	r.cache.Delete(ctx, fmt.Sprintf("user:email:%s", user.Email))

	return nil
}

// ListUsers lists users with pagination
func (r *GormRepository) ListUsers(ctx context.Context, offset, limit int) ([]model.User, int64, error) {
	var users []model.User
	var count int64

	// Get count
	if err := r.db.Model(&model.User{}).Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("database error: %w", err)
	}

	// Get users with pagination
	if err := r.db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("database error: %w", err)
	}

	// Cache not used for list operations as they are less frequently accessed
	// and would consume more cache space

	return users, count, nil
}

// Organization repository methods

// CreateOrganization validates the organization, saves to cache, and then to the database
func (r *GormRepository) CreateOrganization(ctx context.Context, org *model.Organization) error {
	// Validate first
	if err := org.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Save to database
	if err := r.db.Create(org).Error; err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Save to cache
	if err := r.cache.Set(ctx, org.CacheKey(), org, defaultCacheExpiration); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}

	return nil
}

// GetOrganization retrieves an organization by ID, first checking cache then database
func (r *GormRepository) GetOrganization(ctx context.Context, id uint) (*model.Organization, error) {
	org := &model.Organization{}

	// Try to get from cache
	cacheKey := fmt.Sprintf("organization:%d", id)
	err := r.cache.Get(ctx, cacheKey, org)
	if err == nil {
		return org, nil
	}

	// Get from database
	if err := r.db.First(org, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("organization not found: %d", id)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Save to cache
	if err := r.cache.Set(ctx, cacheKey, org, defaultCacheExpiration); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}

	return org, nil
}

// GetOrganizationByUUID retrieves an organization by UUID
func (r *GormRepository) GetOrganizationByUUID(ctx context.Context, uuid uuid.UUID) (*model.Organization, error) {
	org := &model.Organization{}

	// Try to get from cache
	cacheKey := fmt.Sprintf("organization:%s", uuid.String())
	err := r.cache.Get(ctx, cacheKey, org)
	if err == nil {
		return org, nil
	}

	// Get from database
	if err := r.db.Where("uuid = ?", uuid).First(org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("organization not found with UUID: %s", uuid)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Save to cache
	if err := r.cache.Set(ctx, cacheKey, org, defaultCacheExpiration); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}

	return org, nil
}

// GetOrganizationByName retrieves an organization by name
func (r *GormRepository) GetOrganizationByName(ctx context.Context, name string) (*model.Organization, error) {
	org := &model.Organization{}

	// Try to get from cache
	cacheKey := fmt.Sprintf("organization:name:%s", name)
	err := r.cache.Get(ctx, cacheKey, org)
	if err == nil {
		return org, nil
	}

	// Get from database
	if err := r.db.Where("name = ?", name).First(org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("organization not found with name: %s", name)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Save to cache
	if err := r.cache.Set(ctx, org.CacheKey(), org, defaultCacheExpiration); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}
	if err := r.cache.Set(ctx, cacheKey, org, defaultCacheExpiration); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}

	return org, nil
}

// UpdateOrganization validates the organization, updates cache, and then updates the database
func (r *GormRepository) UpdateOrganization(ctx context.Context, org *model.Organization) error {
	// Validate first
	if err := org.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Update database
	if err := r.db.Save(org).Error; err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Update cache
	if err := r.cache.Set(ctx, org.CacheKey(), org, defaultCacheExpiration); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}

	return nil
}

// DeleteOrganization removes an organization from cache and database
func (r *GormRepository) DeleteOrganization(ctx context.Context, id uint) error {
	// Check if organization exists
	org, err := r.GetOrganization(ctx, id)
	if err != nil {
		return err
	}

	// Delete from database
	if err := r.db.Delete(org).Error; err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Delete from cache
	if err := r.cache.Delete(ctx, org.CacheKey()); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}

	// Also delete by name cache
	r.cache.Delete(ctx, fmt.Sprintf("organization:name:%s", org.Name))

	return nil
}

// DeleteOrganizationByUUID removes an organization from cache and database using UUID
func (r *GormRepository) DeleteOrganizationByUUID(ctx context.Context, uuid uuid.UUID) error {
	// Check if organization exists
	org, err := r.GetOrganizationByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	// Delete from database
	if err := r.db.Delete(org).Error; err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Delete from cache
	if err := r.cache.Delete(ctx, org.CacheKey()); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}

	// Also delete by name cache
	r.cache.Delete(ctx, fmt.Sprintf("organization:name:%s", org.Name))

	return nil
}

// ListOrganizations lists organizations with pagination
func (r *GormRepository) ListOrganizations(ctx context.Context, offset, limit int) ([]model.Organization, int64, error) {
	var orgs []model.Organization
	var count int64

	// Get count
	if err := r.db.Model(&model.Organization{}).Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("database error: %w", err)
	}

	// Get organizations with pagination
	if err := r.db.Offset(offset).Limit(limit).Find(&orgs).Error; err != nil {
		return nil, 0, fmt.Errorf("database error: %w", err)
	}

	return orgs, count, nil
}

// Role repository methods

// CreateRole validates the role, saves to cache, and then to the database
func (r *GormRepository) CreateRole(ctx context.Context, role *model.Role) error {
	// Validate first
	if err := role.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Save to database
	if err := r.db.Create(role).Error; err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Save to cache
	if err := r.cache.Set(ctx, role.CacheKey(), role, defaultCacheExpiration); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}

	return nil
}

// GetRole retrieves a role by ID, first checking cache then database
func (r *GormRepository) GetRole(ctx context.Context, id uint) (*model.Role, error) {
	role := &model.Role{}

	// Try to get from cache
	cacheKey := fmt.Sprintf("role:%d", id)
	err := r.cache.Get(ctx, cacheKey, role)
	if err == nil {
		return role, nil
	}

	// Get from database
	if err := r.db.First(role, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("role not found: %d", id)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Save to cache
	if err := r.cache.Set(ctx, cacheKey, role, defaultCacheExpiration); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}

	return role, nil
}

// GetRoleByUUID retrieves a role by UUID
func (r *GormRepository) GetRoleByUUID(ctx context.Context, uuid uuid.UUID) (*model.Role, error) {
	role := &model.Role{}

	// Try to get from cache
	cacheKey := fmt.Sprintf("role:%s", uuid.String())
	err := r.cache.Get(ctx, cacheKey, role)
	if err == nil {
		return role, nil
	}

	// Get from database
	if err := r.db.Where("uuid = ?", uuid).First(role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("role not found with UUID: %s", uuid)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Save to cache
	if err := r.cache.Set(ctx, cacheKey, role, defaultCacheExpiration); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}

	return role, nil
}

// GetRoleByName retrieves a role by name
func (r *GormRepository) GetRoleByName(ctx context.Context, name string) (*model.Role, error) {
	role := &model.Role{}

	// Try to get from cache
	cacheKey := fmt.Sprintf("role:name:%s", name)
	err := r.cache.Get(ctx, cacheKey, role)
	if err == nil {
		return role, nil
	}

	// Get from database
	if err := r.db.Where("name = ?", name).First(role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("role not found with name: %s", name)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Save to cache
	if err := r.cache.Set(ctx, role.CacheKey(), role, defaultCacheExpiration); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}
	if err := r.cache.Set(ctx, cacheKey, role, defaultCacheExpiration); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}

	return role, nil
}

// UpdateRole validates the role, updates cache, and then updates the database
func (r *GormRepository) UpdateRole(ctx context.Context, role *model.Role) error {
	// Validate first
	if err := role.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Update database
	if err := r.db.Save(role).Error; err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Update cache
	if err := r.cache.Set(ctx, role.CacheKey(), role, defaultCacheExpiration); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}

	return nil
}

// DeleteRole removes a role from cache and database
func (r *GormRepository) DeleteRole(ctx context.Context, id uint) error {
	// Check if role exists
	role, err := r.GetRole(ctx, id)
	if err != nil {
		return err
	}

	// Delete from database
	if err := r.db.Delete(role).Error; err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Delete from cache
	if err := r.cache.Delete(ctx, role.CacheKey()); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}

	// Also delete by name cache
	r.cache.Delete(ctx, fmt.Sprintf("role:name:%s", role.Name))

	return nil
}

// DeleteRoleByUUID removes a role from cache and database using UUID
func (r *GormRepository) DeleteRoleByUUID(ctx context.Context, uuid uuid.UUID) error {
	// Check if role exists
	role, err := r.GetRoleByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	// Delete from database
	if err := r.db.Delete(role).Error; err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Delete from cache
	if err := r.cache.Delete(ctx, role.CacheKey()); err != nil {
		fmt.Printf("cache error: %v\n", err)
	}

	// Also delete by name cache
	r.cache.Delete(ctx, fmt.Sprintf("role:name:%s", role.Name))

	return nil
}

// ListRoles lists roles with pagination
func (r *GormRepository) ListRoles(ctx context.Context, offset, limit int) ([]model.Role, int64, error) {
	var roles []model.Role
	var count int64

	// Get count
	if err := r.db.Model(&model.Role{}).Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("database error: %w", err)
	}

	// Get roles with pagination
	if err := r.db.Offset(offset).Limit(limit).Find(&roles).Error; err != nil {
		return nil, 0, fmt.Errorf("database error: %w", err)
	}

	return roles, count, nil
}
