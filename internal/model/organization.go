package model

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// Organization model represents organization information
type Organization struct {
	Base
	Name        string `gorm:"uniqueIndex;size:100;not null" json:"name" validate:"required,min=2,max=100"`
	DisplayName string `gorm:"size:150" json:"display_name" validate:"max=150"`
	Description string `gorm:"size:500" json:"description" validate:"max=500"`
	LogoURL     string `gorm:"size:255" json:"logo_url" validate:"omitempty,url"`
	Website     string `gorm:"size:255" json:"website" validate:"omitempty,url"`
	Active      bool   `gorm:"default:true" json:"active"`
	Users       []User `gorm:"many2many:user_organizations;" json:"users,omitempty"`
}

// Validate validates the Organization model
func (o *Organization) Validate() error {
	validate := validator.New()
	return validate.Struct(o)
}

// CacheKey returns the key used to store the object in cache
func (o *Organization) CacheKey() string {
	return fmt.Sprintf("organization:%s", o.UUID)
}
