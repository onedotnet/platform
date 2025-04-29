package model

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// Role model represents user roles for access control
type Role struct {
	Base
	Name        string `gorm:"uniqueIndex;size:50;not null" json:"name" validate:"required,min=2,max=50"`
	Description string `gorm:"size:255" json:"description" validate:"max=255"`
	Users       []User `gorm:"many2many:user_roles;" json:"users,omitempty"`
}

// Validate validates the Role model
func (r *Role) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// CacheKey returns the key used to store the object in cache
func (r *Role) CacheKey() string {
	return fmt.Sprintf("role:%s", r.UUID)
}
