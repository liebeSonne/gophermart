package auth

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/liebeSonne/gophermart/internal/auth"
	"github.com/liebeSonne/gophermart/internal/handler/cookie"
)

func NewAuthMiddleware(
	next http.Handler,
	tokenService auth.TokenService,
	cookieService cookie.Service,
	logger *logrus.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tokenString, err := cookieService.GetAuthToken(r)
		if err != nil {
			logger.WithError(err).Error("error getting auth token from cookie")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if tokenString != "" {
			tokenData, err := tokenService.Parse(tokenString)
			if err == nil {
				ctx = auth.CreateTokenContext(ctx, tokenData)
			} else {
				logger.WithError(err).Error("parse token error")
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
