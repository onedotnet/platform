package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/onedotnet/platform/internal/model"
	"github.com/onedotnet/platform/internal/service"
	"github.com/onedotnet/platform/pkg/config"
)

func main() {
	log.Println("Starting platform service...")

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

	// Create a channel to listen for interrupt signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	sig := <-sigCh
	log.Printf("Received signal %v, shutting down gracefully...", sig)

	// Cancel context to signal all operations to finish
	cancel()

	// Allow some time for operations to complete
	time.Sleep(2 * time.Second)

	log.Println("Service stopped")
}
