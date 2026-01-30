package middleware

import (
	"net/http"
	"strings"

	"github.com/chattycathy/api/pkg/auth"
	"github.com/gin-gonic/gin"
)

const (
	// AuthorizationHeader is the header key for the bearer token
	AuthorizationHeader = "Authorization"
	// BearerPrefix is the prefix for bearer tokens
	BearerPrefix = "Bearer "
	// ClaimsKey is the context key for storing claims
	ClaimsKey = "claims"
)

// JWTAuth is middleware that validates JWT tokens
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		// Store claims in context for handlers to use
		c.Set(ClaimsKey, claims)
		c.Next()
	}
}

// RequireRole is middleware that checks if the user has a specific role
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsInterface, exists := c.Get(ClaimsKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "not authenticated",
			})
			return
		}

		claims, ok := claimsInterface.(*auth.Claims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "invalid claims",
			})
			return
		}

		for _, role := range roles {
			if claims.Role == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "insufficient permissions",
		})
	}
}

// GetClaims retrieves the JWT claims from the context
func GetClaims(c *gin.Context) (*auth.Claims, bool) {
	claimsInterface, exists := c.Get(ClaimsKey)
	if !exists {
		return nil, false
	}
	claims, ok := claimsInterface.(*auth.Claims)
	return claims, ok
}

// RequirePermission is middleware that checks if the user has specific permissions
func RequirePermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsInterface, exists := c.Get(ClaimsKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "not authenticated",
			})
			return
		}

		claims, ok := claimsInterface.(*auth.Claims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "invalid claims",
			})
			return
		}

		// Check if user has any of the required permissions
		userPerms := make(map[string]bool)
		for _, p := range claims.Permissions {
			userPerms[p] = true
		}

		for _, required := range permissions {
			if userPerms[required] {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":    "insufficient permissions",
			"required": permissions,
		})
	}
}

// RequireAllPermissions is middleware that checks if the user has ALL specified permissions
func RequireAllPermissions(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsInterface, exists := c.Get(ClaimsKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "not authenticated",
			})
			return
		}

		claims, ok := claimsInterface.(*auth.Claims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "invalid claims",
			})
			return
		}

		// Check if user has all required permissions
		userPerms := make(map[string]bool)
		for _, p := range claims.Permissions {
			userPerms[p] = true
		}

		var missing []string
		for _, required := range permissions {
			if !userPerms[required] {
				missing = append(missing, required)
			}
		}

		if len(missing) > 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "insufficient permissions",
				"missing": missing,
			})
			return
		}

		c.Next()
	}
}
