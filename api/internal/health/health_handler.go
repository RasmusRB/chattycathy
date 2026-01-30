package health

import (
	"net/http"

	"github.com/chattycathy/api/pkg/redis"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// RegisterRoutes registers health check routes
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.GET("/health", h.Health)
	router.GET("/ready", h.Ready)
}

// Health is a simple liveness check
// Returns 200 if the service is alive
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

// Ready is a readiness check that verifies all dependencies
// Returns 200 if the service can serve traffic
func (h *Handler) Ready(c *gin.Context) {
	status := "ready"
	httpStatus := http.StatusOK

	checks := gin.H{}

	// Check database
	sqlDB, err := h.db.DB()
	if err != nil {
		checks["database"] = "error: " + err.Error()
		status = "not ready"
		httpStatus = http.StatusServiceUnavailable
	} else if err := sqlDB.Ping(); err != nil {
		checks["database"] = "error: " + err.Error()
		status = "not ready"
		httpStatus = http.StatusServiceUnavailable
	} else {
		checks["database"] = "ok"
	}

	// Check Redis
	if redis.IsConnected() {
		checks["redis"] = "ok"
	} else {
		checks["redis"] = "error: not connected"
		status = "not ready"
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, gin.H{
		"status": status,
		"checks": checks,
	})
}
