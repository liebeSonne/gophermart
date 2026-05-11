package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"

	"github.com/liebeSonne/gophermart/api/server"
	"github.com/liebeSonne/gophermart/internal/handler"
	"github.com/liebeSonne/gophermart/internal/handler/auth"
)

func initRouter(
	dependency *dependencyContainer,
	logger *logrus.Logger,
) (http.Handler, error) {
	s := handler.NewServer(
		dependency.UserService,
		dependency.UserOrderService,
		dependency.UserOrderQueryService,
		dependency.UserBalanceQueryService,
		dependency.UserBalanceService,
		dependency.UserBalanceWithDrawnQueryService,
		dependency.TokenService,
		dependency.CookieService,
		logger,
	)

	r := chi.NewMux()

	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: logger}))
	r.Use(middleware.AllowContentEncoding("deflate", "gzip"))
	r.Use(func(h http.Handler) http.Handler {
		return auth.NewAuthMiddleware(h, dependency.AuthTokenService, dependency.AuthCookieService, logger)
	})

	h := server.HandlerFromMux(s, r)
	return h, nil
}
