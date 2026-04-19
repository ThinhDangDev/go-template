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

	"github.com/user/test-phase08/internal/delivery/rest"
	"github.com/user/test-phase08/internal/infrastructure/auth"
	"github.com/user/test-phase08/internal/infrastructure/database"
	"github.com/user/test-phase08/internal/infrastructure/logger"
	"github.com/user/test-phase08/internal/infrastructure/repository"
	"github.com/user/test-phase08/internal/usecase"
)

func main() {
	// Initialize logger
	zapLogger := logger.NewZapLogger()
	defer zapLogger.Sync()

	// Connect to database
	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	// Initialize JWT service
	jwtService := auth.NewJWTService()

	// Initialize usecases
	userUsecase := usecase.NewUserUsecase(userRepo)
	authUsecase := usecase.NewAuthUsecase(
		userRepo,
		jwtService,
		userUsecase,
	)

	// Initialize router
	router := rest.NewRouter(
		userUsecase,
		authUsecase,
		jwtService,
		zapLogger,
	)

	// Start server
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		zapLogger.Info(fmt.Sprintf("Server starting on %s", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal(fmt.Sprintf("Server failed: %v", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("Server shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zapLogger.Fatal(fmt.Sprintf("Server forced to shutdown: %v", err))
	}

	zapLogger.Info("Server exited")
}
