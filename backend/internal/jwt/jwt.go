package jwt

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const refreshTokenBytes = 64

type TokenManager struct {
	accessTokenSecret string
	accessTokenTTL    time.Duration
	refreshTokenTTL   time.Duration
}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

func NewTokenManager(accessSecret string, accessTTL, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{
		accessTokenSecret: accessSecret,
		accessTokenTTL:    accessTTL,
		refreshTokenTTL:   refreshTTL,
	}
}

func (tm *TokenManager) GenerateAccessToken(ctx context.Context, userID uuid.UUID) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tm.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(tm.accessTokenSecret))
}

func (tm *TokenManager) GenerateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	tokenBytes := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	return token, nil
}

func (tm *TokenManager) ValidateAccessToken(ctx context.Context, tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tm.accessTokenSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func (tm *TokenManager) ParseAccessToken(ctx context.Context, tokenString string) (*Claims, time.Time, error) {
	claims, err := tm.ValidateAccessToken(ctx, tokenString)
	if err != nil {
		return nil, time.Time{}, err
	}

	exp, err := claims.GetExpirationTime()
	if err != nil {
		return nil, time.Time{}, err
	}

	return claims, exp.Time, nil
}

func (tm *TokenManager) GetRefreshTokenTTL() time.Duration {
	return tm.refreshTokenTTL
}
