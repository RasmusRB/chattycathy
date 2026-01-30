package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/chattycathy/api/config"
	"github.com/chattycathy/api/db"
	"github.com/chattycathy/api/docs"
	"github.com/chattycathy/api/internal/admin"
	internalauth "github.com/chattycathy/api/internal/auth"
	"github.com/chattycathy/api/internal/health"
	"github.com/chattycathy/api/internal/ping"
	"github.com/chattycathy/api/internal/protected"
	"github.com/chattycathy/api/pkg/auth"
	"github.com/chattycathy/api/pkg/logger"
	"github.com/chattycathy/api/pkg/middleware"
	"github.com/chattycathy/api/pkg/redis"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	// Initialize logger
	logger.Init(cfg.Log.Level, cfg.Log.Pretty)
	logger.Info().Msg("Starting ChattyCathy API")

	// Connect to database
	database, err := db.Connect(&cfg.Database)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	logger.Info().Msg("Database connected")

	// Connect to Redis
	if err := redis.Connect(&redis.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}); err != nil {
		logger.Warn().Err(err).Msg("Failed to connect to Redis - continuing without cache")
	}

	// Initialize JWT
	if err := auth.Init(&auth.Config{
		PrivateKeyPath:         cfg.JWT.PrivateKeyPath,
		PublicKeyPath:          cfg.JWT.PublicKeyPath,
		Issuer:                 cfg.JWT.Issuer,
		AccessTokenExpiryMins:  cfg.JWT.AccessTokenExpiryMins,
		RefreshTokenExpiryDays: cfg.JWT.RefreshTokenExpiryDays,
	}); err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize JWT")
	}
	logger.Info().
		Int("access_expiry_mins", cfg.JWT.AccessTokenExpiryMins).
		Int("refresh_expiry_days", cfg.JWT.RefreshTokenExpiryDays).
		Msg("JWT initialized with RSA-256")

	// Setup router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Add middlewares
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger())

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:8080", "http://chattycathy.localhost", "http://app.localhost", "http://api.localhost"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID", "X-Refresh-Token"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
	}))

	// Health check routes (no auth required)
	healthHandler := health.NewHandler(database)
	healthHandler.RegisterRoutes(router)

	// Register OpenAPI docs
	docs.RegisterRoutes(router)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Ping routes (public)
		pingService := ping.NewService(database)
		pingHandler := ping.NewHandler(pingService)
		pingHandler.RegisterRoutes(v1)

		// Auth routes (public)
		authHandler := internalauth.NewHandler(cfg.JWT.Issuer, cfg.JWT.AccessTokenExpiryMins, cfg.JWT.RefreshTokenExpiryDays)
		authHandler.RegisterRoutes(v1)

		// Google OAuth routes
		googleHandler := internalauth.NewGoogleHandler(
			database,
			cfg.Google.ClientID,
			cfg.Google.ClientSecret,
			cfg.Google.RedirectURL,
			cfg.JWT.Issuer,
			cfg.JWT.AccessTokenExpiryMins,
			cfg.JWT.RefreshTokenExpiryDays,
		)
		googleHandler.RegisterRoutes(v1)

		// Protected routes (require JWT)
		protectedHandler := protected.NewHandler()
		protectedHandler.RegisterRoutes(v1)

		// Admin routes (require admin role)
		adminHandler := admin.NewHandler(database)
		adminHandler.RegisterRoutes(v1)
	}

	// Create HTTP server
	addr := ":" + cfg.Server.Port
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info().Str("addr", addr).Msg("Server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	// Close Redis connection
	if err := redis.Close(); err != nil {
		logger.Warn().Err(err).Msg("Error closing Redis connection")
	}

	logger.Info().Msg("Server exited gracefully")
}
