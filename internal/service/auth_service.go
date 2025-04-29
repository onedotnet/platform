package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/onedotnet/platform/internal/model"
	"github.com/onedotnet/platform/pkg/config"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrInvalidCredentials is returned when the username or password is incorrect
	ErrInvalidCredentials = errors.New("invalid username or password")
	// ErrUserInactive is returned when the user account is inactive
	ErrUserInactive = errors.New("user account is inactive")
	// ErrInvalidToken is returned when the provided token is invalid
	ErrInvalidToken = errors.New("invalid or expired token")
	// ErrUserNotFound is returned when the user is not found
	ErrUserNotFound = errors.New("user not found")
)

// AuthClaims represents the claims in JWT token
type AuthClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	UUID     string `json:"uuid"`
	jwt.RegisteredClaims
}

// AuthService provides authentication-related operations
type AuthService interface {
	// Authenticate authenticates a user with username/email and password
	Authenticate(ctx context.Context, usernameOrEmail, password string) (*model.User, string, error)

	// AuthenticateWithProvider authenticates using an external provider
	AuthenticateWithProvider(ctx context.Context, provider model.AuthProvider, providerID, email string, userData map[string]interface{}) (*model.User, string, error)

	// VerifyToken verifies a JWT token and returns the claims
	VerifyToken(tokenString string) (*AuthClaims, error)

	// RefreshToken refreshes an existing token
	RefreshToken(ctx context.Context, refreshToken string) (string, error)

	// Register registers a new user
	Register(ctx context.Context, user *model.User) error

	// GetUserByID gets a user by ID
	GetUserByID(ctx context.Context, id uint) (*model.User, error)
}

// authService implements the AuthService interface
type authService struct {
	repo       Repository
	authConfig config.AuthConfig
}

// NewAuthService creates a new authentication service
func NewAuthService(repo Repository, authConfig config.AuthConfig) AuthService {
	return &authService{
		repo:       repo,
		authConfig: authConfig,
	}
}

// Authenticate authenticates a user with username/email and password
func (s *authService) Authenticate(ctx context.Context, usernameOrEmail, password string) (*model.User, string, error) {
	var user *model.User
	var err error

	// Check if input is an email or username
	if isEmail(usernameOrEmail) {
		user, err = s.repo.GetUserByEmail(ctx, usernameOrEmail)
	} else {
		user, err = s.repo.GetUserByUsername(ctx, usernameOrEmail)
	}

	if err != nil {
		return nil, "", ErrInvalidCredentials
	}

	// Check if user account is active
	if !user.Active {
		return nil, "", ErrUserInactive
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", err
	}

	// Update last login time
	now := time.Now()
	user.LastLoginAt = &now
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		// Non-critical error, still return token
		fmt.Printf("Failed to update last login time: %v\n", err)
	}

	return user, token, nil
}

// AuthenticateWithProvider authenticates using an external provider
func (s *authService) AuthenticateWithProvider(ctx context.Context, provider model.AuthProvider, providerID, email string, userData map[string]interface{}) (*model.User, string, error) {
	// Try to find existing user with this provider and ID
	users, _, err := s.repo.ListUsers(ctx, 0, 1000)
	var existingUser *model.User

	// This is inefficient but for the demo it's acceptable
	// In production, you would add a repository method to search by provider and providerID
	for i, u := range users {
		if u.Provider == provider && u.ProviderID == providerID {
			existingUser = &users[i]
			break
		}
	}

	// If no user found with provider ID, try to find by email
	if existingUser == nil && email != "" {
		existingUser, _ = s.repo.GetUserByEmail(ctx, email)
	}

	// Create a new user if no existing user found
	if existingUser == nil {
		// Generate a random username if none provided
		username, ok := userData["username"].(string)
		if !ok || username == "" {
			username = fmt.Sprintf("%s_user_%s", provider, uuid.New().String()[0:8])
		}

		firstName, _ := userData["first_name"].(string)
		lastName, _ := userData["last_name"].(string)
		avatarURL, _ := userData["avatar_url"].(string)

		newUser := &model.User{
			Username:   username,
			Email:      email,
			FirstName:  firstName,
			LastName:   lastName,
			Provider:   provider,
			ProviderID: providerID,
			AvatarURL:  avatarURL,
			Active:     true,
			Password:   "", // No password for social login
		}

		// Create user
		if err := s.repo.CreateUser(ctx, newUser); err != nil {
			return nil, "", fmt.Errorf("failed to create user: %w", err)
		}

		existingUser = newUser
	} else {
		// Update user information if needed
		existingUser.Provider = provider
		existingUser.ProviderID = providerID

		// Update other fields if available
		if firstName, ok := userData["first_name"].(string); ok && firstName != "" {
			existingUser.FirstName = firstName
		}
		if lastName, ok := userData["last_name"].(string); ok && lastName != "" {
			existingUser.LastName = lastName
		}
		if avatarURL, ok := userData["avatar_url"].(string); ok && avatarURL != "" {
			existingUser.AvatarURL = avatarURL
		}

		// Update user
		if err := s.repo.UpdateUser(ctx, existingUser); err != nil {
			// Non-critical error, continue
			fmt.Printf("Failed to update user: %v\n", err)
		}
	}

	// Check if user account is active
	if !existingUser.Active {
		return nil, "", ErrUserInactive
	}

	// Generate JWT token
	token, err := s.generateToken(existingUser)
	if err != nil {
		return nil, "", err
	}

	// Update last login time
	now := time.Now()
	existingUser.LastLoginAt = &now
	if err := s.repo.UpdateUser(ctx, existingUser); err != nil {
		// Non-critical error, still return token
		fmt.Printf("Failed to update last login time: %v\n", err)
	}

	return existingUser, token, nil
}

// VerifyToken verifies a JWT token and returns the claims
func (s *authService) VerifyToken(tokenString string) (*AuthClaims, error) {
	// Parse the token
	claims := &AuthClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			// Validate the algorithm
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(s.authConfig.JWTSecret), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshToken refreshes an existing token
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	// Find user by refresh token
	// This is a simplified implementation. In production, you should store refresh tokens in a separate table
	users, _, err := s.repo.ListUsers(ctx, 0, 1000)
	if err != nil {
		return "", err
	}

	var user *model.User
	for i, u := range users {
		if u.RefreshToken == refreshToken {
			user = &users[i]
			break
		}
	}

	if user == nil {
		return "", ErrInvalidToken
	}

	// Check if user account is active
	if !user.Active {
		return "", ErrUserInactive
	}

	// Generate a new JWT token
	token, err := s.generateToken(user)
	if err != nil {
		return "", err
	}

	return token, nil
}

// Register registers a new user
func (s *authService) Register(ctx context.Context, user *model.User) error {
	// Set default values
	user.Provider = model.AuthProviderLocal
	user.Active = true

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = string(hashedPassword)

	// Create user
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByID gets a user by ID
func (s *authService) GetUserByID(ctx context.Context, id uint) (*model.User, error) {
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// generateToken generates a JWT token for a user
func (s *authService) generateToken(user *model.User) (string, error) {
	// Create token with claims
	claims := &AuthClaims{
		UserID:   user.ID,
		Username: user.Username,
		UUID:     user.UUID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.authConfig.JWTExpirationTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	// Create token with signing method and claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate the token string
	tokenString, err := token.SignedString([]byte(s.authConfig.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	// Generate and store refresh token for the user
	refreshToken := uuid.New().String()
	user.RefreshToken = refreshToken

	return tokenString, nil
}

// Helper function to determine if a string is an email
func isEmail(s string) bool {
	// Simple check for @ symbol
	for i := 0; i < len(s); i++ {
		if s[i] == '@' {
			return true
		}
	}
	return false
}
