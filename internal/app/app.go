package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go-gin-todo/internal/config"
	"go-gin-todo/internal/database"
	httpdelivery "go-gin-todo/internal/delivery/http"
	"go-gin-todo/internal/delivery/http/handler"
	"go-gin-todo/internal/repository"
	"go-gin-todo/internal/security"
	"go-gin-todo/internal/usecase"
	"gorm.io/gorm"
)

// BuildRouter wires all Clean Architecture layers using Dependency Injection
func BuildRouter(cfg *config.Config, db *gorm.DB) *gin.Engine {
	// 1. Repositories (Persistence)
	userRepo := repository.NewUserRepository(db)
	todolistRepo := repository.NewTodolistRepository(db)
	todoRepo := repository.NewTodoRepository(db)

	// 2. Security
	hasher := security.NewBcryptHasher(0)
	tokenManager := security.NewJWTTokenManager(cfg.JWTSecret, cfg.JWTExpiresInHours)

	// 3. Usecases (Business Logic)
	authUsecase := usecase.NewAuthUsecase(userRepo, hasher, tokenManager)
	todolistUsecase := usecase.NewTodolistUsecase(todolistRepo)
	todoUsecase := usecase.NewTodoUsecase(todoRepo)

	// 4. HTTP Handlers (Delivery)
	authHandler := handler.NewAuthHandler(authUsecase)
	todolistHandler := handler.NewTodolistHandler(todolistUsecase)
	todoHandler := handler.NewTodoHandler(todoUsecase)

	// 5. Router & Middleware
	return httpdelivery.SetupRouter(cfg, tokenManager, authHandler, todolistHandler, todoHandler)
}

// Run bootstraps and manages the server lifecycle with graceful shutdown
func Run() {
	cfg := config.LoadConfig()

	db, err := database.NewSQLiteDatabase(cfg.DBPath)
	if err != nil {
		log.Fatalf("Fatal: Database initialization failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	router := BuildRouter(cfg, db)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server listening on port :%s", cfg.Port)
		log.Printf("API Versioned Prefix: http://localhost:%s/api/v1", cfg.Port)
		log.Printf("Documentation:        http://localhost:%s/docs", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server listen error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
