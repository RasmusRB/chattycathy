package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/chattycathy/api/db/models"
	"github.com/chattycathy/api/pkg/auth"
	"github.com/chattycathy/api/pkg/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GoogleUserInfo represents the user info from Google's API
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// GoogleHandler handles Google OAuth authentication
type GoogleHandler struct {
	db                     *gorm.DB
	clientID               string
	clientSecret           string
	redirectURL            string
	issuer                 string
	accessTokenExpiryMins  int
	refreshTokenExpiryDays int
}

// NewGoogleHandler creates a new Google OAuth handler
func NewGoogleHandler(
	db *gorm.DB,
	clientID, clientSecret, redirectURL string,
	issuer string,
	accessExpiryMins, refreshExpiryDays int,
) *GoogleHandler {
	return &GoogleHandler{
		db:                     db,
		clientID:               clientID,
		clientSecret:           clientSecret,
		redirectURL:            redirectURL,
		issuer:                 issuer,
		accessTokenExpiryMins:  accessExpiryMins,
		refreshTokenExpiryDays: refreshExpiryDays,
	}
}

// RegisterRoutes registers Google OAuth routes
func (h *GoogleHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/auth/google", h.GoogleCallback)
	router.GET("/auth/google/config", h.GetGoogleConfig)
}

// GetGoogleConfig returns the Google OAuth client ID for the frontend
func (h *GoogleHandler) GetGoogleConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"client_id":    h.clientID,
		"redirect_uri": h.redirectURL,
	})
}

// GoogleCallbackRequest represents the request from frontend after Google sign-in
type GoogleCallbackRequest struct {
	// Either access_token (from implicit flow) or code (from authorization code flow)
	AccessToken string `json:"access_token"`
	Code        string `json:"code"`
}

// GoogleCallback handles the Google OAuth callback
func (h *GoogleHandler) GoogleCallback(c *gin.Context) {
	var req GoogleCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	var googleUser *GoogleUserInfo
	var err error

	if req.AccessToken != "" {
		// Implicit flow - verify access token directly
		googleUser, err = h.verifyGoogleAccessToken(req.AccessToken)
	} else if req.Code != "" {
		// Authorization code flow - exchange code for tokens first
		googleUser, err = h.exchangeCodeForUserInfo(req.Code)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "access_token or code required"})
		return
	}

	if err != nil {
		logger.Error().Err(err).Msg("Failed to verify Google token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid Google credentials"})
		return
	}

	// Find or create user
	user, err := h.findOrCreateUser(googleUser)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to find or create user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process user"})
		return
	}

	// Get user permissions
	permissions, err := models.GetUserPermissions(h.db, user.ID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get user permissions")
		permissions = []string{} // Continue with empty permissions
	}

	// Generate token pair
	userID := strconv.FormatUint(uint64(user.ID), 10)
	tokenPair, err := auth.GenerateTokenPair(
		userID, user.Email, user.Role,
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
		Username:    user.Email,
		Role:        user.Role,
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

	// Set refresh token as httpOnly cookie
	h.setRefreshTokenCookie(c, tokenPair.RefreshToken)

	// Ensure permissions is never null (JSON serialization)
	if permissions == nil {
		permissions = []string{}
	}

	// Return tokens and user info
	c.JSON(http.StatusOK, gin.H{
		"access_token":            tokenPair.AccessToken,
		"refresh_token":           tokenPair.RefreshToken,
		"access_token_expires_in": tokenPair.AccessTokenExpiresIn,
		"token_type":              tokenPair.TokenType,
		"user": gin.H{
			"id":          user.ID,
			"email":       user.Email,
			"name":        user.Name,
			"picture":     user.Picture,
			"role":        user.Role,
			"permissions": permissions,
		},
	})
}

// verifyGoogleAccessToken verifies a Google access token and returns user info
func (h *GoogleHandler) verifyGoogleAccessToken(accessToken string) (*GoogleUserInfo, error) {
	// Get user info from Google
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google API error: %s", string(body))
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}

// exchangeCodeForUserInfo exchanges an authorization code for user info
func (h *GoogleHandler) exchangeCodeForUserInfo(code string) (*GoogleUserInfo, error) {
	// Exchange code for access token
	resp, err := http.PostForm("https://oauth2.googleapis.com/token", map[string][]string{
		"code":          {code},
		"client_id":     {h.clientID},
		"client_secret": {h.clientSecret},
		"redirect_uri":  {h.redirectURL},
		"grant_type":    {"authorization_code"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange error: %s", string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	// Get user info with the access token
	return h.verifyGoogleAccessToken(tokenResp.AccessToken)
}

// findOrCreateUser finds an existing user or creates a new one
func (h *GoogleHandler) findOrCreateUser(googleUser *GoogleUserInfo) (*models.User, error) {
	var user models.User

	// Try to find by Google ID first
	result := h.db.Where("google_id = ?", googleUser.ID).First(&user)
	if result.Error == nil {
		// User exists, update last login and any changed info
		user.Email = googleUser.Email
		user.Name = googleUser.Name
		user.Picture = googleUser.Picture
		user.LastLoginAt = time.Now()
		if err := h.db.Save(&user).Error; err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
		logger.Info().
			Str("email", user.Email).
			Uint("user_id", user.ID).
			Msg("User logged in")
		return &user, nil
	}

	// Check if user exists with same email (legacy or different provider)
	result = h.db.Where("email = ?", googleUser.Email).First(&user)
	if result.Error == nil {
		// Link Google account to existing user
		user.GoogleID = googleUser.ID
		user.Name = googleUser.Name
		user.Picture = googleUser.Picture
		user.LastLoginAt = time.Now()
		if err := h.db.Save(&user).Error; err != nil {
			return nil, fmt.Errorf("failed to link Google account: %w", err)
		}
		logger.Info().
			Str("email", user.Email).
			Uint("user_id", user.ID).
			Msg("Linked Google account to existing user")
		return &user, nil
	}

	// Create new user
	user = models.User{
		GoogleID:    googleUser.ID,
		Email:       googleUser.Email,
		Name:        googleUser.Name,
		Picture:     googleUser.Picture,
		Role:        "user",
		LastLoginAt: time.Now(),
	}

	if err := h.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Assign default "user" role
	if err := models.AssignRoleToUser(h.db, user.ID, "user"); err != nil {
		logger.Warn().Err(err).Uint("user_id", user.ID).Msg("Failed to assign default role")
	}

	logger.Info().
		Str("email", user.Email).
		Uint("user_id", user.ID).
		Msg("New user created via Google OAuth")

	return &user, nil
}

func (h *GoogleHandler) setRefreshTokenCookie(c *gin.Context, token string) {
	maxAge := h.refreshTokenExpiryDays * 24 * 60 * 60
	c.SetSameSite(http.SameSiteLaxMode) // Lax to allow redirect from Google
	c.SetCookie(
		"refresh_token",
		token,
		maxAge,
		"/api/v1/auth",
		"",
		false, // Set true in production with HTTPS
		true,  // HttpOnly
	)
}
