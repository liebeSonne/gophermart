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
	name string,
	accrualAdapter adapter.AccrualAdapter,
	logger *logrus.Logger,
) Worker[string, OrderIDWorkerResult] {
	return &orderIDWorker{
		ctx:            ctx,
		name:           name,
		accrualAdapter: accrualAdapter,
		logger:         logger,
	}
}

type orderIDWorker struct {
	ctx            context.Context
	name           string
	accrualAdapter adapter.AccrualAdapter
	logger         *logrus.Logger
}

func (w *orderIDWorker) Handle(orderID string, resulCh chan<- OrderIDWorkerResult) {
	orderData, err := w.accrualAdapter.GetOrders(w.ctx, orderID)
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
