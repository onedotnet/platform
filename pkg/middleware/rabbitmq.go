// Package middleware provides HTTP middleware and other service middleware functionalities
package middleware

import (
	"context"

	"github.com/onedotnet/platform/internal/service"
	"github.com/onedotnet/platform/pkg/config"
	"github.com/onedotnet/platform/pkg/middleware/mq"
)

// RabbitMQService alias for backward compatibility
type RabbitMQService = mq.RabbitMQService

// NewRabbitMQService creates a new RabbitMQ service instance (compatibility function)
// Deprecated: Use mq.NewRabbitMQService instead
func NewRabbitMQService(cfg config.RabbitMQConfig, repo service.Repository) (*RabbitMQService, error) {
	return mq.NewRabbitMQService(cfg, repo)
}

// GetGlobalMQService returns the global instance of RabbitMQService (compatibility function)
// Deprecated: Use mq.GetGlobalMQService instead
func GetGlobalMQService(cfg config.RabbitMQConfig, repo service.Repository) (*RabbitMQService, error) {
	return mq.GetGlobalMQService(cfg, repo)
}

// SendMessage is a global function to send a message via RabbitMQ (compatibility function)
// Deprecated: Use mq.SendMessage instead
func SendMessage(ctx context.Context, messageType string, body interface{}, needsAck bool) (string, error) {
	return mq.SendMessage(ctx, messageType, body, needsAck)
}

// AckMessage is a global function to acknowledge a message (compatibility function)
// Deprecated: Use mq.AckMessage instead
func AckMessage(ctx context.Context, messageID string, result interface{}) error {
	return mq.AckMessage(ctx, messageID, result)
}
