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
	"github.com/onedotnet/platform/internal/model"
	"github.com/onedotnet/platform/internal/service"
	"github.com/onedotnet/platform/pkg/config"
	"github.com/onedotnet/platform/pkg/middleware"
	"golang.org/x/crypto/bcrypt"
)

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

	// Connect to database using the configuration
	db, err := config.Connect(&cfg.DB)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate database schemas
	if err := db.AutoMigrate(&model.User{}, &model.Organization{}, &model.Role{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migration completed successfully")

	// Create cache using the configuration
	cacheInstance, err := config.NewCache(&cfg.Cache)
	if err != nil {
		log.Fatalf("Failed to create cache: %v", err)
	}
	defer cacheInstance.Close()

	// Create repository
	repo := service.NewGormRepository(db, cacheInstance)

	// Create auth service
	authService := service.NewAuthService(repo, cfg.Auth)

	// Set up Gin framework
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

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
		log.Printf("API server listening on port %d", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Create a channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	sig := <-quit
	log.Printf("Received signal %v, shutting down server gracefully...", sig)

	// Create a deadline for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 15*time.Second)
	defer shutdownCancel()

	// Shut down the server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server gracefully stopped")
}

// monitorSystemHealth periodically checks system health
// This function makes use of the ctx variable to properly handle cancellation
func monitorSystemHealth(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Health monitoring stopped")
			return
		case <-ticker.C:
			// Perform health checks here
			log.Println("Performing system health check")
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
			log.Printf("Error creating admin role: %v", err)
		} else {
			log.Println("Created admin role successfully")
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
			log.Printf("Error creating user role: %v", err)
		} else {
			log.Println("Created user role successfully")
		}
	}

	// Check if admin user exists
	_, err = repo.GetUserByUsername(ctx, "admin")
	if err != nil {
		// Create admin user
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hashing admin password: %v", err)
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
			log.Printf("Error creating admin user: %v", err)
			return
		}

		// Make sure admin role exists
		if adminRole != nil {
			adminUser.Roles = []model.Role{*adminRole}
			if err := repo.UpdateUser(ctx, adminUser); err != nil {
				log.Printf("Error assigning admin role to admin user: %v", err)
			}
		}

		log.Println("Created admin user successfully")
	}
}
