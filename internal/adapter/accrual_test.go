package adapter

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/liebeSonne/gophermart/api/client"
	"github.com/liebeSonne/gophermart/internal/service"
)

func TestAccrualAdapter_GetOrders(t *testing.T) {
	orderID1 := "123"
	error1 := errors.New("error 1")
	accrual1 := float64(10.5)
	accrualDecimal1 := decimal.NewFromFloat(accrual1)
	retryAfterStr1 := "60"
	retryAfterDuration1 := time.Duration(60) * time.Second

	type when struct {
		getOrder    *client.GetOrdersResponse
		getOrderErr error
	}
	type on struct {
		orderID string
	}
	type want struct {
		orderData service.OrderData
		err       error
	}
	testCases := []struct {
		name string
		on   on
		when when
		want want
	}{
		{
			"error",
			on{orderID1},
			when{getOrderErr: error1},
			want{err: error1},
		},
		{
			"registered",
			on{orderID1},
			when{getOrder: &client.GetOrdersResponse{
				HTTPResponse: &http.Response{
					StatusCode: http.StatusOK,
				},
				JSON200: &client.GetOrdersResponseData{
					Order:  orderID1,
					Status: client.REGISTERED,
				},
			}},
			want{orderData: service.OrderData{
				OrderID: orderID1,
				Status:  service.OrderStatusRegistered,
			}},
		},
		{
			"processing",
			on{orderID1},
			when{getOrder: &client.GetOrdersResponse{
				HTTPResponse: &http.Response{
					StatusCode: http.StatusOK,
				},
				JSON200: &client.GetOrdersResponseData{
					Order:  orderID1,
					Status: client.PROCESSING,
				},
			}},
			want{orderData: service.OrderData{
				OrderID: orderID1,
				Status:  service.OrderStatusProcessing,
			}},
		},
		{
			"processed",
			on{orderID1},
			when{getOrder: &client.GetOrdersResponse{
				HTTPResponse: &http.Response{
					StatusCode: http.StatusOK,
				},
				JSON200: &client.GetOrdersResponseData{
					Order:   orderID1,
					Status:  client.PROCESSED,
					Accrual: &accrual1,
				},
			}},
			want{orderData: service.OrderData{
				OrderID: orderID1,
				Status:  service.OrderStatusProcessed,
				Accrual: &accrualDecimal1,
			}},
		},
		{
			"processed",
			on{orderID1},
			when{getOrder: &client.GetOrdersResponse{
				HTTPResponse: &http.Response{
					StatusCode: http.StatusOK,
				},
				JSON200: &client.GetOrdersResponseData{
					Order:  orderID1,
					Status: client.INVALID,
				},
			}},
			want{orderData: service.OrderData{
				OrderID: orderID1,
				Status:  service.OrderStatusInvalid,
			}},
		},
		{
			"no content",
			on{orderID1},
			when{getOrder: &client.GetOrdersResponse{
				HTTPResponse: &http.Response{
					StatusCode: http.StatusNoContent,
				},
			}},
			want{err: service.ErrOrderNotFound},
		},
		{
			"server error",
			on{orderID1},
			when{getOrder: &client.GetOrdersResponse{
				HTTPResponse: &http.Response{
					StatusCode: http.StatusInternalServerError,
				},
			}},
			want{err: service.ErrServerError},
		},
		{
			"unexpected error on undefined status",
			on{orderID1},
			when{getOrder: &client.GetOrdersResponse{
				HTTPResponse: &http.Response{
					StatusCode: 0,
				},
			}},
			want{err: service.ErrUnexpectedError},
		},
		{
			"too many retries",
			on{orderID1},
			when{getOrder: &client.GetOrdersResponse{
				HTTPResponse: &http.Response{
					StatusCode: http.StatusTooManyRequests,
				},
			}},
			want{err: service.ErrTooManyRetries},
		},
		{
			"too many retries with retry after",
			on{orderID1},
			when{getOrder: &client.GetOrdersResponse{
				HTTPResponse: &http.Response{
					StatusCode: http.StatusTooManyRequests,
					Header: http.Header{
						"Retry-After": []string{retryAfterStr1},
					},
				},
			}},
			want{err: service.NewErrTooManyRetriesRetryAfter(service.ErrTooManyRetries, retryAfterDuration1)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			apiClient := NewMockAccrualClientWithResponsesInterface(t)
			apiClient.EXPECT().GetOrdersWithResponse(t.Context(), tc.on.orderID, mock.Anything).Return(tc.when.getOrder, tc.when.getOrderErr).Maybe()

			a := NewAccrualService(apiClient)

			orderData, err := a.GetOrders(t.Context(), tc.on.orderID)

			if tc.want.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.want.err.Error())
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want.orderData, orderData)
		})
	}
}
