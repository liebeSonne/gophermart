package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"

	"github.com/liebeSonne/gophermart/api/server"
	"github.com/liebeSonne/gophermart/internal/handler"
)

func initRouter(
	dependency *dependencyContainer,
	logger *logrus.Logger,
) (http.Handler, error) {
	s := handler.NewServer(
		dependency.UserService,
		dependency.UserOrderService,
		dependency.TokenService,
		dependency.CookieService,
		logger,
	)
	r := chi.NewMux()
	h := server.HandlerFromMux(s, r)
	return h, nil
}
