package auth

import (
	"net/http"

	"github.com/liebeSonne/gophermart/internal/auth"
	ilogger "github.com/liebeSonne/gophermart/internal/logger"
)

type TokenService interface {
	Parse(tokenString string) (auth.Token, error)
}

type CookieService interface {
	GetAuthToken(r *http.Request) (string, error)
}

func NewAuthMiddleware(
	next http.Handler,
	tokenService TokenService,
	cookieService CookieService,
	logger ilogger.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tokenString, err := cookieService.GetAuthToken(r)
		if err != nil {
			logger.Errorw("error getting auth token from cookie", "err", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if tokenString != "" {
			tokenData, err := tokenService.Parse(tokenString)
			if err == nil {
				ctx = auth.CreateTokenContext(ctx, tokenData)
			} else {
				logger.Errorw("parse token error", "err", err)
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
