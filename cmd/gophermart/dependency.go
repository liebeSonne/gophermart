package main

import (
	"github.com/sirupsen/logrus"

	"github.com/liebeSonne/gophermart/internal/adapter"
	"github.com/liebeSonne/gophermart/internal/auth"
	"github.com/liebeSonne/gophermart/internal/config"
	"github.com/liebeSonne/gophermart/internal/handler/cookie"
	"github.com/liebeSonne/gophermart/internal/provider"
	"github.com/liebeSonne/gophermart/internal/repository"
	"github.com/liebeSonne/gophermart/internal/repository/uow"
	"github.com/liebeSonne/gophermart/internal/service"
	"github.com/liebeSonne/gophermart/internal/service/async"
)

type dependencyContainer struct {
	UserService                  service.UserService
	UserOrderService             service.UserOrderService
	UserOrderProvider            provider.UserOrderProvider
	UserBalanceProvider          provider.UserBalanceProvider
	UserBalanceService           service.UserBalanceService
	UserBalanceWithDrawnProvider provider.UserBalanceWithDrawnProvider
	TokenService                 auth.TokenService
	CookieService                cookie.Service
	AccrualAdapter               adapter.AccrualAdapter
	UOWFactory                   uow.UnitOfWorkFactory

	RequestProducer async.Producer[string]
	RetryProducer   async.Producer[string]
	SetupProducer   async.OrderIDsProducer
	JobProducer     async.Producer[string]
}

func newDependencyContainer(
	cfg config.Config,
	logger *logrus.Logger,
	connection *connectionContainer,
) (*dependencyContainer, error) {
	uowFactory := uow.NewUnitOfWorkFactory(connection.DBClient.Pool())

	userProvider := repository.NewUserRepository(connection.DBClient.Pool())
	userOrderProvider := repository.NewUserOrderRepository(connection.DBClient.Pool())
	userBalanceProvider := repository.NewUserBalanceRepository(connection.DBClient.Pool())
	userBalanceWithDrawnProvider := repository.NewUserBalanceWithdrawnRepository(connection.DBClient.Pool())

	accrualAdapter := adapter.NewAccrualAdapter(connection.AccrualClient)

	requestProducer := NewRequestProducer(logger)
	retryProducer := NewRetryProducer(logger)
	setupProducer := NewSetupProducer(logger, userOrderProvider)
	jobProducer := NewJobProducer(logger)

	passwordService := service.NewPasswordService([]byte(cfg.PasswordSecretKey))
	userService := service.NewUserService(uowFactory, passwordService, userProvider)
	userOrderService := service.NewUserOrderService(uowFactory, requestProducer)
	userBalanceService := service.NewUserBalanceService(uowFactory)

	tokenService := auth.NewTokenService(cfg.AuthSecretKey, cfg.AuthTokenExpires)
	cookieService := cookie.NewService(cfg.AuthCookieTokenKey)

	return &dependencyContainer{
		UserService:                  userService,
		UserOrderService:             userOrderService,
		UserOrderProvider:            userOrderProvider,
		UserBalanceProvider:          userBalanceProvider,
		UserBalanceService:           userBalanceService,
		UserBalanceWithDrawnProvider: userBalanceWithDrawnProvider,
		TokenService:                 tokenService,
		CookieService:                cookieService,
		AccrualAdapter:               accrualAdapter,
		UOWFactory:                   uowFactory,
		RequestProducer:              requestProducer,
		RetryProducer:                retryProducer,
		SetupProducer:                setupProducer,
		JobProducer:                  jobProducer,
	}, nil
}
