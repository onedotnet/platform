// Package mq provides message queue middleware functionality
package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"github.com/onedotnet/platform/internal/model"
	"github.com/onedotnet/platform/internal/service"
	"github.com/onedotnet/platform/pkg/config"
	"github.com/onedotnet/platform/pkg/logger"
)

var (
	// Global MQ service for sending messages
	mqService     *RabbitMQService
	mqServiceOnce sync.Once
)

// RabbitMQService provides functionality to work with RabbitMQ
type RabbitMQService struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	config       config.RabbitMQConfig
	queueName    string
	repo         service.Repository
	nodeName     string
	nodeUUID     string
	exchangeName string
	mu           sync.Mutex
}

// NewRabbitMQService creates a new RabbitMQ service instance
func NewRabbitMQService(cfg config.RabbitMQConfig, repo service.Repository) (*RabbitMQService, error) {
	nodeName, err := os.Hostname()
	if err != nil {
		nodeName = "unknown"
	}

	nodeUUID := uuid.New().String()
	queueName := fmt.Sprintf("platform-%s-%s", nodeName, nodeUUID)

	// Create RabbitMQ service
	s := &RabbitMQService{
		config:       cfg,
		queueName:    queueName,
		repo:         repo,
		nodeName:     nodeName,
		nodeUUID:     nodeUUID,
		exchangeName: "platform-exchange",
	}

	// Connect to RabbitMQ server
	if err := s.Connect(); err != nil {
		return nil, err
	}

	return s, nil
}

// GetGlobalMQService returns the global instance of RabbitMQService
func GetGlobalMQService(cfg config.RabbitMQConfig, repo service.Repository) (*RabbitMQService, error) {
	var err error
	mqServiceOnce.Do(func() {
		mqService, err = NewRabbitMQService(cfg, repo)
	})
	return mqService, err
}

// Connect establishes connection to RabbitMQ and sets up the channel
func (s *RabbitMQService) Connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var err error
	connStr := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		s.config.Username, s.config.Password, s.config.Host, s.config.Port, s.config.VHost)

	s.conn, err = amqp.Dial(connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	s.channel, err = s.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %v", err)
	}

	// Declare exchange
	err = s.channel.ExchangeDeclare(
		s.exchangeName, // name
		"direct",       // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare an exchange: %v", err)
	}

	// Declare a queue with our unique node name
	_, err = s.channel.QueueDeclare(
		s.queueName, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %v", err)
	}

	// Bind the queue to the exchange with our node's routing key
	err = s.channel.QueueBind(
		s.queueName,    // queue name
		s.queueName,    // routing key
		s.exchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind a queue: %v", err)
	}

	logger.Log.Info("Connected to RabbitMQ",
		zap.String("host", s.config.Host),
		zap.Int("port", s.config.Port),
		zap.String("queue", s.queueName))

	return nil
}

// Close closes the RabbitMQ connection and channel
func (s *RabbitMQService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var err error
	if s.channel != nil {
		err = s.channel.Close()
		s.channel = nil
	}
	if s.conn != nil && !s.conn.IsClosed() {
		err = s.conn.Close()
		s.conn = nil
	}
	return err
}

// GetQueueName returns the unique queue name for this service instance
func (s *RabbitMQService) GetQueueName() string {
	return s.queueName
}

// SendMessage sends a message to the RabbitMQ queue and records it in the database
func (s *RabbitMQService) SendMessage(ctx context.Context, messageType string, body interface{}, needsAck bool) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if connection is alive, reconnect if needed
	if s.conn == nil || s.conn.IsClosed() {
		if err := s.Connect(); err != nil {
			return "", fmt.Errorf("failed to reconnect to RabbitMQ: %v", err)
		}
	}

	// Create message ID
	messageID := uuid.New().String()

	// Create base message
	baseMsg := model.BaseMessage{
		MessageID:        messageID,
		MessageType:      messageType,
		MessageCreatedAt: time.Now(),
		SentBy:           fmt.Sprintf("%s-%s", s.nodeName, s.nodeUUID),
		MessageBody:      body,
		MessageAck:       needsAck,
	}

	// Marshal message to JSON
	msgBytes, err := json.Marshal(baseMsg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message: %v", err)
	}

	// Save task to database
	task := model.Task{
		MessageID:   messageID,
		MessageType: messageType,
		QueueName:   s.queueName,
		Status:      model.MessageStatusPending,
		Payload:     string(msgBytes),
		SentBy:      baseMsg.SentBy,
	}

	if err := s.repo.CreateTask(ctx, &task); err != nil {
		return "", fmt.Errorf("failed to create task record: %v", err)
	}

	// Publishing context with timeout
	pubCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Publish message
	err = s.channel.PublishWithContext(
		pubCtx,
		s.exchangeName, // exchange
		s.queueName,    // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    messageID,
			Timestamp:    time.Now(),
			Body:         msgBytes,
		},
	)
	if err != nil {
		// Update task status
		task.Status = model.MessageStatusFailed
		task.Error = err.Error()
		if updateErr := s.repo.UpdateTask(ctx, &task); updateErr != nil {
			logger.Log.Error("Failed to update task status",
				zap.String("message_id", messageID),
				zap.Error(updateErr))
		}
		return "", fmt.Errorf("failed to publish message: %v", err)
	}

	// Update task status to delivered
	task.Status = model.MessageStatusDelivered
	if err := s.repo.UpdateTask(ctx, &task); err != nil {
		logger.Log.Error("Failed to update task status to delivered",
			zap.String("message_id", messageID),
			zap.Error(err))
	}

	logger.Log.Info("Message sent successfully",
		zap.String("message_id", messageID),
		zap.String("type", messageType),
		zap.String("queue", s.queueName))

	return messageID, nil
}

// AckMessage acknowledges a message by ID
func (s *RabbitMQService) AckMessage(ctx context.Context, messageID string, result interface{}) error {
	task, err := s.repo.GetTaskByMessageID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to find task: %v", err)
	}

	// Marshal result to JSON if provided
	if result != nil {
		resultBytes, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %v", err)
		}
		task.Result = string(resultBytes)
	}

	// Update task status
	now := time.Now()
	task.Status = model.MessageStatusAcked
	task.AckedAt = &now

	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("failed to update task status: %v", err)
	}

	logger.Log.Info("Message acknowledged successfully",
		zap.String("message_id", messageID))

	return nil
}

// SendMessage is a global function to send a message via RabbitMQ
func SendMessage(ctx context.Context, messageType string, body interface{}, needsAck bool) (string, error) {
	if mqService == nil {
		return "", fmt.Errorf("RabbitMQ service not initialized")
	}
	return mqService.SendMessage(ctx, messageType, body, needsAck)
}

// AckMessage is a global function to acknowledge a message
func AckMessage(ctx context.Context, messageID string, result interface{}) error {
	if mqService == nil {
		return fmt.Errorf("RabbitMQ service not initialized")
	}
	return mqService.AckMessage(ctx, messageID, result)
}
