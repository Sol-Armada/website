package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// Claims represents the JWT claims for a user session
type Claims struct {
	UserID    string   `json:"user_id"`
	DiscordID string   `json:"discord_id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	jwt.RegisteredClaims
}

// TokenService handles JWT token generation and validation
type TokenService struct {
	secretKey []byte
	issuer    string
	expiryDuration time.Duration
}

// NewTokenService creates a new token service
func NewTokenService(secretKey string, issuer string, expiryHours int) *TokenService {
	return &TokenService{
		secretKey:      []byte(secretKey),
		issuer:         issuer,
		expiryDuration: time.Duration(expiryHours) * time.Hour,
	}
}

// GenerateToken creates a new JWT token for a user
func (s *TokenService) GenerateToken(userID, discordID, username, email string, roles []string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		DiscordID: discordID,
		Username:  username,
		Email:     email,
		Roles:     roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiryDuration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateToken validates a JWT token and returns the claims
func (s *TokenService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshToken generates a new token with updated expiry
func (s *TokenService) RefreshToken(oldToken string) (string, error) {
	claims, err := s.ValidateToken(oldToken)
	if err != nil && !errors.Is(err, ErrExpiredToken) {
		return "", err
	}

	// Generate new token with same claims but fresh expiry
	return s.GenerateToken(claims.UserID, claims.DiscordID, claims.Username, claims.Email, claims.Roles)
}
