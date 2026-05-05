package main

import (
	"github.com/liebeSonne/gophermart/internal/auth"
	"github.com/liebeSonne/gophermart/internal/config"
	"github.com/liebeSonne/gophermart/internal/handler/cookie"
	"github.com/liebeSonne/gophermart/internal/repository"
	"github.com/liebeSonne/gophermart/internal/repository/uow"
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
	passwordService := service.NewPasswordService([]byte(cfg.PasswordSecretKey))
	uowFactory := uow.NewUnitOfWorkFactory(connection.DBClient.Pool())
	userProvider := repository.NewUserRepository(connection.DBClient.Pool())
	userService := service.NewUserService(uowFactory, passwordService, userProvider)
	tokenService := auth.NewTokenService(cfg.AuthSecretKey, cfg.AuthTokenExpires)
	cookieService := cookie.NewService(cfg.AuthCookieTokenKey)

	return &dependencyContainer{
		UserService:   userService,
		TokenService:  tokenService,
		CookieService: cookieService,
	}, nil
}
