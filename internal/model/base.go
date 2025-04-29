package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base contains common columns for all tables
type Base struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id" comment:"Auto-incrementing sortable ID"`
	UUID      uuid.UUID      `gorm:"type:uuid;uniqueIndex;not null;default:gen_random_uuid()" json:"uuid" comment:"Unique identifier"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate will set a UUID rather than numeric ID as the primary key
func (base *Base) BeforeCreate(tx *gorm.DB) error {
	if base.UUID == uuid.Nil {
		base.UUID = uuid.New()
	}
	return nil
}
