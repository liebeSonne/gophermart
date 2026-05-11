package service

import (
	"context"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

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
	accrualService AccrualService,
	logger *logrus.Logger,
) async.Worker[string, OrderIDWorkerResult] {
	return &orderIDWorker{
		ctx:            ctx,
		name:           name,
		accrualService: accrualService,
		logger:         logger,
	}
}

type orderIDWorker struct {
	ctx            context.Context
	name           string
	accrualService AccrualService
	logger         *logrus.Logger
}

func (w *orderIDWorker) Handle(orderID string, resulCh chan<- OrderIDWorkerResult) {
	orderData, err := w.accrualService.GetOrders(w.ctx, orderID)
	w.logger.Debugf("'%s' worker get order data (%+v) error (%v)", w.name, orderData, err)

	if err != nil {
		resulCh <- OrderIDWorkerResult{
			OrderID: orderID,
			Err:     err,
		}
		return
	}

	resulCh <- OrderIDWorkerResult{
		OrderID: orderID,
		Status:  &orderData.Status,
		Accrual: orderData.Accrual,
		Err:     nil,
	}
}
