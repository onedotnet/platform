package mq

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/onedotnet/platform/internal/model"
	"github.com/onedotnet/platform/pkg/config"
)

// MockRepository is a mock implementation of the Repository interface
type MockRepository struct {
	mock.Mock
}

// CreateTask is a mock implementation
func (m *MockRepository) CreateTask(ctx context.Context, task *model.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

// GetTaskByMessageID is a mock implementation
func (m *MockRepository) GetTaskByMessageID(ctx context.Context, messageID string) (*model.Task, error) {
	args := m.Called(ctx, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

// UpdateTask is a mock implementation
func (m *MockRepository) UpdateTask(ctx context.Context, task *model.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

// Implement other Repository interface methods as needed for tests
func (m *MockRepository) GetTask(ctx context.Context, id uint) (*model.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *MockRepository) GetTaskByUUID(ctx context.Context, uuid uuid.UUID) (*model.Task, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *MockRepository) DeleteTask(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) ListTasks(ctx context.Context, offset, limit int) ([]model.Task, int64, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]model.Task), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) ListTasksByStatus(ctx context.Context, status model.MessageStatus, offset, limit int) ([]model.Task, int64, error) {
	args := m.Called(ctx, status, offset, limit)
	return args.Get(0).([]model.Task), args.Get(1).(int64), args.Error(2)
}

// Methods required by service.Repository interface
func (m *MockRepository) CreateUser(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) GetUser(ctx context.Context, id uint) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockRepository) GetUserByUUID(ctx context.Context, uuid uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockRepository) UpdateUser(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) DeleteUser(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) DeleteUserByUUID(ctx context.Context, uuid uuid.UUID) error {
	args := m.Called(ctx, uuid)
	return args.Error(0)
}

func (m *MockRepository) ListUsers(ctx context.Context, offset, limit int) ([]model.User, int64, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]model.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) CreateOrganization(ctx context.Context, org *model.Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *MockRepository) GetOrganization(ctx context.Context, id uint) (*model.Organization, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Organization), args.Error(1)
}

func (m *MockRepository) GetOrganizationByUUID(ctx context.Context, uuid uuid.UUID) (*model.Organization, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Organization), args.Error(1)
}

func (m *MockRepository) GetOrganizationByName(ctx context.Context, name string) (*model.Organization, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Organization), args.Error(1)
}

func (m *MockRepository) UpdateOrganization(ctx context.Context, org *model.Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *MockRepository) DeleteOrganization(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) DeleteOrganizationByUUID(ctx context.Context, uuid uuid.UUID) error {
	args := m.Called(ctx, uuid)
	return args.Error(0)
}

func (m *MockRepository) ListOrganizations(ctx context.Context, offset, limit int) ([]model.Organization, int64, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]model.Organization), args.Get(1).(int64), args.Error(2)
}

// Methods required by service.Repository interface for Role operations
func (m *MockRepository) CreateRole(ctx context.Context, role *model.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRepository) GetRole(ctx context.Context, id uint) (*model.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRepository) GetRoleByUUID(ctx context.Context, uuid uuid.UUID) (*model.Role, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRepository) GetRoleByName(ctx context.Context, name string) (*model.Role, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRepository) UpdateRole(ctx context.Context, role *model.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRepository) DeleteRole(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) DeleteRoleByUUID(ctx context.Context, uuid uuid.UUID) error {
	args := m.Called(ctx, uuid)
	return args.Error(0)
}

func (m *MockRepository) ListRoles(ctx context.Context, offset, limit int) ([]model.Role, int64, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]model.Role), args.Get(1).(int64), args.Error(2)
}

// TestSendMessage tests the SendMessage function
func TestSendMessage(t *testing.T) {
	// Create mock repository
	mockRepo := new(MockRepository)

	// Set up test config
	cfg := config.RabbitMQConfig{
		Host:     "localhost", // Use fake values for testing
		Port:     5672,
		Username: "guest",
		Password: "guest",
		VHost:    "/",
	}

	// Create test message
	type TestMessage struct {
		Content string `json:"content"`
	}
	testMsg := TestMessage{Content: "Hello, World!"}

	// Set up mock expectations
	mockRepo.On("CreateTask", mock.Anything, mock.AnythingOfType("*model.Task")).Return(nil)
	mockRepo.On("UpdateTask", mock.Anything, mock.AnythingOfType("*model.Task")).Return(nil)

	// Create the service with our mock
	mockService := &RabbitMQService{
		config:       cfg,
		repo:         mockRepo,
		queueName:    "test-queue",
		nodeName:     "test-node",
		nodeUUID:     "test-uuid",
		exchangeName: "test-exchange",
	}

	// Replace global service for testing
	origService := mqService
	mqService = mockService
	defer func() { mqService = origService }()

	// Test without actual RabbitMQ connection
	mockService.conn = nil // Force a reconnect error

	// Test sending a message (should fail without a real connection)
	ctx := context.Background()
	messageID, err := SendMessage(ctx, "test", testMsg, true)

	// Since we're not actually connecting to RabbitMQ, we expect an error
	assert.Error(t, err)
	assert.Empty(t, messageID)

	// Verify that the task was created but failed
	mockRepo.AssertCalled(t, "CreateTask", mock.Anything, mock.AnythingOfType("*model.Task"))
}

// TestMessageSerialization tests that the message serialization works correctly
func TestMessageSerialization(t *testing.T) {
	// Create a message
	messageID := "test-message-id"
	messageType := "test-type"
	sentBy := "test-node"
	messageBody := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	baseMsg := model.BaseMessage{
		MessageID:        messageID,
		MessageType:      messageType,
		MessageCreatedAt: time.Now(),
		SentBy:           sentBy,
		MessageBody:      messageBody,
		MessageAck:       true,
	}

	// Convert to JSON
	bytes, err := json.Marshal(baseMsg)
	assert.NoError(t, err)

	// Convert back
	var decodedMsg model.BaseMessage
	err = json.Unmarshal(bytes, &decodedMsg)
	assert.NoError(t, err)

	// Check the values
	assert.Equal(t, messageID, decodedMsg.MessageID)
	assert.Equal(t, messageType, decodedMsg.MessageType)
	assert.Equal(t, sentBy, decodedMsg.SentBy)
	assert.Equal(t, true, decodedMsg.MessageAck)

	// Check time serialization (ignoring subsecond precision)
	assert.WithinDuration(t, baseMsg.MessageCreatedAt, decodedMsg.MessageCreatedAt, time.Second)

	// Check message body (need to convert to map)
	bodyMap, ok := decodedMsg.MessageBody.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "value1", bodyMap["key1"])
	assert.Equal(t, float64(123), bodyMap["key2"]) // JSON converts numbers to float64
}

// TestAckMessage tests the AckMessage function
func TestAckMessage(t *testing.T) {
	// Create mock repository
	mockRepo := new(MockRepository)

	// Create a task for the mock to return
	task := &model.Task{
		MessageID: "test-message-id",
		Status:    model.MessageStatusDelivered,
		QueueName: "test-queue",
		Payload:   `{"message_id":"test-message-id","message_type":"test","message_body":"test"}`,
	}

	// Set up mock expectations
	mockRepo.On("GetTaskByMessageID", mock.Anything, "test-message-id").Return(task, nil)
	mockRepo.On("UpdateTask", mock.Anything, mock.AnythingOfType("*model.Task")).Run(func(args mock.Arguments) {
		updatedTask := args.Get(1).(*model.Task)
		// Verify the task was updated correctly
		assert.Equal(t, model.MessageStatusAcked, updatedTask.Status)
		assert.NotNil(t, updatedTask.AckedAt)
	}).Return(nil)

	// Create the service with our mock
	mockService := &RabbitMQService{
		repo:      mockRepo,
		queueName: "test-queue",
	}

	// Replace global service for testing
	origService := mqService
	mqService = mockService
	defer func() { mqService = origService }()

	// Test acknowledging a message
	ctx := context.Background()
	resultData := map[string]string{"result": "success"}
	err := AckMessage(ctx, "test-message-id", resultData)

	// Verify the result
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
