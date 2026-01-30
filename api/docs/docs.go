package docs

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed openapi.yaml
var openAPISpec embed.FS

//go:embed swagger-ui/*
var swaggerUI embed.FS

// RegisterRoutes registers OpenAPI documentation routes
func RegisterRoutes(router *gin.Engine) {
	// Serve OpenAPI spec
	router.GET("/api/openapi.yaml", func(c *gin.Context) {
		data, err := openAPISpec.ReadFile("openapi.yaml")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load OpenAPI spec"})
			return
		}
		c.Header("Content-Type", "application/x-yaml")
		c.String(http.StatusOK, string(data))
	})

	// Serve Swagger UI
	swaggerFS, _ := fs.Sub(swaggerUI, "swagger-ui")
	router.StaticFS("/api/docs", http.FS(swaggerFS))
}
