package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chattycathy/api/pkg/redis"
)

const (
	refreshTokenPrefix = "refresh_token:"
	userTokensPrefix   = "user_tokens:"
)

// RefreshTokenData stores metadata about a refresh token
type RefreshTokenData struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UserAgent   string    `json:"user_agent"`
	IP          string    `json:"ip"`
}

// StoreRefreshToken stores a refresh token in Redis with associated user data
func StoreRefreshToken(ctx context.Context, token string, data *RefreshTokenData, expiry time.Duration) error {
	if redis.Client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	// Store token data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal token data: %w", err)
	}

	key := refreshTokenPrefix + token
	if err := redis.Client.Set(ctx, key, jsonData, expiry).Err(); err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Add token to user's token set (for listing/revoking all user tokens)
	userKey := userTokensPrefix + data.UserID
	if err := redis.Client.SAdd(ctx, userKey, token).Err(); err != nil {
		return fmt.Errorf("failed to add token to user set: %w", err)
	}
	// Set expiry on user's token set (refresh if exists)
	redis.Client.Expire(ctx, userKey, expiry)

	return nil
}

// GetRefreshTokenData retrieves the data associated with a refresh token
func GetRefreshTokenData(ctx context.Context, token string) (*RefreshTokenData, error) {
	if redis.Client == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	key := refreshTokenPrefix + token
	jsonData, err := redis.Client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("refresh token not found or expired")
	}

	var data RefreshTokenData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token data: %w", err)
	}

	return &data, nil
}

// RevokeRefreshToken removes a refresh token from Redis
func RevokeRefreshToken(ctx context.Context, token string) error {
	if redis.Client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	// Get token data first to remove from user set
	data, err := GetRefreshTokenData(ctx, token)
	if err == nil && data != nil {
		userKey := userTokensPrefix + data.UserID
		redis.Client.SRem(ctx, userKey, token)
	}

	key := refreshTokenPrefix + token
	return redis.Client.Del(ctx, key).Err()
}

// RevokeAllUserTokens removes all refresh tokens for a user
func RevokeAllUserTokens(ctx context.Context, userID string) error {
	if redis.Client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	userKey := userTokensPrefix + userID
	tokens, err := redis.Client.SMembers(ctx, userKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get user tokens: %w", err)
	}

	// Delete all refresh tokens
	for _, token := range tokens {
		key := refreshTokenPrefix + token
		redis.Client.Del(ctx, key)
	}

	// Delete the user's token set
	return redis.Client.Del(ctx, userKey).Err()
}

// ListUserTokens returns all active refresh tokens for a user (for session management)
func ListUserTokens(ctx context.Context, userID string) ([]*RefreshTokenData, error) {
	if redis.Client == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	userKey := userTokensPrefix + userID
	tokens, err := redis.Client.SMembers(ctx, userKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user tokens: %w", err)
	}

	var result []*RefreshTokenData
	for _, token := range tokens {
		data, err := GetRefreshTokenData(ctx, token)
		if err != nil {
			// Token might have expired, clean it up
			redis.Client.SRem(ctx, userKey, token)
			continue
		}
		result = append(result, data)
	}

	return result, nil
}
