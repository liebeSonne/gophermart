package auth

import (
	"context"

	"github.com/google/uuid"
)

type tokenContextKey struct{}

var tokenKey = tokenContextKey{}

func CreateTokenContext(ctx context.Context, token Token) context.Context {
	return context.WithValue(ctx, tokenKey, token)
}

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	tokenPtr, ok := getTokenFromContext(ctx)
	if !ok || tokenPtr.UserID == "" {
		return uuid.UUID{}, false
	}

	userID, err := uuid.Parse(tokenPtr.UserID)
	if err != nil {
		return uuid.UUID{}, false
	}

	return userID, true
}

func getTokenFromContext(ctx context.Context) (Token, bool) {
	token, ok := ctx.Value(tokenKey).(Token)
	if !ok {
		return Token{}, false
	}

	return token, true
}
