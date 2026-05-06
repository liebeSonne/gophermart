package adapter

import "github.com/shopspring/decimal"

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
