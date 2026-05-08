package main

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/liebeSonne/gophermart/internal/adapter"
	"github.com/liebeSonne/gophermart/internal/provider"
	"github.com/liebeSonne/gophermart/internal/repository/uow"
	"github.com/liebeSonne/gophermart/internal/service"
	"github.com/liebeSonne/gophermart/internal/service/async"
)

// Названия процессов
const requestProducerName = "request"
const retryProducerName = "retry"
const setupProducerName = "setup"
const jobProducerName = "job"
const workerHandlerName = "worker"
const resultHandlerName = "result"

// Настройки request producer
const requestProducerChannelSize = 500

// Настройки retry producer
const retryProducerChannelSize = 1000

// Настройки setup producer
const setupProducerChannelSize = 500
const setupSelectLimit = 500
const setupLimitRetriesOnError = 3
const setupWaitingOnError = time.Second * 3

// Настройки job producer
const jobProducerChannelSize = 1000

// Настройки канала результатов
const resultChannelSize = 1000

// Настройки worker handler
const jobCountWorkers = 5

// Настройки result handler
const resultRetryDelay = time.Minute * 1
const resultTooManyRetriesDelay = time.Minute * 5
const resultCountWorkers = 3

func NewRequestProducer(
	logger *logrus.Logger,
) async.Producer[string] {
	return async.NewProducer[string](requestProducerName, requestProducerChannelSize, logger)
}

func NewRetryProducer(
	logger *logrus.Logger,
) async.Producer[string] {
	return async.NewProducer[string](retryProducerName, retryProducerChannelSize, logger)
}

func NewSetupProducer(
	logger *logrus.Logger,
	userOrderProvider provider.UserOrderProvider,
) async.OrderIDsProducer {
	limit := uint(setupSelectLimit)
	return async.NewOrderIDsProducer(
		setupProducerName,
		setupProducerChannelSize,
		&limit,
		setupLimitRetriesOnError,
		setupWaitingOnError,
		userOrderProvider,
		logger,
	)
}

func NewWorkerHandler(
	ctx context.Context,
	logger *logrus.Logger,
	accrualAdapter adapter.AccrualAdapter,
) async.WorkerHandler[string, async.OrderIDWorkerResult] {
	worker := async.NewOrderIDWorker(ctx, accrualAdapter, logger)
	return async.NewWorkerHandler[string, async.OrderIDWorkerResult](ctx, workerHandlerName, worker.Handle, logger)
}

func NewResultHandler(
	ctx context.Context,
	logger *logrus.Logger,
	retryProducer async.Producer[string],
	uowFactory uow.UnitOfWorkFactory,
	userOrderProvider provider.UserOrderProvider,
) async.WorkerHandler[async.OrderIDWorkerResult, struct{}] {
	worker := service.NewOrderIDResultWorker(
		ctx,
		resultRetryDelay,
		resultTooManyRetriesDelay,
		retryProducer,
		uowFactory,
		userOrderProvider,
		logger,
	)
	return async.NewWorkerHandler[async.OrderIDWorkerResult, struct{}](ctx, resultHandlerName, worker.Handle, logger)
}

func NewJobProducer(
	logger *logrus.Logger,
) async.Producer[string] {
	return async.NewProducer[string](jobProducerName, jobProducerChannelSize, logger)
}

func runProducers(
	ctx context.Context,
	dependency *dependencyContainer,
) (
	jobProducer async.Producer[string],
) {
	// Поставщик задач из http запросов
	requestProducer := dependency.RequestProducer
	requestProducer.Start(ctx)
	requestCh := requestProducer.Produce()

	// Поставщик задач из повторных попыток
	retryProducer := dependency.RetryProducer
	retryProducer.Start(ctx)
	retryCh := retryProducer.Produce()

	// Поставщик задач из БД
	setupProducer := dependency.SetupProducer
	setupProducer.Start(ctx)
	setupCh := setupProducer.Produce()

	// Сливаем всех поставщиков задач в один канал
	jobCh := async.FanIn(ctx, requestCh, retryCh, setupCh)

	// Формируем из общего канала задач одного поставщика, для удобства
	jobProducer = dependency.JobProducer
	jobProducer.Start(ctx)
	jobProducer.Setup(jobCh)

	return jobProducer
}

func runWorkers(
	ctx context.Context,
	logger *logrus.Logger,
	dependency *dependencyContainer,
) {
	jobCh := dependency.JobProducer.Produce()

	// Канал результатов
	resultCh := make(chan async.OrderIDWorkerResult, resultChannelSize)

	go func() {
		<-ctx.Done()
		close(resultCh)
	}()

	// Обработчик задач
	workerHandler := NewWorkerHandler(ctx, logger, dependency.AccrualAdapter)
	workerHandler.Handle(jobCh, resultCh, jobCountWorkers)

	// Обработчик результатов обработки - сохраняет результаты, и, при необходимости, планирует отправку на повторную проверку
	resultHandler := NewResultHandler(ctx, logger, dependency.RetryProducer, dependency.UOWFactory, dependency.UserOrderProvider)
	resultHandler.Handle(resultCh, make(chan<- struct{}), resultCountWorkers)
}
