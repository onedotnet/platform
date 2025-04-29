package service

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/onedotnet/platform/internal/cache"
	"github.com/onedotnet/platform/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// RepositoryTestSuite defines the test suite for repository
type RepositoryTestSuite struct {
	suite.Suite
	DB         *gorm.DB
	mock       sqlmock.Sqlmock
	repository Repository
	cache      cache.Cache
}

// SetupTest is called before each test
func (s *RepositoryTestSuite) SetupTest() {
	var (
		db  *sql.DB
		err error
	)

	// Create new mock database
	db, s.mock, err = sqlmock.New()
	require.NoError(s.T(), err)

	// Create GORM DB with the mock SQL driver
	dialector := postgres.New(postgres.Config{
		DSN:                  "sqlmock_db_0",
		DriverName:           "postgres",
		Conn:                 db,
		PreferSimpleProtocol: true,
	})

	s.DB, err = gorm.Open(dialector, &gorm.Config{})
	require.NoError(s.T(), err)

	// Create mock cache
	s.cache = cache.NewMockCache()

	// Create repository with mock DB and mock cache
	s.repository = NewGormRepository(s.DB, s.cache)
}

// TestUserCRUD tests user CRUD operations
func (s *RepositoryTestSuite) TestUserCRUD() {
	t := s.T()
	ctx := context.Background()

	// Test data
	now := time.Now()
	mockUUID := uuid.New()
	userID := uint(1)
	user := &model.User{
		Base: model.Base{
			ID:        userID,
			UUID:      mockUUID,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
		Phone:     "1234567890",
		Active:    true,
	}

	// Test Create
	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
		WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			user.Username,
			user.Email,
			user.Password,
			user.FirstName,
			user.LastName,
			user.Phone,
			user.Active,
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))
	s.mock.ExpectCommit()

	err := s.repository.CreateUser(ctx, user)
	assert.NoError(t, err)
	assert.Equal(t, userID, user.ID)

	// Test Get
	rows := sqlmock.NewRows([]string{"id", "uuid", "created_at", "updated_at", "deleted_at", "username", "email", "password", "first_name", "last_name", "phone", "active", "last_login_at"}).
		AddRow(user.ID, user.UUID, user.CreatedAt, user.UpdatedAt, nil, user.Username, user.Email, user.Password, user.FirstName, user.LastName, user.Phone, user.Active, nil)
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`)).
		WithArgs(userID).
		WillReturnRows(rows)

	retrievedUser, err := s.repository.GetUser(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, user.Username, retrievedUser.Username)
	assert.Equal(t, user.Email, retrievedUser.Email)

	// Test GetByUsername
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`)).
		WithArgs(user.Username).
		WillReturnRows(rows)

	userByUsername, err := s.repository.GetUserByUsername(ctx, user.Username)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, userByUsername.ID)
	assert.Equal(t, user.Username, userByUsername.Username)

	// Test GetByEmail
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`)).
		WithArgs(user.Email).
		WillReturnRows(rows)

	userByEmail, err := s.repository.GetUserByEmail(ctx, user.Email)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, userByEmail.ID)
	assert.Equal(t, user.Email, userByEmail.Email)

	// Test GetByUUID
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE uuid = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`)).
		WithArgs(user.UUID).
		WillReturnRows(rows)

	userByUUID, err := s.repository.GetUserByUUID(ctx, user.UUID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, userByUUID.ID)
	assert.Equal(t, user.UUID, userByUUID.UUID)

	// Test Update
	updatedUser := *user
	updatedUser.FirstName = "Updated"

	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET`)).
		WithArgs(
			sqlmock.AnyArg(), // uuid
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
			sqlmock.AnyArg(), // deleted_at
			updatedUser.Username,
			updatedUser.Email,
			updatedUser.Password,
			updatedUser.FirstName,
			updatedUser.LastName,
			updatedUser.Phone,
			updatedUser.Active,
			updatedUser.LastLoginAt,
			updatedUser.ID,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	err = s.repository.UpdateUser(ctx, &updatedUser)
	assert.NoError(t, err)

	// Test List
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "users" WHERE "users"."deleted_at" IS NULL`)).
		WillReturnRows(countRows)

	listRows := sqlmock.NewRows([]string{"id", "uuid", "created_at", "updated_at", "deleted_at", "username", "email", "password", "first_name", "last_name", "phone", "active", "last_login_at"}).
		AddRow(user.ID, user.UUID, user.CreatedAt, user.UpdatedAt, nil, user.Username, user.Email, user.Password, user.FirstName, user.LastName, user.Phone, user.Active, nil)
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."deleted_at" IS NULL LIMIT 10 OFFSET 0`)).
		WillReturnRows(listRows)

	users, count, err := s.repository.ListUsers(ctx, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
	assert.Equal(t, 1, len(users))
	assert.Equal(t, userID, users[0].ID)

	// Test Delete
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`)).
		WithArgs(userID).
		WillReturnRows(rows)

	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET "deleted_at"=$1 WHERE "users"."id" = $2 AND "users"."deleted_at" IS NULL`)).
		WithArgs(sqlmock.AnyArg(), userID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	err = s.repository.DeleteUser(ctx, userID)
	assert.NoError(t, err)

	// Test Get after deletion (should return error)
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`)).
		WithArgs(userID).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err = s.repository.GetUser(ctx, userID)
	assert.Error(t, err)
}

// TestOrganizationCRUD tests organization CRUD operations
func (s *RepositoryTestSuite) TestOrganizationCRUD() {
	t := s.T()
	ctx := context.Background()

	// Test data
	now := time.Now()
	mockUUID := uuid.New()
	orgID := uint(1)
	org := &model.Organization{
		Base: model.Base{
			ID:        orgID,
			UUID:      mockUUID,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Name:        "testorg",
		DisplayName: "Test Organization",
		Description: "A test organization",
		Website:     "https://example.com",
		Active:      true,
	}

	// Test Create
	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "organizations"`)).
		WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			org.Name,
			org.DisplayName,
			org.Description,
			org.LogoURL,
			org.Website,
			org.Active,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(orgID))
	s.mock.ExpectCommit()

	err := s.repository.CreateOrganization(ctx, org)
	assert.NoError(t, err)
	assert.Equal(t, orgID, org.ID)

	// Test Get
	rows := sqlmock.NewRows([]string{"id", "uuid", "created_at", "updated_at", "deleted_at", "name", "display_name", "description", "logo_url", "website", "active"}).
		AddRow(org.ID, org.UUID, org.CreatedAt, org.UpdatedAt, nil, org.Name, org.DisplayName, org.Description, org.LogoURL, org.Website, org.Active)
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE "organizations"."id" = $1 AND "organizations"."deleted_at" IS NULL ORDER BY "organizations"."id" LIMIT 1`)).
		WithArgs(orgID).
		WillReturnRows(rows)

	retrievedOrg, err := s.repository.GetOrganization(ctx, orgID)
	assert.NoError(t, err)
	assert.Equal(t, org.Name, retrievedOrg.Name)
	assert.Equal(t, org.DisplayName, retrievedOrg.DisplayName)

	// Test GetByName
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE name = $1 AND "organizations"."deleted_at" IS NULL ORDER BY "organizations"."id" LIMIT 1`)).
		WithArgs(org.Name).
		WillReturnRows(rows)

	orgByName, err := s.repository.GetOrganizationByName(ctx, org.Name)
	assert.NoError(t, err)
	assert.Equal(t, org.ID, orgByName.ID)
	assert.Equal(t, org.Name, orgByName.Name)

	// Test Update
	updatedOrg := *org
	updatedOrg.DisplayName = "Updated Organization"

	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "organizations" SET`)).
		WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			updatedOrg.Name,
			updatedOrg.DisplayName,
			updatedOrg.Description,
			updatedOrg.LogoURL,
			updatedOrg.Website,
			updatedOrg.Active,
			updatedOrg.ID,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	err = s.repository.UpdateOrganization(ctx, &updatedOrg)
	assert.NoError(t, err)

	// Test List
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "organizations" WHERE "organizations"."deleted_at" IS NULL`)).
		WillReturnRows(countRows)

	listRows := sqlmock.NewRows([]string{"id", "uuid", "created_at", "updated_at", "deleted_at", "name", "display_name", "description", "logo_url", "website", "active"}).
		AddRow(org.ID, org.UUID, org.CreatedAt, org.UpdatedAt, nil, org.Name, org.DisplayName, org.Description, org.LogoURL, org.Website, org.Active)
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE "organizations"."deleted_at" IS NULL LIMIT 10 OFFSET 0`)).
		WillReturnRows(listRows)

	orgs, count, err := s.repository.ListOrganizations(ctx, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
	assert.Equal(t, 1, len(orgs))
	assert.Equal(t, orgID, orgs[0].ID)

	// Test Delete
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE "organizations"."id" = $1 AND "organizations"."deleted_at" IS NULL ORDER BY "organizations"."id" LIMIT 1`)).
		WithArgs(orgID).
		WillReturnRows(rows)

	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "organizations" SET "deleted_at"=$1 WHERE "organizations"."id" = $2 AND "organizations"."deleted_at" IS NULL`)).
		WithArgs(sqlmock.AnyArg(), orgID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	err = s.repository.DeleteOrganization(ctx, orgID)
	assert.NoError(t, err)
}

// TestRoleCRUD tests role CRUD operations
func (s *RepositoryTestSuite) TestRoleCRUD() {
	t := s.T()
	ctx := context.Background()

	// Test data
	now := time.Now()
	mockUUID := uuid.New()
	roleID := uint(1)
	role := &model.Role{
		Base: model.Base{
			ID:        roleID,
			UUID:      mockUUID,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Name:        "testrole",
		Description: "A test role",
	}

	// Test Create
	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "roles"`)).
		WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			role.Name,
			role.Description,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(roleID))
	s.mock.ExpectCommit()

	err := s.repository.CreateRole(ctx, role)
	assert.NoError(t, err)
	assert.Equal(t, roleID, role.ID)

	// Test Get
	rows := sqlmock.NewRows([]string{"id", "uuid", "created_at", "updated_at", "deleted_at", "name", "description"}).
		AddRow(role.ID, role.UUID, role.CreatedAt, role.UpdatedAt, nil, role.Name, role.Description)
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "roles" WHERE "roles"."id" = $1 AND "roles"."deleted_at" IS NULL ORDER BY "roles"."id" LIMIT 1`)).
		WithArgs(roleID).
		WillReturnRows(rows)

	retrievedRole, err := s.repository.GetRole(ctx, roleID)
	assert.NoError(t, err)
	assert.Equal(t, role.Name, retrievedRole.Name)
	assert.Equal(t, role.Description, retrievedRole.Description)

	// Test GetByName
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "roles" WHERE name = $1 AND "roles"."deleted_at" IS NULL ORDER BY "roles"."id" LIMIT 1`)).
		WithArgs(role.Name).
		WillReturnRows(rows)

	roleByName, err := s.repository.GetRoleByName(ctx, role.Name)
	assert.NoError(t, err)
	assert.Equal(t, role.ID, roleByName.ID)
	assert.Equal(t, role.Name, roleByName.Name)

	// Test Update
	updatedRole := *role
	updatedRole.Description = "Updated role description"

	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "roles" SET`)).
		WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			updatedRole.Name,
			updatedRole.Description,
			updatedRole.ID,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	err = s.repository.UpdateRole(ctx, &updatedRole)
	assert.NoError(t, err)

	// Test List
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "roles" WHERE "roles"."deleted_at" IS NULL`)).
		WillReturnRows(countRows)

	listRows := sqlmock.NewRows([]string{"id", "uuid", "created_at", "updated_at", "deleted_at", "name", "description"}).
		AddRow(role.ID, role.UUID, role.CreatedAt, role.UpdatedAt, nil, role.Name, role.Description)
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "roles" WHERE "roles"."deleted_at" IS NULL LIMIT 10 OFFSET 0`)).
		WillReturnRows(listRows)

	roles, count, err := s.repository.ListRoles(ctx, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
	assert.Equal(t, 1, len(roles))
	assert.Equal(t, roleID, roles[0].ID)

	// Test Delete
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "roles" WHERE "roles"."id" = $1 AND "roles"."deleted_at" IS NULL ORDER BY "roles"."id" LIMIT 1`)).
		WithArgs(roleID).
		WillReturnRows(rows)

	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "roles" SET "deleted_at"=$1 WHERE "roles"."id" = $2 AND "roles"."deleted_at" IS NULL`)).
		WithArgs(sqlmock.AnyArg(), roleID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	err = s.repository.DeleteRole(ctx, roleID)
	assert.NoError(t, err)
}

// TestUserValidation tests user validation
func (s *RepositoryTestSuite) TestUserValidation() {
	t := s.T()
	ctx := context.Background()

	// Create a user with invalid data
	invalidUser := &model.User{
		Username:  "u", // too short, validation should fail
		Email:     "notanemail",
		Password:  "short",
		FirstName: "Test",
		LastName:  "User",
	}

	// Test validation on create - no DB interaction should occur since validation fails first
	err := s.repository.CreateUser(ctx, invalidUser)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation error")

	// Create a valid user
	userID := uint(1)
	validUser := &model.User{
		Base: model.Base{
			ID: userID,
		},
		Username:  "validuser",
		Email:     "valid@example.com",
		Password:  "password123",
		FirstName: "Valid",
		LastName:  "User",
		Phone:     "1234567890",
		Active:    true,
	}

	// Setup DB mock for valid user
	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
		WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			validUser.Username,
			validUser.Email,
			validUser.Password,
			validUser.FirstName,
			validUser.LastName,
			validUser.Phone,
			validUser.Active,
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))
	s.mock.ExpectCommit()

	err = s.repository.CreateUser(ctx, validUser)
	assert.NoError(t, err)

	// Update with invalid data - again, no DB interaction since validation fails
	validUser.Email = "invalid"
	err = s.repository.UpdateUser(ctx, validUser)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation error")
}

// TestUUIDOperations tests the UUID-related operations for all models
func (s *RepositoryTestSuite) TestUUIDOperations() {
	t := s.T()
	ctx := context.Background()

	// Test data for User
	userUUID := uuid.New()
	userID := uint(1)
	user := &model.User{
		Base: model.Base{
			ID:        userID,
			UUID:      userUUID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
		Phone:     "1234567890",
		Active:    true,
	}

	// Test GetUserByUUID
	userRows := sqlmock.NewRows([]string{"id", "uuid", "created_at", "updated_at", "deleted_at", "username", "email", "password", "first_name", "last_name", "phone", "active", "last_login_at"}).
		AddRow(user.ID, user.UUID, user.CreatedAt, user.UpdatedAt, nil, user.Username, user.Email, user.Password, user.FirstName, user.LastName, user.Phone, user.Active, nil)
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE uuid = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`)).
		WithArgs(user.UUID).
		WillReturnRows(userRows)

	userByUUID, err := s.repository.GetUserByUUID(ctx, user.UUID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, userByUUID.ID)
	assert.Equal(t, user.UUID, userByUUID.UUID)
	assert.Equal(t, user.Username, userByUUID.Username)

	// Test DeleteUserByUUID
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE uuid = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1`)).
		WithArgs(user.UUID).
		WillReturnRows(userRows)

	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET "deleted_at"=$1 WHERE "users"."id" = $2 AND "users"."deleted_at" IS NULL`)).
		WithArgs(sqlmock.AnyArg(), user.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	err = s.repository.DeleteUserByUUID(ctx, user.UUID)
	assert.NoError(t, err)

	// Test data for Organization
	orgUUID := uuid.New()
	orgID := uint(1)
	org := &model.Organization{
		Base: model.Base{
			ID:        orgID,
			UUID:      orgUUID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "testorg",
		DisplayName: "Test Organization",
		Description: "A test organization",
		Website:     "https://example.com",
		Active:      true,
	}

	// Test GetOrganizationByUUID
	orgRows := sqlmock.NewRows([]string{"id", "uuid", "created_at", "updated_at", "deleted_at", "name", "display_name", "description", "logo_url", "website", "active"}).
		AddRow(org.ID, org.UUID, org.CreatedAt, org.UpdatedAt, nil, org.Name, org.DisplayName, org.Description, org.LogoURL, org.Website, org.Active)
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE uuid = $1 AND "organizations"."deleted_at" IS NULL ORDER BY "organizations"."id" LIMIT 1`)).
		WithArgs(org.UUID).
		WillReturnRows(orgRows)

	orgByUUID, err := s.repository.GetOrganizationByUUID(ctx, org.UUID)
	assert.NoError(t, err)
	assert.Equal(t, org.ID, orgByUUID.ID)
	assert.Equal(t, org.UUID, orgByUUID.UUID)
	assert.Equal(t, org.Name, orgByUUID.Name)

	// Test DeleteOrganizationByUUID
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE uuid = $1 AND "organizations"."deleted_at" IS NULL ORDER BY "organizations"."id" LIMIT 1`)).
		WithArgs(org.UUID).
		WillReturnRows(orgRows)

	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "organizations" SET "deleted_at"=$1 WHERE "organizations"."id" = $2 AND "organizations"."deleted_at" IS NULL`)).
		WithArgs(sqlmock.AnyArg(), org.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	err = s.repository.DeleteOrganizationByUUID(ctx, org.UUID)
	assert.NoError(t, err)

	// Test data for Role
	roleUUID := uuid.New()
	roleID := uint(1)
	role := &model.Role{
		Base: model.Base{
			ID:        roleID,
			UUID:      roleUUID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "testrole",
		Description: "A test role",
	}

	// Test GetRoleByUUID
	roleRows := sqlmock.NewRows([]string{"id", "uuid", "created_at", "updated_at", "deleted_at", "name", "description"}).
		AddRow(role.ID, role.UUID, role.CreatedAt, role.UpdatedAt, nil, role.Name, role.Description)
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "roles" WHERE uuid = $1 AND "roles"."deleted_at" IS NULL ORDER BY "roles"."id" LIMIT 1`)).
		WithArgs(role.UUID).
		WillReturnRows(roleRows)

	roleByUUID, err := s.repository.GetRoleByUUID(ctx, role.UUID)
	assert.NoError(t, err)
	assert.Equal(t, role.ID, roleByUUID.ID)
	assert.Equal(t, role.UUID, roleByUUID.UUID)
	assert.Equal(t, role.Name, roleByUUID.Name)

	// Test DeleteRoleByUUID
	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "roles" WHERE uuid = $1 AND "roles"."deleted_at" IS NULL ORDER BY "roles"."id" LIMIT 1`)).
		WithArgs(role.UUID).
		WillReturnRows(roleRows)

	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "roles" SET "deleted_at"=$1 WHERE "roles"."id" = $2 AND "roles"."deleted_at" IS NULL`)).
		WithArgs(sqlmock.AnyArg(), role.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	err = s.repository.DeleteRoleByUUID(ctx, role.UUID)
	assert.NoError(t, err)
}

// TestSuite executes the test suite
func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

// Test error conditions
func TestRepositoryErrors(t *testing.T) {
	// Create a separate test for error conditions
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	dialector := postgres.New(postgres.Config{
		DSN:                  "sqlmock_db_0",
		DriverName:           "postgres",
		Conn:                 db,
		PreferSimpleProtocol: true,
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	mockCache := cache.NewMockCache()
	repo := NewGormRepository(gormDB, mockCache)

	ctx := context.Background()

	t.Run("GetUserNotFound", func(t *testing.T) {
		// Test Get user when not found
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
			WithArgs(uint(999)).
			WillReturnError(gorm.ErrRecordNotFound)

		user, err := repo.GetUser(ctx, uint(999))
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound) ||
			err.Error() == "user not found: 999")
	})

	t.Run("CreateUserDBError", func(t *testing.T) {
		user := &model.User{
			Username:  "testuser",
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
			Phone:     "1234567890",
			Active:    true,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
			WillReturnError(errors.New("database error"))
		mock.ExpectRollback()

		err := repo.CreateUser(ctx, user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
	})
}
