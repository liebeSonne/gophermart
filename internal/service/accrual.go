package service

import (
	"context"

	"github.com/shopspring/decimal"
)

type OrderStatus int

const (
	OrderStatusRegistered OrderStatus = iota
	OrderStatusProcessing
	OrderStatusInvalid
	OrderStatusProcessed
)

type OrderData struct {
	OrderID string
	Status  OrderStatus
	Accrual *decimal.Decimal
}

type AccrualService interface {
	GetOrders(ctx context.Context, orderID string) (OrderData, error)
}
