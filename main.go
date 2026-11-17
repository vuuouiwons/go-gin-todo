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

	"go-gin-todo/internal/app"
	"go-gin-todo/internal/config"
	"go-gin-todo/internal/database"
)

func main() {
	cfg := config.LoadConfig()

	// Initialize Database with auto-migrations
	db, err := database.NewSQLiteDatabase(cfg.DBPath)
	if err != nil {
		log.Fatalf("Fatal: Database initialization failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	// Build Clean Architecture application router
	router := app.BuildRouter(cfg, db)

	// Configure HTTP Server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in background
	go func() {
		log.Printf("Server listening on port :%s", cfg.Port)
		log.Printf("API Versioned Prefix: http://localhost:%s/api/v1", cfg.Port)
		log.Printf("API Default Prefix:   http://localhost:%s/api", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server listen error: %v", err)
		}
	}()

	// Graceful shutdown
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
