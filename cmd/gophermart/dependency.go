package main

import (
	"github.com/liebeSonne/gophermart/internal/auth"
	"github.com/liebeSonne/gophermart/internal/config"
	"github.com/liebeSonne/gophermart/internal/handler/cookie"
	"github.com/liebeSonne/gophermart/internal/repository"
	"github.com/liebeSonne/gophermart/internal/service"
)

type dependencyContainer struct {
	UserService   service.UserService
	TokenService  auth.TokenService
	CookieService cookie.Service
}

func newDependencyContainer(
	cfg config.Config,
	connection *connectionContainer,
) (*dependencyContainer, error) {
	userRepository := repository.NewUserRepository(connection.DBClient.Pool())
	passwordService := service.NewPasswordService([]byte(cfg.PasswordSecretKey))
	userService := service.NewUserService(userRepository, passwordService)
	tokenService := auth.NewTokenService(cfg.AuthSecretKey, cfg.AuthTokenExpires)
	cookieService := cookie.NewService(cfg.AuthCookieTokenKey)

	return &dependencyContainer{
		UserService:   userService,
		TokenService:  tokenService,
		CookieService: cookieService,
	}, nil
}
