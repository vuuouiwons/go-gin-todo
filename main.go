package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-gin-todo/src/business"
	"go-gin-todo/src/data"
	"go-gin-todo/src/presentation"
	"go-gin-todo/src/shared"
)

func main() {
	cfg := shared.LoadConfig()

	database, err := data.Connect(cfg.DBPath)
	if err != nil {
		log.Fatalf("Fatal: Database initialization failed: %v", err)
	}

	sqlDB, err := database.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	// 1. Data Access layer (Repositories)
	userRepo := data.NewUserRepository(database)
	todolistRepo := data.NewTodolistRepository(database)
	todoRepo := data.NewTodoRepository(database)

	// 2. Business Logic layer (Services)
	authService := business.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiresInHours)
	todolistService := business.NewTodolistService(todolistRepo)
	todoService := business.NewTodoService(todoRepo)

	// 3. Presentation layer (Controllers & Router)
	authController := presentation.NewAuthController(authService)
	todolistController := presentation.NewTodolistController(todolistService)
	todoController := presentation.NewTodoController(todoService)

	router := presentation.SetupRouter(cfg.GinMode, cfg.JWTSecret, authController, todolistController, todoController)

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
