package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageStatus represents the status of a message task
type MessageStatus string

const (
	MessageStatusPending   MessageStatus = "pending"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusFailed    MessageStatus = "failed"
	MessageStatusAcked     MessageStatus = "acknowledged"
)

// BaseMessage is the core structure for all messages
type BaseMessage struct {
	MessageID        string      `json:"message_id"`
	MessageType      string      `json:"message_type"`
	MessageCreatedAt time.Time   `json:"message_created_time"`
	SentBy           string      `json:"sentby"`
	MessageBody      interface{} `json:"message_body"`
	MessageAck       bool        `json:"message_ack"`
}

// Task represents a message task record in the database
type Task struct {
	gorm.Model
	UUID          uuid.UUID     `gorm:"type:uuid;index;not null" json:"uuid"`
	MessageID     string        `gorm:"size:128;index;not null" json:"message_id"`
	MessageType   string        `gorm:"size:128;not null" json:"message_type"`
	QueueName     string        `gorm:"size:255;not null" json:"queue_name"`
	Status        MessageStatus `gorm:"size:20;not null;default:'pending'" json:"status"`
	RetryCount    int           `gorm:"default:0" json:"retry_count"`
	NextRetryTime *time.Time    `json:"next_retry_time,omitempty"`
	Error         string        `gorm:"size:512" json:"error,omitempty"`
	Payload       string        `gorm:"type:text" json:"payload"`
	Result        string        `gorm:"type:text" json:"result,omitempty"`
	SentBy        string        `gorm:"size:128" json:"sent_by"`
	AckedAt       *time.Time    `json:"acked_at,omitempty"`
}

// BeforeCreate will set UUID before creating
func (t *Task) BeforeCreate(tx *gorm.DB) (err error) {
	if t.UUID == uuid.Nil {
		t.UUID = uuid.New()
	}
	return
}
