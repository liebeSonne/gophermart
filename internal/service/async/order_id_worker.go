package async

import (
	"context"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"github.com/liebeSonne/gophermart/internal/adapter"
)

type OrderIDWorkerResult struct {
	OrderID string
	Err     error
	Status  *adapter.OrderStatus
	Accrual *decimal.Decimal
}

func NewOrderIDWorker(
	ctx context.Context,
	accrualAdapter adapter.AccrualAdapter,
	logger *logrus.Logger,
) Worker[string, OrderIDWorkerResult] {
	return &orderIDWorker{
		ctx:            ctx,
		accrualAdapter: accrualAdapter,
		logger:         logger,
	}
}

type orderIDWorker struct {
	ctx            context.Context
	accrualAdapter adapter.AccrualAdapter
	logger         *logrus.Logger
}

func (w *orderIDWorker) Handle(orderID string, resulCh chan<- OrderIDWorkerResult) {
	orderData, err := w.accrualAdapter.GetOrders(w.ctx, orderID)
	w.logger.Debugf("order id worker get order data (%+v) error (%v)", orderData, err)

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
