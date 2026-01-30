package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/chattycathy/api/pkg/auth"
	"github.com/chattycathy/api/pkg/logger"
	"github.com/chattycathy/api/pkg/middleware"
	"github.com/gin-gonic/gin"
)

const (
	refreshTokenCookie = "refresh_token"
)

// Handler handles authentication endpoints
type Handler struct {
	issuer                 string
	accessTokenExpiryMins  int
	refreshTokenExpiryDays int
}

// NewHandler creates a new auth handler
func NewHandler(issuer string, accessExpiryMins, refreshExpiryDays int) *Handler {
	return &Handler{
		issuer:                 issuer,
		accessTokenExpiryMins:  accessExpiryMins,
		refreshTokenExpiryDays: refreshExpiryDays,
	}
}

// RegisterRoutes registers auth routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/auth/login", h.Login)
	router.POST("/auth/refresh", h.Refresh)
	router.POST("/auth/logout", h.Logout)
	router.POST("/auth/logout-all", middleware.JWTAuth(), h.LogoutAll)
	router.GET("/auth/sessions", middleware.JWTAuth(), h.ListSessions)
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login authenticates a user and returns access + refresh tokens
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Demo authentication - in production, verify against DB
	if req.Password != "password123" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// In production, get user ID and role from database
	userID := "user-" + req.Username
	role := "user"
	if req.Username == "admin" {
		role = "admin"
	}

	// Demo permissions - in production, get from database
	permissions := []string{"ping:read", "news:read"}
	if role == "admin" {
		permissions = []string{"ping:read", "news:read", "news:create", "news:update", "news:delete", "users:read", "users:update"}
	}

	// Generate token pair
	tokenPair, err := auth.GenerateTokenPair(
		userID, req.Username, role,
		permissions,
		h.issuer,
		h.accessTokenExpiryMins,
		h.refreshTokenExpiryDays,
	)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to generate token pair")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	// Store refresh token in Redis
	ctx := context.Background()
	refreshData := &auth.RefreshTokenData{
		UserID:      userID,
		Username:    req.Username,
		Role:        role,
		Permissions: permissions,
		CreatedAt:   time.Now(),
		UserAgent:   c.Request.UserAgent(),
		IP:          c.ClientIP(),
	}
	expiry := time.Duration(h.refreshTokenExpiryDays) * 24 * time.Hour
	if err := auth.StoreRefreshToken(ctx, tokenPair.RefreshToken, refreshData, expiry); err != nil {
		logger.Error().Err(err).Msg("Failed to store refresh token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	// Set refresh token as httpOnly cookie (for web clients)
	h.setRefreshTokenCookie(c, tokenPair.RefreshToken)

	// Return tokens in response body (for mobile/API clients)
	c.JSON(http.StatusOK, tokenPair)
}

// RefreshRequest for mobile clients that send refresh token in body
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Refresh exchanges a valid refresh token for a new access token
func (h *Handler) Refresh(c *gin.Context) {
	// Try to get refresh token from cookie first (web), then from body (mobile)
	refreshToken := h.getRefreshToken(c)
	if refreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token required"})
		return
	}

	// Validate refresh token
	ctx := context.Background()
	tokenData, err := auth.GetRefreshTokenData(ctx, refreshToken)
	if err != nil {
		logger.Warn().Err(err).Msg("Invalid refresh token")
		h.clearRefreshTokenCookie(c)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	// Rotate refresh token (issue new one, revoke old)
	if err := auth.RevokeRefreshToken(ctx, refreshToken); err != nil {
		logger.Warn().Err(err).Msg("Failed to revoke old refresh token")
	}

	// Generate new token pair
	tokenPair, err := auth.GenerateTokenPair(
		tokenData.UserID, tokenData.Username, tokenData.Role,
		tokenData.Permissions,
		h.issuer,
		h.accessTokenExpiryMins,
		h.refreshTokenExpiryDays,
	)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to generate new token pair")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to refresh tokens"})
		return
	}

	// Store new refresh token
	newRefreshData := &auth.RefreshTokenData{
		UserID:      tokenData.UserID,
		Username:    tokenData.Username,
		Role:        tokenData.Role,
		Permissions: tokenData.Permissions,
		CreatedAt:   time.Now(),
		UserAgent:   c.Request.UserAgent(),
		IP:          c.ClientIP(),
	}
	expiry := time.Duration(h.refreshTokenExpiryDays) * 24 * time.Hour
	if err := auth.StoreRefreshToken(ctx, tokenPair.RefreshToken, newRefreshData, expiry); err != nil {
		logger.Error().Err(err).Msg("Failed to store new refresh token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	// Update cookie
	h.setRefreshTokenCookie(c, tokenPair.RefreshToken)

	c.JSON(http.StatusOK, tokenPair)
}

// Logout revokes the current refresh token
func (h *Handler) Logout(c *gin.Context) {
	refreshToken := h.getRefreshToken(c)
	if refreshToken != "" {
		ctx := context.Background()
		if err := auth.RevokeRefreshToken(ctx, refreshToken); err != nil {
			logger.Warn().Err(err).Msg("Failed to revoke refresh token")
		}
	}

	h.clearRefreshTokenCookie(c)
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// LogoutAll revokes all refresh tokens for the current user
func (h *Handler) LogoutAll(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	ctx := context.Background()
	if err := auth.RevokeAllUserTokens(ctx, claims.UserID); err != nil {
		logger.Error().Err(err).Msg("Failed to revoke all tokens")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout all sessions"})
		return
	}

	h.clearRefreshTokenCookie(c)
	c.JSON(http.StatusOK, gin.H{"message": "all sessions logged out"})
}

// ListSessions returns all active sessions for the current user
func (h *Handler) ListSessions(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	ctx := context.Background()
	sessions, err := auth.ListUserTokens(ctx, claims.UserID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to list sessions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list sessions"})
		return
	}

	// Format sessions for response (don't expose tokens)
	var result []gin.H
	for _, s := range sessions {
		result = append(result, gin.H{
			"created_at": s.CreatedAt,
			"user_agent": s.UserAgent,
			"ip":         s.IP,
		})
	}

	c.JSON(http.StatusOK, gin.H{"sessions": result})
}

// Helper methods for refresh token handling

func (h *Handler) getRefreshToken(c *gin.Context) string {
	// Try cookie first (web clients)
	if token, err := c.Cookie(refreshTokenCookie); err == nil && token != "" {
		return token
	}

	// Try request body (mobile clients)
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err == nil && req.RefreshToken != "" {
		return req.RefreshToken
	}

	// Try Authorization header with "Refresh" prefix
	if header := c.GetHeader("X-Refresh-Token"); header != "" {
		return header
	}

	return ""
}

func (h *Handler) setRefreshTokenCookie(c *gin.Context, token string) {
	maxAge := h.refreshTokenExpiryDays * 24 * 60 * 60
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		refreshTokenCookie,
		token,
		maxAge,
		"/api/v1/auth", // Only send to auth endpoints
		"",             // Domain (empty = current domain)
		false,          // Secure (set true in production with HTTPS)
		true,           // HttpOnly
	)
}

func (h *Handler) clearRefreshTokenCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		refreshTokenCookie,
		"",
		-1,
		"/api/v1/auth",
		"",
		false,
		true,
	)
}
