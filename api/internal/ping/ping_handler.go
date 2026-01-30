package ping

import (
	"net/http"

	"github.com/chattycathy/api/pkg/middleware"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers all ping routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	// Ping requires authentication and ping:read permission
	router.GET("/ping", middleware.JWTAuth(), middleware.RequirePermission("ping:read"), h.Ping)
}

// Ping godoc
// @Summary      Ping the server
// @Description  Returns pong and records the ping in the database
// @Tags         health
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  PingResponse
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /ping [get]
func (h *Handler) Ping(c *gin.Context) {
	response, err := h.service.Ping()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
