package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/onedotnet/platform/internal/model"
	"github.com/onedotnet/platform/internal/service"
	"github.com/onedotnet/platform/pkg/config"
)

// LoginRequest represents the login request body
type LoginRequest struct {
	UsernameOrEmail string `json:"username_or_email" binding:"required"`
	Password        string `json:"password" binding:"required"`
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=50"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

// TokenResponse represents the token response
type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expires_in"` // in seconds
}

// AuthHandler handles HTTP requests related to authentication
type AuthHandler struct {
	authService service.AuthService
	authConfig  config.AuthConfig
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService service.AuthService, authConfig config.AuthConfig) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		authConfig:  authConfig,
	}
}

// Register registers authentication routes with the provided router
func (h *AuthHandler) Register(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		// Standard authentication endpoints
		auth.POST("/login", h.Login)
		auth.POST("/register", h.RegisterUser)
		auth.POST("/refresh-token", h.RefreshToken)

		// OAuth endpoints
		auth.GET("/google", h.GoogleLogin)
		auth.GET("/google/callback", h.GoogleCallback)

		auth.GET("/microsoft", h.MicrosoftLogin)
		auth.GET("/microsoft/callback", h.MicrosoftCallback)

		auth.GET("/github", h.GitHubLogin)
		auth.GET("/github/callback", h.GitHubCallback)

		auth.GET("/wechat", h.WeChatLogin)
		auth.GET("/wechat/callback", h.WeChatCallback)
	}
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Authenticate user
	user, token, err := h.authService.Authenticate(c.Request.Context(), req.UsernameOrEmail, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Return user and token
	c.JSON(http.StatusOK, gin.H{
		"user": user,
		"token": TokenResponse{
			Token:     token,
			ExpiresIn: int64(h.authConfig.JWTExpirationTime.Seconds()),
		},
	})
}

// RegisterUser handles user registration
func (h *AuthHandler) RegisterUser(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create user
	user := &model.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Active:    true,
	}

	// Register user
	if err := h.authService.Register(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return success
	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Get the refresh token from the request body
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Refresh the token
	newToken, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Return the new token
	c.JSON(http.StatusOK, TokenResponse{
		Token:     newToken,
		ExpiresIn: int64(h.authConfig.JWTExpirationTime.Seconds()),
	})
}

// GoogleLogin initiates Google OAuth2 login flow
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	// This is a placeholder for the actual implementation
	// In a real implementation, you would redirect to Google's OAuth URL
	googleAuthURL := "https://accounts.google.com/o/oauth2/auth" +
		"?client_id=" + h.authConfig.GoogleClientID +
		"&redirect_uri=" + h.authConfig.CallbackURLBase + "/api/v1/auth/google/callback" +
		"&scope=email%20profile" +
		"&response_type=code"

	c.Redirect(http.StatusTemporaryRedirect, googleAuthURL)
}

// GoogleCallback handles Google OAuth2 callback
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	// Get the authorization code from the callback
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code is missing"})
		return
	}

	// TODO: Exchange code for Google token and user data
	// For this draft implementation, we'll simulate this with mock data

	userData := map[string]interface{}{
		"email":      "google_user@example.com",
		"first_name": "Google",
		"last_name":  "User",
		"username":   "googleuser",
		"avatar_url": "https://example.com/avatar.png",
	}

	// Authenticate with provider
	user, token, err := h.authService.AuthenticateWithProvider(
		c.Request.Context(),
		model.AuthProviderGoogle,
		"google123", // This would be the actual Google user ID
		userData["email"].(string),
		userData,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// In a real implementation, you would redirect to your frontend with the token
	// For this draft, we'll just return the token
	c.JSON(http.StatusOK, gin.H{
		"user": user,
		"token": TokenResponse{
			Token:     token,
			ExpiresIn: int64(h.authConfig.JWTExpirationTime.Seconds()),
		},
	})
}

// MicrosoftLogin initiates Microsoft Entra ID OAuth2 login flow
func (h *AuthHandler) MicrosoftLogin(c *gin.Context) {
	// This is a placeholder for the actual implementation
	// In a real implementation, you would redirect to Microsoft's OAuth URL
	msAuthURL := "https://login.microsoftonline.com/" + h.authConfig.MicrosoftTenantID + "/oauth2/v2.0/authorize" +
		"?client_id=" + h.authConfig.MicrosoftClientID +
		"&redirect_uri=" + h.authConfig.CallbackURLBase + "/api/v1/auth/microsoft/callback" +
		"&scope=openid%20profile%20email" +
		"&response_type=code"

	c.Redirect(http.StatusTemporaryRedirect, msAuthURL)
}

// MicrosoftCallback handles Microsoft Entra ID OAuth2 callback
func (h *AuthHandler) MicrosoftCallback(c *gin.Context) {
	// Get the authorization code from the callback
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code is missing"})
		return
	}

	// TODO: Exchange code for Microsoft token and user data
	// For this draft implementation, we'll simulate this with mock data

	userData := map[string]interface{}{
		"email":      "ms_user@example.com",
		"first_name": "Microsoft",
		"last_name":  "User",
		"username":   "msuser",
		"avatar_url": "https://example.com/ms_avatar.png",
	}

	// Authenticate with provider
	user, token, err := h.authService.AuthenticateWithProvider(
		c.Request.Context(),
		model.AuthProviderMicrosoft,
		"ms456", // This would be the actual Microsoft user ID
		userData["email"].(string),
		userData,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// In a real implementation, you would redirect to your frontend with the token
	// For this draft, we'll just return the token
	c.JSON(http.StatusOK, gin.H{
		"user": user,
		"token": TokenResponse{
			Token:     token,
			ExpiresIn: int64(h.authConfig.JWTExpirationTime.Seconds()),
		},
	})
}

// GitHubLogin initiates GitHub OAuth2 login flow
func (h *AuthHandler) GitHubLogin(c *gin.Context) {
	// This is a placeholder for the actual implementation
	// In a real implementation, you would redirect to GitHub's OAuth URL
	githubAuthURL := "https://github.com/login/oauth/authorize" +
		"?client_id=" + h.authConfig.GitHubClientID +
		"&redirect_uri=" + h.authConfig.CallbackURLBase + "/api/v1/auth/github/callback" +
		"&scope=user:email"

	c.Redirect(http.StatusTemporaryRedirect, githubAuthURL)
}

// GitHubCallback handles GitHub OAuth2 callback
func (h *AuthHandler) GitHubCallback(c *gin.Context) {
	// Get the authorization code from the callback
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code is missing"})
		return
	}

	// TODO: Exchange code for GitHub token and user data
	// For this draft implementation, we'll simulate this with mock data

	userData := map[string]interface{}{
		"email":      "github_user@example.com",
		"first_name": "GitHub",
		"last_name":  "User",
		"username":   "githubuser",
		"avatar_url": "https://example.com/github_avatar.png",
	}

	// Authenticate with provider
	user, token, err := h.authService.AuthenticateWithProvider(
		c.Request.Context(),
		model.AuthProviderGitHub,
		"github789", // This would be the actual GitHub user ID
		userData["email"].(string),
		userData,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// In a real implementation, you would redirect to your frontend with the token
	// For this draft, we'll just return the token
	c.JSON(http.StatusOK, gin.H{
		"user": user,
		"token": TokenResponse{
			Token:     token,
			ExpiresIn: int64(h.authConfig.JWTExpirationTime.Seconds()),
		},
	})
}

// WeChatLogin initiates WeChat OAuth2 login flow
func (h *AuthHandler) WeChatLogin(c *gin.Context) {
	// This is a placeholder for the actual implementation
	// In a real implementation, you would redirect to WeChat's OAuth URL or show a QR code
	wechatAuthURL := "https://open.weixin.qq.com/connect/qrconnect" +
		"?appid=" + h.authConfig.WeChatAppID +
		"&redirect_uri=" + h.authConfig.CallbackURLBase + "/api/v1/auth/wechat/callback" +
		"&response_type=code" +
		"&scope=snsapi_login"

	c.Redirect(http.StatusTemporaryRedirect, wechatAuthURL)
}

// WeChatCallback handles WeChat OAuth2 callback
func (h *AuthHandler) WeChatCallback(c *gin.Context) {
	// Get the authorization code from the callback
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code is missing"})
		return
	}

	// TODO: Exchange code for WeChat token and user data
	// For this draft implementation, we'll simulate this with mock data

	userData := map[string]interface{}{
		"email":      "wechat_user@example.com", // WeChat might not provide email
		"first_name": "WeChat",
		"last_name":  "User",
		"username":   "wechatuser",
		"avatar_url": "https://example.com/wechat_avatar.png",
	}

	// Authenticate with provider
	user, token, err := h.authService.AuthenticateWithProvider(
		c.Request.Context(),
		model.AuthProviderWeChat,
		"wechat123", // This would be the actual WeChat user ID
		userData["email"].(string),
		userData,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// In a real implementation, you would redirect to your frontend with the token
	// For this draft, we'll just return the token
	c.JSON(http.StatusOK, gin.H{
		"user": user,
		"token": TokenResponse{
			Token:     token,
			ExpiresIn: int64(h.authConfig.JWTExpirationTime.Seconds()),
		},
	})
}
