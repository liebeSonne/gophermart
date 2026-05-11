package service

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"

	ilogger "github.com/liebeSonne/gophermart/internal/logger"
	"github.com/liebeSonne/gophermart/internal/service/async"
)

type OrderIDWorkerResult struct {
	OrderID string
	Err     error
	Status  *OrderStatus
	Accrual *decimal.Decimal
}

func NewOrderIDWorker(
	ctx context.Context,
	name string,
	minTooManyRequestsRetryDelay time.Duration,
	maxTooManyRequestsRetryDelay time.Duration,
	accrualService AccrualService,
	logger ilogger.Logger,
) async.Worker[string, OrderIDWorkerResult] {
	return &orderIDWorker{
		ctx:                          ctx,
		name:                         name,
		minTooManyRequestsRetryDelay: minTooManyRequestsRetryDelay,
		maxTooManyRequestsRetryDelay: maxTooManyRequestsRetryDelay,
		accrualService:               accrualService,
		logger:                       logger,
	}
}

type orderIDWorker struct {
	ctx                          context.Context
	name                         string
	minTooManyRequestsRetryDelay time.Duration
	maxTooManyRequestsRetryDelay time.Duration
	accrualService               AccrualService
	logger                       ilogger.Logger
}

func (w *orderIDWorker) Handle(orderID string) OrderIDWorkerResult {
	orderData, err := w.accrualService.GetOrders(w.ctx, orderID)
	w.logger.Debugf("'%s' worker get order data (%+v) error (%v)", w.name, orderData, err)

	if err != nil {
		return OrderIDWorkerResult{
			OrderID: orderID,
			Err:     err,
		}
	}

	return OrderIDWorkerResult{
		OrderID: orderID,
		Status:  &orderData.Status,
		Accrual: orderData.Accrual,
		Err:     nil,
	}
}

func (w *orderIDWorker) SleepingHandle(result OrderIDWorkerResult) (bool, time.Duration) {
	if result.Err != nil {
		var retryErr *ErrTooManyRequestsRetryAfter
		if errors.As(result.Err, &retryErr) && retryErr.RetryAfter > 0 {
			return true, min(max(w.minTooManyRequestsRetryDelay, retryErr.RetryAfter), w.maxTooManyRequestsRetryDelay)
		}

		if errors.Is(result.Err, ErrTooManyRequests) {
			return true, w.minTooManyRequestsRetryDelay
		}
	}

	return false, 0
}
