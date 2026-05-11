package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRetryMiddlewareAccrualService_GetOrders(t *testing.T) {
	orderID1 := "123"

	type on struct {
		orderID     string
		maxAttempts uint
		delay       time.Duration
	}
	type resultData struct {
		getOrdersData OrderData
		getOrdersErr  error
	}
	type when struct {
		results []resultData
	}
	type want struct {
		orderData OrderData
		err       error
		times     int
	}
	testCases := []struct {
		name string
		on   on
		when when
		want want
	}{
		{
			"on server error",
			on{orderID1, 3, time.Millisecond * 10},
			when{[]resultData{
				{OrderData{}, ErrServerError},
				{OrderData{}, ErrServerError},
				{OrderData{}, ErrServerError},
			}},
			want{OrderData{}, ErrServerError, 3},
		},
		{
			"on unexpected error",
			on{orderID1, 3, time.Millisecond * 10},
			when{[]resultData{
				{OrderData{}, ErrUnexpectedError},
				{OrderData{}, ErrUnexpectedError},
				{OrderData{}, ErrUnexpectedError},
			}},
			want{OrderData{}, ErrUnexpectedError, 3},
		},
		{
			"result on retry",
			on{orderID1, 5, time.Millisecond * 10},
			when{[]resultData{
				{OrderData{}, ErrServerError},
				{OrderData{}, ErrUnexpectedError},
				{OrderData{OrderID: orderID1, Status: OrderStatusRegistered, Accrual: nil}, nil},
			}},
			want{OrderData{OrderID: orderID1, Status: OrderStatusRegistered, Accrual: nil}, nil, 3},
		},
		{
			"result on first try",
			on{orderID1, 5, time.Millisecond * 10},
			when{[]resultData{
				{OrderData{OrderID: orderID1, Status: OrderStatusRegistered, Accrual: nil}, nil},
			}},
			want{OrderData{OrderID: orderID1, Status: OrderStatusRegistered, Accrual: nil}, nil, 1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			next := NewMockAccrualService(t)
			resultIndex := 0
			next.EXPECT().GetOrders(mock.Anything, tc.on.orderID).RunAndReturn(func(_ context.Context, _ string) (OrderData, error) {
				require.True(t, len(tc.when.results) > resultIndex)
				result := tc.when.results[resultIndex]
				resultIndex++
				return result.getOrdersData, result.getOrdersErr
			}).Times(tc.want.times)

			s := NewRetryMiddlewareAccrualService(next, tc.on.maxAttempts, tc.on.delay)
			orderData, err := s.GetOrders(t.Context(), tc.on.orderID)

			if tc.want.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.want.err.Error())
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want.orderData.OrderID, orderData.OrderID)
			assert.Equal(t, tc.want.orderData.Status, orderData.Status)
			assert.Equal(t, tc.want.orderData.Accrual, orderData.Accrual)
		})
	}
}
