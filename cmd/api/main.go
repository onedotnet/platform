package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	v1 "github.com/onedotnet/platform/api/v1"
	_ "github.com/onedotnet/platform/docs" // Import for Swagger
	"github.com/onedotnet/platform/internal/model"
	"github.com/onedotnet/platform/internal/service"
	"github.com/onedotnet/platform/pkg/config"
	"github.com/onedotnet/platform/pkg/logger"
	"github.com/onedotnet/platform/pkg/middleware"
	"github.com/onedotnet/platform/pkg/middleware/mq" // Import the new mq package
	"github.com/onedotnet/platform/pkg/swagger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// @title          Platform API
// @version        1.0
// @description    Go-based backend platform service featuring user, organization, and role management
// @termsOfService http://swagger.io/terms/

// @contact.name  OneDotNet Team
// @contact.url   https://onedotnet.org
// @contact.email support@onedotnet.org

// @license.name Apache 2.0
// @license.url  http://www.apache.org/licenses/LICENSE-2.0.html

// @host     localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter the token with the `Bearer: ` prefix, e.g. "Bearer abcde12345".

func main() {
	log.Println("Starting API server...")

	// Set up context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize configuration
	cfg, err := config.Setup()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger.Setup(logger.Config{
		Level:       cfg.Logger.Level,
		Development: cfg.Logger.Development,
		OutputPaths: cfg.Logger.OutputPaths,
	})
	defer logger.Sync()

	logger.Log.Info("Logger initialized successfully")

	// Connect to database using the configuration
	db, err := config.Connect(&cfg.DB)
	if err != nil {
		logger.Log.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Auto migrate database schemas
	if err := db.AutoMigrate(&model.User{}, &model.Organization{}, &model.Role{}, &model.Task{}); err != nil {
		logger.Log.Fatal("Failed to migrate database", zap.Error(err))
	}
	logger.Log.Info("Database migration completed successfully")

	// Create cache using the configuration
	cacheInstance, err := config.NewCache(&cfg.Cache)
	if err != nil {
		logger.Log.Fatal("Failed to create cache", zap.Error(err))
	}
	defer cacheInstance.Close()

	// Create repository
	repo := service.NewGormRepository(db, cacheInstance)

	// Initialize RabbitMQ service from mq package
	mqService, err := mq.GetGlobalMQService(cfg.RabbitMQ, repo)
	if err != nil {
		logger.Log.Error("Failed to initialize RabbitMQ service", zap.Error(err))
		// Continue without RabbitMQ as it's not critical for API functionality
	} else {
		defer mqService.Close()
		logger.Log.Info("RabbitMQ service initialized with queue", zap.String("queue", mqService.GetQueueName()))
	}

	// Create auth service
	authService := service.NewAuthService(repo, cfg.Auth)

	// Set up Gin framework
	if cfg.Logger.Development {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add middleware
	router.Use(middleware.Logger())
	router.Use(middleware.ErrorHandler())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// Set up Swagger
	swagger.Setup(router)

	// API version group
	apiV1 := router.Group("/api/v1")

	// Register API handlers
	userHandler := v1.NewUserHandler(repo)
	userHandler.Register(apiV1)

	orgHandler := v1.NewOrganizationHandler(repo)
	orgHandler.Register(apiV1)

	roleHandler := v1.NewRoleHandler(repo)
	roleHandler.Register(apiV1)

	// Register auth handler
	authHandler := v1.NewAuthHandler(authService, cfg.Auth)
	authHandler.Register(apiV1)

	// Add health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().Format(time.RFC3339)})
	})

	// Protected routes that require authentication
	protectedRoutes := apiV1.Group("/protected")
	protectedRoutes.Use(middleware.AuthMiddleware(authService))
	{
		protectedRoutes.GET("/profile", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			user, err := repo.GetUser(c.Request.Context(), userID.(uint))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"user": user})
		})
	}

	// Admin-only routes that require authentication and admin role
	adminRoutes := apiV1.Group("/admin")
	adminRoutes.Use(middleware.AuthMiddleware(authService), middleware.RoleAuthMiddleware(authService, "admin"))
	{
		adminRoutes.GET("/dashboard", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Admin dashboard"})
		})
	}

	// Use ctx for any context-based operations
	go monitorSystemHealth(ctx)

	// Create default admin user if it doesn't exist
	go createDefaultAdminUser(ctx, repo)

	// Server configuration
	port := cfg.API.Port
	timeout := cfg.API.Timeout

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		IdleTimeout:  timeout * 2,
	}

	// Start the server in a goroutine
	go func() {
		logger.Log.Info("API server starting", zap.Int("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Create a channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	sig := <-quit
	logger.Log.Info("Received signal, shutting down server gracefully...", zap.String("signal", sig.String()))

	// Create a deadline for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 15*time.Second)
	defer shutdownCancel()

	// Shut down the server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Log.Fatal("Server shutdown failed", zap.Error(err))
	}

	logger.Log.Info("Server gracefully stopped")
}

// monitorSystemHealth periodically checks system health
// This function makes use of the ctx variable to properly handle cancellation
func monitorSystemHealth(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Health monitoring stopped")
			return
		case <-ticker.C:
			// Perform health checks here
			logger.Log.Debug("Performing system health check")
		}
	}
}

// createDefaultAdminUser creates a default admin user if it doesn't exist
func createDefaultAdminUser(ctx context.Context, repo service.Repository) {
	// Use the context for operations that should be cancelable
	// Example: Create default admin role if it doesn't exist
	adminRole, err := repo.GetRoleByName(ctx, "admin")
	if err != nil {
		// Create admin role
		adminRole = &model.Role{
			Name:        "admin",
			Description: "Administrator role with full access",
		}
		if err := repo.CreateRole(ctx, adminRole); err != nil {
			logger.Log.Error("Error creating admin role", zap.Error(err))
		} else {
			logger.Log.Info("Created admin role successfully")
		}
	}

	// Example: Create default user role if it doesn't exist
	userRole, err := repo.GetRoleByName(ctx, "user")
	if err != nil {
		// Create user role
		userRole = &model.Role{
			Name:        "user",
			Description: "Regular user with limited access",
		}
		if err := repo.CreateRole(ctx, userRole); err != nil {
			logger.Log.Error("Error creating user role", zap.Error(err))
		} else {
			logger.Log.Info("Created user role successfully")
		}
	}

	// Check if admin user exists
	_, err = repo.GetUserByUsername(ctx, "admin")
	if err != nil {
		// Create admin user
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			logger.Log.Error("Error hashing admin password", zap.Error(err))
			return
		}

		adminUser := &model.User{
			Username:  "admin",
			Email:     "admin@example.com",
			Password:  string(hashedPassword),
			FirstName: "Admin",
			LastName:  "User",
			Active:    true,
			Provider:  model.AuthProviderLocal,
		}

		// Create user
		if err := repo.CreateUser(ctx, adminUser); err != nil {
			logger.Log.Error("Error creating admin user", zap.Error(err))
			return
		}

		// Make sure admin role exists
		if adminRole != nil {
			adminUser.Roles = []model.Role{*adminRole}
			if err := repo.UpdateUser(ctx, adminUser); err != nil {
				logger.Log.Error("Error assigning admin role to admin user", zap.Error(err))
			}
		}

		logger.Log.Info("Created admin user successfully")
	}
}
