package service

import (
	"context"
	"errors"
	"time"

	"github.com/avast/retry-go"
)

func NewRetryMiddlewareAccrualService(
	next AccrualService,
	maxAttempts uint,
	delay time.Duration,
) AccrualService {
	return &accrualServiceRetryMiddleware{
		next:        next,
		maxAttempts: maxAttempts,
		delay:       delay,
	}
}

type accrualServiceRetryMiddleware struct {
	next        AccrualService
	maxAttempts uint
	delay       time.Duration
}

func (s *accrualServiceRetryMiddleware) GetOrders(ctx context.Context, orderID string) (OrderData, error) {
	var resultOrderData OrderData

	err := retry.Do(
		func() error {
			orderData, err := s.next.GetOrders(ctx, orderID)
			if err != nil {
				return err
			}
			resultOrderData = orderData
			return nil
		},
		retry.Attempts(s.maxAttempts),
		retry.Delay(s.delay),
		retry.RetryIf(func(err error) bool {
			return errors.Is(err, ErrServerError) || errors.Is(err, ErrUnexpectedError)
		}),
	)
	if err != nil {
		return OrderData{}, err
	}

	return resultOrderData, err
}
