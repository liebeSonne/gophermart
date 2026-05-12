package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	ilogger "github.com/liebeSonne/gophermart/internal/logger"
)

func TestOrderIDWorker_Handle(t *testing.T) {
	error1 := errors.New("errro 1")
	orderID1 := "123"
	statusRegistered := OrderStatusRegistered
	statusProcessing := OrderStatusProcessing
	statusInvalid := OrderStatusInvalid
	statusProcessed := OrderStatusProcessed
	accrual1 := decimal.NewFromFloat(10.5)

	type on struct {
		orderID string
		waiting time.Duration
	}
	type when struct {
		getOrder    OrderData
		getOrderErr error
	}
	type want struct {
		result OrderIDWorkerResult
	}
	testCases := []struct {
		name string
		on   on
		when when
		want want
	}{
		{
			"error on get order data",
			on{orderID1, time.Millisecond * 200},
			when{getOrderErr: error1},
			want{OrderIDWorkerResult{OrderID: orderID1, Err: error1}},
		},
		{
			"status registered",
			on{orderID1, time.Millisecond * 200},
			when{getOrder: OrderData{OrderID: orderID1, Status: statusRegistered, Accrual: nil}},
			want{OrderIDWorkerResult{OrderID: orderID1, Status: &statusRegistered, Accrual: nil, Err: nil}},
		},
		{
			"status processing",
			on{orderID1, time.Millisecond * 200},
			when{getOrder: OrderData{OrderID: orderID1, Status: statusProcessing, Accrual: nil}},
			want{OrderIDWorkerResult{OrderID: orderID1, Status: &statusProcessing, Accrual: nil, Err: nil}},
		},
		{
			"status invalid",
			on{orderID1, time.Millisecond * 200},
			when{getOrder: OrderData{OrderID: orderID1, Status: statusInvalid, Accrual: nil}},
			want{OrderIDWorkerResult{OrderID: orderID1, Status: &statusInvalid, Accrual: nil, Err: nil}},
		},
		{
			"status processed",
			on{orderID1, time.Millisecond * 200},
			when{getOrder: OrderData{OrderID: orderID1, Status: statusProcessed, Accrual: &accrual1}},
			want{OrderIDWorkerResult{OrderID: orderID1, Status: &statusProcessed, Accrual: &accrual1, Err: nil}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			accrualService := NewMockAccrualService(t)
			accrualService.EXPECT().GetOrders(mock.Anything, tc.on.orderID).Return(tc.when.getOrder, tc.when.getOrderErr).Once()

			l := ilogger.NewNullLogger()

			w := NewOrderIDWorker(ctx, "name", time.Second, time.Second, accrualService, l)

			outCh := make(chan OrderIDWorkerResult, 1)

			out := w.Handle(tc.on.orderID)
			outCh <- out

			values := make([]OrderIDWorkerResult, 0)
			timeout := time.After(tc.on.waiting)

		loop:
			for {
				select {
				case value, ok := <-outCh:
					if !ok {
						break loop
					}
					values = append(values, value)
				case <-ctx.Done():
					break loop
				case <-timeout:
					break loop
				}
			}

			require.Len(t, values, 1)
			orderData := values[0]

			assert.Equal(t, tc.want.result.OrderID, orderData.OrderID)
			assert.Equal(t, tc.want.result.Status, orderData.Status)
			assert.Equal(t, tc.want.result.Accrual, orderData.Accrual)
			if tc.want.result.Err != nil {
				require.Error(t, tc.want.result.Err)
				require.ErrorContains(t, orderData.Err, tc.want.result.Err.Error())
			} else {
				require.NoError(t, orderData.Err)
			}
		})
	}
}

func TestOrderIDWorker_SleepingHandle(t *testing.T) {
	minTooManyRequestsRetryDelay := time.Second * 10
	maxTooManyRequestsRetryDelay := time.Second * 30
	retryAfterLessThenMin1 := time.Second * 5
	retryAfterMoreThenMinAndLessThenMax1 := time.Second * 20
	retryAfterMoreThenMax1 := time.Second * 40

	type on struct {
		result OrderIDWorkerResult
	}
	type want struct {
		needDelay bool
		delayTime time.Duration
	}
	testCases := []struct {
		name string
		on   on
		want want
	}{
		{
			"on too many requests error",
			on{result: OrderIDWorkerResult{Err: ErrTooManyRequests}},
			want{true, minTooManyRequestsRetryDelay},
		},
		{
			"on too many requests error with retry after less then min",
			on{result: OrderIDWorkerResult{Err: NewErrTooManyRequestsRetryAfter(ErrTooManyRequests, retryAfterLessThenMin1)}},
			want{true, minTooManyRequestsRetryDelay},
		},
		{
			"on too many requests error with retry after more then max",
			on{result: OrderIDWorkerResult{Err: NewErrTooManyRequestsRetryAfter(ErrTooManyRequests, retryAfterMoreThenMax1)}},
			want{true, maxTooManyRequestsRetryDelay},
		},
		{
			"on too many requests error with retry after between mind and max",
			on{result: OrderIDWorkerResult{Err: NewErrTooManyRequestsRetryAfter(ErrTooManyRequests, retryAfterMoreThenMinAndLessThenMax1)}},
			want{true, retryAfterMoreThenMinAndLessThenMax1},
		},
		{
			"on unknown error",
			on{result: OrderIDWorkerResult{Err: errors.New("error 1")}},
			want{false, 0},
		},
		{
			"on no error",
			on{result: OrderIDWorkerResult{Err: nil}},
			want{false, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			accrualService := NewMockAccrualService(t)
			l := ilogger.NewNullLogger()

			w := NewOrderIDWorker(t.Context(), "name", minTooManyRequestsRetryDelay, maxTooManyRequestsRetryDelay, accrualService, l)

			needDelay, delayTime := w.SleepingHandle(tc.on.result)

			assert.Equal(t, tc.want.needDelay, needDelay)
			assert.Equal(t, tc.want.delayTime, delayTime)
		})
	}
}
