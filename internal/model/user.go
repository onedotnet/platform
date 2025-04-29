package model

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

// AuthProvider represents the authentication provider type
type AuthProvider string

const (
	// AuthProviderLocal represents local username/password authentication
	AuthProviderLocal AuthProvider = "local"
	// AuthProviderGoogle represents Google OAuth authentication
	AuthProviderGoogle AuthProvider = "google"
	// AuthProviderMicrosoft represents Microsoft Entra ID authentication
	AuthProviderMicrosoft AuthProvider = "microsoft"
	// AuthProviderGitHub represents GitHub OAuth authentication
	AuthProviderGitHub AuthProvider = "github"
	// AuthProviderWeChat represents WeChat authentication
	AuthProviderWeChat AuthProvider = "wechat"
)

// User model represents user information
type User struct {
	Base
	Username      string         `gorm:"uniqueIndex;size:50;not null" json:"username" validate:"required,min=3,max=50"`
	Email         string         `gorm:"uniqueIndex;size:100;not null" json:"email" validate:"required,email"`
	Password      string         `gorm:"size:255;not null" json:"-" validate:"omitempty,min=8"`
	FirstName     string         `gorm:"size:50" json:"first_name" validate:"max=50"`
	LastName      string         `gorm:"size:50" json:"last_name" validate:"max=50"`
	Phone         string         `gorm:"size:20" json:"phone" validate:"omitempty,min=10,max=20"`
	Active        bool           `gorm:"default:true" json:"active"`
	LastLoginAt   *time.Time     `json:"last_login_at,omitempty"`
	Provider      AuthProvider   `gorm:"size:20;default:'local'" json:"provider"`
	ProviderID    string         `gorm:"size:255" json:"provider_id,omitempty"`
	RefreshToken  string         `gorm:"size:255" json:"-"`
	AvatarURL     string         `gorm:"size:255" json:"avatar_url,omitempty"`
	Organizations []Organization `gorm:"many2many:user_organizations;" json:"organizations,omitempty"`
	Roles         []Role         `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

// Validate validates the User model
func (u *User) Validate() error {
	validate := validator.New()

	// If using social login, password is not required
	if u.Provider != AuthProviderLocal {
		return validate.StructExcept(u, "Password")
	}

	return validate.Struct(u)
}

// CacheKey returns the key used to store the object in cache
func (u *User) CacheKey() string {
	return fmt.Sprintf("user:%s", u.UUID)
}
