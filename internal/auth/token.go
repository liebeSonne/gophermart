package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var ErrTokenIsNotValid = errors.New("token is not valid")

type Claims struct {
	jwt.RegisteredClaims
	Token
}

type Token struct {
	UserID string
}

type TokenService interface {
	Create(tokenData Token) (string, error)
	Parse(tokenString string) (Token, error)
}

func NewTokenService(
	secretKey string,
	tokenExpiry time.Duration,
) TokenService {
	return &tokenServiceImpl{
		secretKey:   secretKey,
		tokenExpiry: tokenExpiry,
	}
}

type tokenServiceImpl struct {
	secretKey   string
	tokenExpiry time.Duration
}

func (s *tokenServiceImpl) Create(tokenData Token) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenExpiry)),
		},
		Token: tokenData,
	})

	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *tokenServiceImpl) Parse(tokenString string) (Token, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})
	if err != nil {
		return Token{}, fmt.Errorf("parse token error: %w", err)
	}

	if !token.Valid {
		return Token{}, ErrTokenIsNotValid
	}

	return claims.Token, nil
}
