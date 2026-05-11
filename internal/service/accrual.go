package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

var ErrUnexpectedError = errors.New("unexpected error")
var ErrUnknownOrderStatus = errors.New("unknown order status")
var ErrOrderNotFound = errors.New("order not found")
var ErrTooManyRequests = errors.New("too many requests")
var ErrServerError = errors.New("server error")

type ErrTooManyRequestsRetryAfter struct {
	Err        error
	RetryAfter time.Duration
}

func (e *ErrTooManyRequestsRetryAfter) Error() string {
	return fmt.Sprintf("retryAfter: %d: %v", e.RetryAfter, e.Err)
}

func (e *ErrTooManyRequestsRetryAfter) Unwrap() error {
	return e.Err
}

func NewErrTooManyRequestsRetryAfter(err error, retryAfter time.Duration) error {
	return &ErrTooManyRequestsRetryAfter{
		Err:        err,
		RetryAfter: retryAfter,
	}
}

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
