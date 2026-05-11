package main

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/liebeSonne/gophermart/internal/adapter"
	"github.com/liebeSonne/gophermart/internal/auth"
	"github.com/liebeSonne/gophermart/internal/config"
	"github.com/liebeSonne/gophermart/internal/handler"
	handlerauth "github.com/liebeSonne/gophermart/internal/handler/auth"
	"github.com/liebeSonne/gophermart/internal/handler/cookie"
	"github.com/liebeSonne/gophermart/internal/repository"
	"github.com/liebeSonne/gophermart/internal/repository/database"
	"github.com/liebeSonne/gophermart/internal/repository/uow"
	"github.com/liebeSonne/gophermart/internal/service"
	"github.com/liebeSonne/gophermart/internal/service/async"
)

const accrualServiceRetryMaxAttempts = 3
const accrualServiceRetryDelay = time.Second * 1

const dbClientRetryMaxAttempts = 3
const dbClientRetryDelay = time.Second * 1

type dependencyContainer struct {
	UserService                      handler.UserService
	UserOrderService                 handler.UserOrderService
	UserBalanceService               handler.UserBalanceService
	UserBalanceQueryService          handler.UserBalanceQueryService
	UserBalanceWithDrawnQueryService handler.UserBalanceWithDrawnQueryService
	UserOrderQueryService            handler.UserOrderQueryService
	TokenService                     handler.TokenService
	CookieService                    handler.CookieService

	AuthTokenService  handlerauth.TokenService
	AuthCookieService handlerauth.CookieService

	AccrualService    service.AccrualService
	UserOrderProvider service.UserOrderProvider
	UOWFactory        service.UnitOfWorkFactory

	RequestProducer async.Producer[string]
	RetryProducer   async.Producer[string]
	SetupProducer   service.OrderIDsProducer
	JobProducer     async.Producer[string]
}

func newDependencyContainer(
	cfg config.Config,
	logger *logrus.Logger,
	connection *connectionContainer,
) (*dependencyContainer, error) {
	dbPool := database.NewPool(connection.DBClient.Pool())
	dbPool = database.NewPoolRetryMiddleware(dbPool, dbClientRetryMaxAttempts, dbClientRetryDelay)
	dbClient := database.NewRetryMiddlewareContextClient(connection.DBClient.Pool(), dbClientRetryMaxAttempts, dbClientRetryDelay)

	uowFactory := uow.NewUnitOfWorkFactory(dbPool)

	userRepository := repository.NewUserRepository(dbClient)
	userOrderRepository := repository.NewUserOrderRepository(dbClient)
	userBalanceRepository := repository.NewUserBalanceRepository(dbClient)
	userBalanceWithDrawnRepository := repository.NewUserBalanceWithdrawnRepository(dbClient)

	accrualService := adapter.NewAccrualService(connection.AccrualClient)
	accrualService = service.NewRetryMiddlewareAccrualService(accrualService, accrualServiceRetryMaxAttempts, accrualServiceRetryDelay)

	requestProducer := NewRequestProducer(logger)
	retryProducer := NewRetryProducer(logger)
	setupProducer := NewSetupProducer(logger, userOrderRepository, retryProducer)
	jobProducer := NewJobProducer(logger)

	passwordService := service.NewPasswordService([]byte(cfg.PasswordSecretKey))
	userService := service.NewUserService(uowFactory, passwordService, userRepository)
	userOrderService := service.NewUserOrderService(uowFactory, requestProducer)
	userBalanceService := service.NewUserBalanceService(uowFactory)

	tokenService := auth.NewTokenService(cfg.AuthSecretKey, cfg.AuthTokenExpires)
	cookieService := cookie.NewService(cfg.AuthCookieTokenKey)

	return &dependencyContainer{
		UserService:                      userService,
		UserOrderService:                 userOrderService,
		UserBalanceService:               userBalanceService,
		UserBalanceQueryService:          userBalanceRepository,
		UserBalanceWithDrawnQueryService: userBalanceWithDrawnRepository,
		UserOrderQueryService:            userOrderRepository,
		TokenService:                     tokenService,
		CookieService:                    cookieService,

		AuthTokenService:  tokenService,
		AuthCookieService: cookieService,

		AccrualService:    accrualService,
		UserOrderProvider: userOrderRepository,
		UOWFactory:        uowFactory,

		RequestProducer: requestProducer,
		RetryProducer:   retryProducer,
		SetupProducer:   setupProducer,
		JobProducer:     jobProducer,
	}, nil
}
