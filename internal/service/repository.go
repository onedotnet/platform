package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/onedotnet/platform/internal/model"
)

// Repository defines the interface for persistence operations
type Repository interface {
	// User operations
	CreateUser(ctx context.Context, user *model.User) error
	GetUser(ctx context.Context, id uint) (*model.User, error)
	GetUserByUUID(ctx context.Context, uuid uuid.UUID) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id uint) error
	DeleteUserByUUID(ctx context.Context, uuid uuid.UUID) error
	ListUsers(ctx context.Context, offset, limit int) ([]model.User, int64, error)

	// Organization operations
	CreateOrganization(ctx context.Context, org *model.Organization) error
	GetOrganization(ctx context.Context, id uint) (*model.Organization, error)
	GetOrganizationByUUID(ctx context.Context, uuid uuid.UUID) (*model.Organization, error)
	GetOrganizationByName(ctx context.Context, name string) (*model.Organization, error)
	UpdateOrganization(ctx context.Context, org *model.Organization) error
	DeleteOrganization(ctx context.Context, id uint) error
	DeleteOrganizationByUUID(ctx context.Context, uuid uuid.UUID) error
	ListOrganizations(ctx context.Context, offset, limit int) ([]model.Organization, int64, error)

	// Role operations
	CreateRole(ctx context.Context, role *model.Role) error
	GetRole(ctx context.Context, id uint) (*model.Role, error)
	GetRoleByUUID(ctx context.Context, uuid uuid.UUID) (*model.Role, error)
	GetRoleByName(ctx context.Context, name string) (*model.Role, error)
	UpdateRole(ctx context.Context, role *model.Role) error
	DeleteRole(ctx context.Context, id uint) error
	DeleteRoleByUUID(ctx context.Context, uuid uuid.UUID) error
	ListRoles(ctx context.Context, offset, limit int) ([]model.Role, int64, error)

	// Task operations
	CreateTask(ctx context.Context, task *model.Task) error
	GetTask(ctx context.Context, id uint) (*model.Task, error)
	GetTaskByUUID(ctx context.Context, uuid uuid.UUID) (*model.Task, error)
	GetTaskByMessageID(ctx context.Context, messageID string) (*model.Task, error)
	UpdateTask(ctx context.Context, task *model.Task) error
	DeleteTask(ctx context.Context, id uint) error
	ListTasks(ctx context.Context, offset, limit int) ([]model.Task, int64, error)
	ListTasksByStatus(ctx context.Context, status model.MessageStatus, offset, limit int) ([]model.Task, int64, error)
}
