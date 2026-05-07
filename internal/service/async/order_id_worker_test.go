package async

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/liebeSonne/gophermart/internal/adapter"
)

func TestOrderIDWorker_Handle(t *testing.T) {
	error1 := errors.New("errro 1")
	orderID1 := "123"
	statusRegistered := adapter.OrderStatusRegistered
	statusProcessing := adapter.OrderStatusProcessing
	statusInvalid := adapter.OrderStatusInvalid
	statusProcessed := adapter.OrderStatusProcessed
	accrual1 := decimal.NewFromFloat(10.5)

	type on struct {
		orderID string
		waiting time.Duration
	}
	type when struct {
		getOrder    adapter.OrderData
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
			when{getOrder: adapter.OrderData{OrderID: orderID1, Status: statusRegistered, Accrual: nil}},
			want{OrderIDWorkerResult{OrderID: orderID1, Status: &statusRegistered, Accrual: nil, Err: nil}},
		},
		{
			"status processing",
			on{orderID1, time.Millisecond * 200},
			when{getOrder: adapter.OrderData{OrderID: orderID1, Status: statusProcessing, Accrual: nil}},
			want{OrderIDWorkerResult{OrderID: orderID1, Status: &statusProcessing, Accrual: nil, Err: nil}},
		},
		{
			"status invalid",
			on{orderID1, time.Millisecond * 200},
			when{getOrder: adapter.OrderData{OrderID: orderID1, Status: statusInvalid, Accrual: nil}},
			want{OrderIDWorkerResult{OrderID: orderID1, Status: &statusInvalid, Accrual: nil, Err: nil}},
		},
		{
			"status processed",
			on{orderID1, time.Millisecond * 200},
			when{getOrder: adapter.OrderData{OrderID: orderID1, Status: statusProcessed, Accrual: &accrual1}},
			want{OrderIDWorkerResult{OrderID: orderID1, Status: &statusProcessed, Accrual: &accrual1, Err: nil}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			accrualAdapter := adapter.NewMockAccrualAdapter(t)
			accrualAdapter.EXPECT().GetOrders(mock.Anything, tc.on.orderID).Return(tc.when.getOrder, tc.when.getOrderErr).Once()

			w := NewOrderIDWorker(ctx, accrualAdapter)

			outCh := make(chan OrderIDWorkerResult, 1)

			w.Handle(tc.on.orderID, outCh)

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
