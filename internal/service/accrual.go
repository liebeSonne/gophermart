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
var ErrTooManyRetries = errors.New("too many retries")
var ErrServerError = errors.New("server error")

type ErrTooManyRetriesRetryAfter struct {
	Err        error
	RetryAfter time.Duration
}

func (e *ErrTooManyRetriesRetryAfter) Error() string {
	return fmt.Sprintf("retryAfter: %d: %v", e.RetryAfter, e.Err)
}

func (e *ErrTooManyRetriesRetryAfter) Unwrap() error {
	return e.Err
}

func NewErrTooManyRetriesRetryAfter(err error, retryAfter time.Duration) error {
	return &ErrTooManyRetriesRetryAfter{
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
