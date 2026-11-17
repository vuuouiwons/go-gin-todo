package app

import (
	"github.com/gin-gonic/gin"
	"go-gin-todo/internal/config"
	httpdelivery "go-gin-todo/internal/delivery/http"
	"go-gin-todo/internal/delivery/http/handler"
	"go-gin-todo/internal/repository"
	"go-gin-todo/internal/security"
	"go-gin-todo/internal/usecase"
	"gorm.io/gorm"
)

// BuildRouter wires all Clean Architecture layers using Dependency Injection
func BuildRouter(cfg *config.Config, db *gorm.DB) *gin.Engine {
	// 1. Repositories (Infrastructure / Adapters)
	userRepo := repository.NewUserRepository(db)
	todolistRepo := repository.NewTodolistRepository(db)
	todoRepo := repository.NewTodoRepository(db)

	// 2. Security (Infrastructure / Adapters)
	hasher := security.NewBcryptHasher(0)
	tokenManager := security.NewJWTTokenManager(cfg.JWTSecret, cfg.JWTExpiresInHours)

	// 3. Usecases (Application Layer)
	authUsecase := usecase.NewAuthUsecase(userRepo, hasher, tokenManager)
	todolistUsecase := usecase.NewTodolistUsecase(todolistRepo)
	todoUsecase := usecase.NewTodoUsecase(todoRepo)

	// 4. HTTP Handlers (Delivery Layer)
	authHandler := handler.NewAuthHandler(authUsecase)
	todolistHandler := handler.NewTodolistHandler(todolistUsecase)
	todoHandler := handler.NewTodoHandler(todoUsecase)

	// 5. Router & Middleware (Delivery Layer)
	return httpdelivery.SetupRouter(cfg, tokenManager, authHandler, todolistHandler, todoHandler)
}
