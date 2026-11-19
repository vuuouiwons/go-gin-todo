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

	"go-gin-todo/src/db"
	"go-gin-todo/src/db/repository"
	"go-gin-todo/src/routes"
	"go-gin-todo/src/services"
	"go-gin-todo/src/utils"
)

func main() {
	cfg := utils.LoadConfig()

	database, err := db.Connect(cfg.DBPath)
	if err != nil {
		log.Fatalf("Fatal: Database initialization failed: %v", err)
	}

	sqlDB, err := database.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	// 1. Repositories in src/db/repository (reusable database queries)
	userRepo := repository.NewUserRepository(database)
	todolistRepo := repository.NewTodolistRepository(database)
	todoRepo := repository.NewTodoRepository(database)

	// 2. Services layer (Business logic & database transactions)
	authService := services.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiresInHours)
	todolistService := services.NewTodolistService(database, todolistRepo)
	todoService := services.NewTodoService(database, todoRepo)

	// 3. Routes layer (Handlers & Router)
	authHandler := routes.NewAuthHandler(authService)
	todolistHandler := routes.NewTodolistHandler(todolistService)
	todoHandler := routes.NewTodoHandler(todoService)

	router := routes.SetupRouter(cfg.GinMode, cfg.JWTSecret, authHandler, todolistHandler, todoHandler)

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
