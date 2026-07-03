// Package token defines the tokenizer port used by the auth module.
//
// The domain/application layer depends on the Tokenizer interface, not on JWT
// directly. This allows swapping to PASETO or opaque tokens later.
package token

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents the custom claims embedded in access tokens.
type Claims struct {
	UserID    uuid.UUID `json:"uid"`
	SessionID uuid.UUID `json:"sid"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	TokenType string    `json:"typ"`
	jwt.RegisteredClaims
}

// Tokenizer is the domain port for creating and validating tokens.
type Tokenizer interface {
	GenerateAccessToken(userID, sessionID uuid.UUID, email, role string, ttl time.Duration) (string, error)
	GenerateRefreshToken(sessionID uuid.UUID, ttl time.Duration) (string, error)
	ParseAccessToken(ctx context.Context, tokenString string) (*Claims, error)
	ParseRefreshToken(ctx context.Context, tokenString string) (uuid.UUID, error)
}

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

// NewJWT creates a JWT-based Tokenizer.
func NewJWT(secret string) Tokenizer {
	return &jwtTokenizer{secret: []byte(secret)}
}

type jwtTokenizer struct {
	secret []byte
}

func (t *jwtTokenizer) GenerateAccessToken(userID, sessionID uuid.UUID, email, role string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		SessionID: sessionID,
		Email:     email,
		Role:      role,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        uuid.Must(uuid.NewV7()).String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.secret)
}

func (t *jwtTokenizer) GenerateRefreshToken(sessionID uuid.UUID, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   sessionID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		ID:        uuid.Must(uuid.NewV7()).String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.secret)
}

func (t *jwtTokenizer) ParseAccessToken(ctx context.Context, tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return t.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}
	if !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims.TokenType != "access" {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func (t *jwtTokenizer) ParseRefreshToken(ctx context.Context, tokenString string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return t.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return uuid.Nil, ErrTokenExpired
		}
		return uuid.Nil, ErrInvalidToken
	}
	if !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}
	return uuid.Parse(claims.Subject)
}
