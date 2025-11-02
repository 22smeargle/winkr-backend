package utils

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims structure
type Claims struct {
	UserID        string `json:"user_id"`
	Email         string `json:"email"`
	IsAdmin       bool   `json:"is_admin"`
	TokenType     string `json:"token_type"` // "access" or "refresh"
	DeviceID      string `json:"device_id,omitempty"`
	SessionID     string `json:"session_id,omitempty"`
	JTI           string `json:"jti"` // JWT ID for token identification
	jwt.RegisteredClaims
}

// DeviceInfo represents device information for fingerprinting
type DeviceInfo struct {
	UserAgent   string `json:"user_agent"`
	IPAddress   string `json:"ip_address"`
	Platform    string `json:"platform"`
	Device      string `json:"device"`
	Browser     string `json:"browser"`
	Fingerprint string `json:"fingerprint"`
}

// TokenBlacklist represents a token blacklist interface
type TokenBlacklist interface {
	IsBlacklisted(ctx interface{}, jti string) (bool, error)
	BlacklistToken(ctx interface{}, jti string, expiry time.Time) error
	RemoveFromBlacklist(ctx interface{}, jti string) error
}

// JWTUtils provides JWT token generation and validation utilities
type JWTUtils struct {
	secretKey             string
	accessTokenExpiry     time.Duration
	refreshTokenExpiry    time.Duration
	blacklist             TokenBlacklist
}

// NewJWTUtils creates a new JWTUtils instance
func NewJWTUtils(secretKey string, accessTokenExpiry, refreshTokenExpiry time.Duration, blacklist TokenBlacklist) *JWTUtils {
	return &JWTUtils{
		secretKey:          secretKey,
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
		blacklist:          blacklist,
	}
}

// NewJWTUtilsWithoutBlacklist creates a new JWTUtils instance without blacklist
func NewJWTUtilsWithoutBlacklist(secretKey string, accessTokenExpiry, refreshTokenExpiry time.Duration) *JWTUtils {
	return &JWTUtils{
		secretKey:          secretKey,
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
		blacklist:          nil,
	}
}

// GenerateAccessToken generates a new access token
func (j *JWTUtils) GenerateAccessToken(userID, email string, isAdmin bool) (string, error) {
	return j.generateToken(userID, email, isAdmin, "access", j.accessTokenExpiry, "", "")
}

// GenerateRefreshToken generates a new refresh token
func (j *JWTUtils) GenerateRefreshToken(userID, email string, isAdmin bool) (string, error) {
	return j.generateToken(userID, email, isAdmin, "refresh", j.refreshTokenExpiry, "", "")
}

// GenerateAccessTokenWithDevice generates a new access token with device information
func (j *JWTUtils) GenerateAccessTokenWithDevice(userID, email string, isAdmin bool, deviceID, sessionID string) (string, error) {
	return j.generateToken(userID, email, isAdmin, "access", j.accessTokenExpiry, deviceID, sessionID)
}

// GenerateRefreshTokenWithDevice generates a new refresh token with device information
func (j *JWTUtils) GenerateRefreshTokenWithDevice(userID, email string, isAdmin bool, deviceID, sessionID string) (string, error) {
	return j.generateToken(userID, email, isAdmin, "refresh", j.refreshTokenExpiry, deviceID, sessionID)
}

// GenerateTokenPair generates both access and refresh tokens
func (j *JWTUtils) GenerateTokenPair(userID, email string, isAdmin bool) (accessToken, refreshToken string, err error) {
	accessToken, err = j.GenerateAccessToken(userID, email, isAdmin)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err = j.GenerateRefreshToken(userID, email, isAdmin)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// GenerateTokenPairWithDevice generates both access and refresh tokens with device information
func (j *JWTUtils) GenerateTokenPairWithDevice(userID, email string, isAdmin bool, deviceID, sessionID string) (accessToken, refreshToken string, err error) {
	accessToken, err = j.GenerateAccessTokenWithDevice(userID, email, isAdmin, deviceID, sessionID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err = j.GenerateRefreshTokenWithDevice(userID, email, isAdmin, deviceID, sessionID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// generateToken generates a JWT token with the specified parameters
func (j *JWTUtils) generateToken(userID, email string, isAdmin bool, tokenType string, expiry time.Duration, deviceID, sessionID string) (string, error) {
	now := time.Now()
	jti := fmt.Sprintf("%s-%d", userID, now.UnixNano())
	
	claims := Claims{
		UserID:    userID,
		Email:     email,
		IsAdmin:   isAdmin,
		TokenType: tokenType,
		DeviceID:  deviceID,
		SessionID: sessionID,
		JTI:       jti,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "winkr-backend",
			Subject:   userID,
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTUtils) ValidateToken(tokenString string) (*Claims, error) {
	return j.validateTokenWithContext(tokenString, nil)
}

// ValidateTokenWithContext validates a JWT token with context and returns the claims
func (j *JWTUtils) ValidateTokenWithContext(tokenString string, ctx interface{}) (*Claims, error) {
	return j.validateTokenWithContext(tokenString, ctx)
}

// validateTokenWithContext validates a JWT token with optional context for blacklist checking
func (j *JWTUtils) validateTokenWithContext(tokenString string, ctx interface{}) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Check if token is blacklisted
		if j.blacklist != nil && ctx != nil {
			isBlacklisted, err := j.blacklist.IsBlacklisted(ctx, claims.JTI)
			if err != nil {
				return nil, fmt.Errorf("failed to check blacklist: %w", err)
			}
			if isBlacklisted {
				return nil, fmt.Errorf("token is blacklisted")
			}
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// ValidateAccessToken validates an access token
func (j *JWTUtils) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "access" {
		return nil, fmt.Errorf("invalid token type: expected access token")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (j *JWTUtils) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("invalid token type: expected refresh token")
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token from a valid refresh token
func (j *JWTUtils) RefreshAccessToken(refreshTokenString string) (string, error) {
	claims, err := j.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	return j.GenerateAccessToken(claims.UserID, claims.Email, claims.IsAdmin)
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is empty")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return authHeader[len(bearerPrefix):], nil
}

// GetTokenExpiration returns the expiration time of a token
func (j *JWTUtils) GetTokenExpiration(tokenString string) (*time.Time, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	return &claims.ExpiresAt.Time, nil
}

// IsTokenExpired checks if a token is expired
func (j *JWTUtils) IsTokenExpired(tokenString string) bool {
	expiration, err := j.GetTokenExpiration(tokenString)
	if err != nil {
		return true
	}

	return time.Now().After(*expiration)
}

// GetTokenRemainingTime returns the remaining time until token expiration
func (j *JWTUtils) GetTokenRemainingTime(tokenString string) (time.Duration, error) {
	expiration, err := j.GetTokenExpiration(tokenString)
	if err != nil {
		return 0, err
	}

	remaining := time.Until(*expiration)
	if remaining < 0 {
		return 0, fmt.Errorf("token is expired")
	}

	return remaining, nil
}

// BlacklistToken adds a token to the blacklist
func (j *JWTUtils) BlacklistToken(ctx interface{}, tokenString string) error {
	if j.blacklist == nil {
		return nil // No blacklist configured
	}

	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return fmt.Errorf("failed to validate token for blacklisting: %w", err)
	}

	expiration := claims.ExpiresAt.Time
	return j.blacklist.BlacklistToken(ctx, claims.JTI, expiration)
}

// RemoveFromBlacklist removes a token from the blacklist
func (j *JWTUtils) RemoveFromBlacklist(ctx interface{}, jti string) error {
	if j.blacklist == nil {
		return nil // No blacklist configured
	}

	return j.blacklist.RemoveFromBlacklist(ctx, jti)
}

// IsTokenBlacklisted checks if a token is blacklisted
func (j *JWTUtils) IsTokenBlacklisted(ctx interface{}, tokenString string) (bool, error) {
	if j.blacklist == nil {
		return false, nil // No blacklist configured
	}

	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return false, fmt.Errorf("failed to validate token: %w", err)
	}

	return j.blacklist.IsBlacklisted(ctx, claims.JTI)
}

// GenerateDeviceFingerprint generates a device fingerprint from device info
func (j *JWTUtils) GenerateDeviceFingerprint(deviceInfo *DeviceInfo) string {
	data := fmt.Sprintf("%s|%s|%s|%s|%s",
		deviceInfo.UserAgent,
		deviceInfo.IPAddress,
		deviceInfo.Platform,
		deviceInfo.Device,
		deviceInfo.Browser)
	
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// ParseDeviceInfo parses user agent string to extract device information
func (j *JWTUtils) ParseDeviceInfo(userAgent, ipAddress string) *DeviceInfo {
	// Simple parsing - in production, you might want to use a more sophisticated library
	platform := "unknown"
	device := "unknown"
	browser := "unknown"

	ua := strings.ToLower(userAgent)
	
	// Detect platform
	if strings.Contains(ua, "windows") {
		platform = "windows"
	} else if strings.Contains(ua, "mac") || strings.Contains(ua, "darwin") {
		platform = "macos"
	} else if strings.Contains(ua, "linux") {
		platform = "linux"
	} else if strings.Contains(ua, "android") {
		platform = "android"
		device = "mobile"
	} else if strings.Contains(ua, "ios") || strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") {
		platform = "ios"
		device = "mobile"
	}

	// Detect browser
	if strings.Contains(ua, "chrome") && !strings.Contains(ua, "edg") {
		browser = "chrome"
	} else if strings.Contains(ua, "firefox") {
		browser = "firefox"
	} else if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		browser = "safari"
	} else if strings.Contains(ua, "edg") {
		browser = "edge"
	}

	deviceInfo := &DeviceInfo{
		UserAgent: userAgent,
		IPAddress: ipAddress,
		Platform:  platform,
		Device:    device,
		Browser:   browser,
	}

	deviceInfo.Fingerprint = j.GenerateDeviceFingerprint(deviceInfo)
	return deviceInfo
}

// GetTokenID extracts the JTI (JWT ID) from a token
func (j *JWTUtils) GetTokenID(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("failed to validate token: %w", err)
	}
	return claims.JTI, nil
}

// GetDeviceIDFromToken extracts the device ID from a token
func (j *JWTUtils) GetDeviceIDFromToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("failed to validate token: %w", err)
	}
	return claims.DeviceID, nil
}

// GetSessionIDFromToken extracts the session ID from a token
func (j *JWTUtils) GetSessionIDFromToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("failed to validate token: %w", err)
	}
	return claims.SessionID, nil
}