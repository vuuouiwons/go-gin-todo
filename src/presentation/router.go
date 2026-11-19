package presentation

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go-gin-todo/docs"
)

// SetupRouter initializes the Gin engine, middlewares, documentation, and versioned API routes.
func SetupRouter(
	ginMode string,
	jwtSecret string,
	authController *AuthController,
	todolistController *TodolistController,
	todoController *TodoController,
) *gin.Engine {
	if ginMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Use(CORSMiddleware())

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":        "3-Layer Architecture Go Gin Todo Server",
			"prefix":         "/api",
			"defaultVersion": "v1",
			"docs":           "/docs",
			"openapi":        "/openapi.json",
		})
	})

	router.GET("/openapi.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json; charset=utf-8", docs.OpenAPISpec)
	})

	const swaggerHTML = `<!DOCTYPE html>
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

	router.GET("/docs", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, swaggerHTML)
	})

	apiGroup := router.Group("/api")
	apiGroup.Use(func(c *gin.Context) {
		c.Header("X-API-Version", "v1")
		c.Next()
	})

	v1Group := apiGroup.Group("/v1")
	registerRoutes(v1Group, jwtSecret, authController, todolistController, todoController)

	// Transparently route unversioned /api/* requests to default version /api/v1/*
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") && !strings.HasPrefix(path, "/api/v") {
			c.Request.URL.Path = "/api/v1" + strings.TrimPrefix(path, "/api")
			router.HandleContext(c)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
	})

	return router
}

func registerRoutes(
	rg *gin.RouterGroup,
	jwtSecret string,
	authController *AuthController,
	todolistController *TodolistController,
	todoController *TodoController,
) {
	rg.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP", "version": "v1"})
	})

	rg.POST("/register", authController.Register)
	rg.POST("/login", authController.Login)

	protected := rg.Group("")
	protected.Use(AuthMiddleware(jwtSecret))
	{
		todolistGroup := protected.Group("/todolist")
		{
			todolistGroup.GET("", todolistController.GetBulk)
			todolistGroup.POST("", todolistController.Create)
			todolistGroup.PATCH("", todolistController.PatchOneOrBulk)
			todolistGroup.DELETE("", todolistController.Delete)

			todolistGroup.GET("/:id", todolistController.GetOne)
			todolistGroup.PATCH("/:id", todolistController.PatchByID)
			todolistGroup.DELETE("/:id", todolistController.DeleteByID)
		}

		todoGroup := protected.Group("/todo")
		{
			todoGroup.GET("", todoController.GetBulk)
			todoGroup.POST("", todoController.Create)
			todoGroup.PATCH("", todoController.PatchOneOrBulk)
			todoGroup.DELETE("", todoController.Delete)

			todoGroup.GET("/:id", todoController.GetOne)
			todoGroup.PATCH("/:id", todoController.PatchByID)
			todoGroup.DELETE("/:id", todoController.DeleteByID)
		}
	}
}
