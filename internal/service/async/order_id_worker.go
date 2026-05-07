package async

import (
	"context"

	"github.com/shopspring/decimal"

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
) Worker[string, OrderIDWorkerResult] {
	return &orderIDWorker{
		ctx:            ctx,
		accrualAdapter: accrualAdapter,
	}
}

type orderIDWorker struct {
	ctx            context.Context
	accrualAdapter adapter.AccrualAdapter
}

func (w *orderIDWorker) Handle(orderID string, resulCh chan<- OrderIDWorkerResult) {
	orderData, err := w.accrualAdapter.GetOrders(w.ctx, orderID)
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
