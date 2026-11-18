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

	"go-gin-todo/internal/api"
	"go-gin-todo/internal/app"
	"go-gin-todo/internal/config"
	"go-gin-todo/internal/infra"
)

func main() {
	cfg := config.LoadConfig()

	db, err := infra.NewSQLiteDatabase(cfg.DBPath)
	if err != nil {
		log.Fatalf("Fatal: Database initialization failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	// 1. Infrastructure layer (ACL adapters)
	dbAdapter := infra.NewDatabase(db)
	userRepo := infra.NewUserRepository(db)
	todolistRepo := infra.NewTodolistRepository(db)
	todoRepo := infra.NewTodoRepository(db)
	hasher := infra.NewBcryptHasher(0)
	tokenManager := infra.NewJWTTokenManager(cfg.JWTSecret, cfg.JWTExpiresInHours)

	// 2. Application layer (Commands & Queries)
	authCommands := app.NewAuthCommands(userRepo, hasher, tokenManager)
	todolistCommands := app.NewTodolistCommands(todolistRepo)
	todolistQueries := app.NewTodolistQueries(dbAdapter, todolistRepo)
	todoCommands := app.NewTodoCommands(todoRepo)
	todoQueries := app.NewTodoQueries(dbAdapter, todoRepo)

	// 3. API layer (Handlers & Router)
	authHandler := api.NewAuthHandler(authCommands)
	todolistHandler := api.NewTodolistHandler(todolistCommands, todolistQueries)
	todoHandler := api.NewTodoHandler(todoCommands, todoQueries)

	router := api.SetupRouter(cfg.GinMode, tokenManager, dbAdapter, authHandler, todolistHandler, todoHandler)

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
