package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/chattycathy/api/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
)

var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
)

// Claims represents the JWT claims
type Claims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken           string `json:"access_token"`
	RefreshToken          string `json:"refresh_token"`
	AccessTokenExpiresIn  int    `json:"access_token_expires_in"`  // seconds
	RefreshTokenExpiresIn int    `json:"refresh_token_expires_in"` // seconds
	TokenType             string `json:"token_type"`
}

// Config holds JWT configuration
type Config struct {
	PrivateKeyPath         string
	PublicKeyPath          string
	Issuer                 string
	AccessTokenExpiryMins  int
	RefreshTokenExpiryDays int
}

// Init initializes the JWT module with RSA keys
func Init(cfg *Config) error {
	// Try to load existing keys
	if cfg.PrivateKeyPath != "" && cfg.PublicKeyPath != "" {
		if err := loadKeys(cfg.PrivateKeyPath, cfg.PublicKeyPath); err == nil {
			logger.Info().Msg("JWT keys loaded from files")
			return nil
		}
	}

	// Generate new keys if not found
	logger.Info().Msg("Generating new RSA key pair for JWT")
	var err error
	privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate RSA key: %w", err)
	}
	publicKey = &privateKey.PublicKey

	// Save keys if paths provided
	if cfg.PrivateKeyPath != "" && cfg.PublicKeyPath != "" {
		if err := saveKeys(cfg.PrivateKeyPath, cfg.PublicKeyPath); err != nil {
			logger.Warn().Err(err).Msg("Failed to save JWT keys to files")
		}
	}

	return nil
}

// loadKeys loads RSA keys from PEM files
func loadKeys(privatePath, publicPath string) error {
	// Load private key
	privateBytes, err := os.ReadFile(privatePath)
	if err != nil {
		return err
	}
	block, _ := pem.Decode(privateBytes)
	if block == nil {
		return errors.New("failed to decode private key PEM")
	}
	privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}

	// Load public key
	publicBytes, err := os.ReadFile(publicPath)
	if err != nil {
		return err
	}
	block, _ = pem.Decode(publicBytes)
	if block == nil {
		return errors.New("failed to decode public key PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	var ok bool
	publicKey, ok = pub.(*rsa.PublicKey)
	if !ok {
		return errors.New("not an RSA public key")
	}

	return nil
}

// saveKeys saves RSA keys to PEM files
func saveKeys(privatePath, publicPath string) error {
	// Save private key
	privateBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateBytes,
	})
	if err := os.WriteFile(privatePath, privatePEM, 0600); err != nil {
		return err
	}

	// Save public key
	publicBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return err
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicBytes,
	})
	if err := os.WriteFile(publicPath, publicPEM, 0644); err != nil {
		return err
	}

	return nil
}

// GenerateAccessToken creates a new short-lived JWT access token
func GenerateAccessToken(userID, username, role string, permissions []string, issuer string, expiryMins int) (string, error) {
	if privateKey == nil {
		return "", errors.New("JWT not initialized")
	}

	now := time.Now()
	claims := Claims{
		UserID:      userID,
		Username:    username,
		Role:        role,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expiryMins) * time.Minute)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

// GenerateRefreshToken creates a cryptographically secure refresh token
func GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateTokenPair creates both access and refresh tokens
func GenerateTokenPair(userID, username, role string, permissions []string, issuer string, accessExpiryMins, refreshExpiryDays int) (*TokenPair, error) {
	accessToken, err := GenerateAccessToken(userID, username, role, permissions, issuer, accessExpiryMins)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresIn:  accessExpiryMins * 60,
		RefreshTokenExpiresIn: refreshExpiryDays * 24 * 60 * 60,
		TokenType:             "Bearer",
	}, nil
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string) (*Claims, error) {
	if publicKey == nil {
		return nil, errors.New("JWT not initialized")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GetPublicKey returns the public key for external verification
func GetPublicKey() *rsa.PublicKey {
	return publicKey
}
