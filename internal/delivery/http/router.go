package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go-gin-todo/docs"
	"go-gin-todo/internal/config"
	"go-gin-todo/internal/delivery/http/handler"
	"go-gin-todo/internal/delivery/http/middleware"
	"go-gin-todo/internal/security"
)

func SetupRouter(
	cfg *config.Config,
	tokenManager security.TokenManager,
	authHandler *handler.AuthHandler,
	todolistHandler *handler.TodolistHandler,
	todoHandler *handler.TodoHandler,
) *gin.Engine {
	if cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Global middleware
	router.Use(middleware.CORSMiddleware())

	// Root welcome / info
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":        "Clean Architecture Go Gin Todo Server",
			"prefix":         "/api",
			"defaultVersion": "v1",
			"swagger":        "/swagger",
			"openapi":        "/openapi.json",
		})
	})

	// Serve openapi.json
	router.GET("/openapi.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json; charset=utf-8", docs.OpenAPISpec)
	})

	// Serve Swagger UI
	swaggerHTML := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Go Gin Todo API - Swagger UI</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: '/openapi.json',
        dom_id: '#swagger-ui',
      });
    };
  </script>
</body>
</html>`

	serveSwagger := func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, swaggerHTML)
	}

	router.GET("/docs", serveSwagger)

	// Base API group: /api
	apiGroup := router.Group("/api")
	apiGroup.Use(func(c *gin.Context) {
		c.Header("X-API-Version", "v1")
		c.Next()
	})

	// Version 1 group: /api/v1
	v1Group := apiGroup.Group("/v1")

	// Register routes on /api/v1
	registerRoutes(v1Group, tokenManager, authHandler, todolistHandler, todoHandler)

	return router
}

func registerRoutes(
	rg *gin.RouterGroup,
	tokenManager security.TokenManager,
	authHandler *handler.AuthHandler,
	todolistHandler *handler.TodolistHandler,
	todoHandler *handler.TodoHandler,
) {
	rg.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP", "version": "v1"})
	})

	// Public Auth endpoints
	// 5. POST: /register, /login (use jwt for user auth)
	rg.POST("/register", authHandler.Register)
	rg.POST("/login", authHandler.Login)

	// Protected routes (JWT middleware)
	protected := rg.Group("")
	protected.Use(middleware.AuthMiddleware(tokenManager))
	{
		// 1. GET (bulk), POST (one, bulk), PATCH (update by id, one, bulk), DELETE (one, bulk): /todolist
		// 2. GET (one): /todolist/:id
		todolistGroup := protected.Group("/todolist")
		{
			todolistGroup.GET("", todolistHandler.GetBulk)
			todolistGroup.POST("", todolistHandler.Create)
			todolistGroup.PATCH("", todolistHandler.PatchOneOrBulk)
			todolistGroup.DELETE("", todolistHandler.DeleteOneOrBulk)

			todolistGroup.GET("/:id", todolistHandler.GetOne)
			todolistGroup.PATCH("/:id", todolistHandler.PatchByID)
			todolistGroup.DELETE("/:id", todolistHandler.DeleteByID)
		}

		// 3. GET (bulk), POST (one, bulk), PATCH (update by id, one, bulk), DELETE (one, bulk): /todo
		// 4. GET (one): /todo/:id
		todoGroup := protected.Group("/todo")
		{
			todoGroup.GET("", todoHandler.GetBulk)
			todoGroup.POST("", todoHandler.Create)
			todoGroup.PATCH("", todoHandler.PatchOneOrBulk)
			todoGroup.DELETE("", todoHandler.DeleteOneOrBulk)

			todoGroup.GET("/:id", todoHandler.GetOne)
			todoGroup.PATCH("/:id", todoHandler.PatchByID)
			todoGroup.DELETE("/:id", todoHandler.DeleteByID)
		}
	}
}
