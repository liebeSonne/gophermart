package context

import (
	"context"

	"github.com/google/uuid"

	"github.com/liebeSonne/gophermart/internal/auth"
)

type tokenContextKey struct{}

var tokenKey = tokenContextKey{}

func CreateTokenContext(ctx context.Context, token auth.Token) context.Context {
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

func getTokenFromContext(ctx context.Context) (auth.Token, bool) {
	token, ok := ctx.Value(tokenKey).(auth.Token)
	if !ok {
		return auth.Token{}, false
	}

	return token, true
}
