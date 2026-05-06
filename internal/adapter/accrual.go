package adapter

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/shopspring/decimal"

	"github.com/liebeSonne/gophermart/api/client"
)

type AccrualAdapter interface {
	GetOrders(ctx context.Context, orderID string) (OrderData, error)
}

func NewAccrualAdapter(
	apiClient client.ClientWithResponsesInterface,
) AccrualAdapter {
	return &accrualAdapter{
		client: apiClient,
	}
}

type accrualAdapter struct {
	client client.ClientWithResponsesInterface
}

func (a *accrualAdapter) GetOrders(ctx context.Context, orderID string) (OrderData, error) {
	resp, err := a.client.GetOrdersWithResponse(ctx, orderID)
	if err != nil {
		return OrderData{}, fmt.Errorf("unable to get order: %w", err)
	}

	statusCode := resp.StatusCode()
	switch statusCode {
	case http.StatusOK:
		orderData, err := a.convertOrderData(*resp.JSON200)
		if err != nil {
			return OrderData{}, fmt.Errorf("unable to convert order data: %w", err)
		}
		return orderData, nil
	case http.StatusNoContent:
		return OrderData{}, ErrOrderNotFound
	case http.StatusTooManyRequests:
		retryAfter := resp.HTTPResponse.Header.Get("Retry-After")
		if retryAfter != "" {
			seconds, err := strconv.Atoi(retryAfter)
			if err != nil {
				retryAfterDuration := time.Duration(seconds) * time.Second
				return OrderData{}, NewErrTooManyRetriesRetryAfter(ErrTooManyRetries, retryAfterDuration)
			}
		}
		return OrderData{}, ErrTooManyRetries
	case http.StatusInternalServerError:
		return OrderData{}, ErrServerError
	default:
		return OrderData{}, fmt.Errorf("unexpected status code (%d): %w", statusCode, ErrUnexpectedError)
	}
}

func (a *accrualAdapter) convertOrderData(orderData client.GetOrdersResponseData) (OrderData, error) {
	var status OrderStatus

	switch orderData.Status {
	case client.REGISTERED:
		status = OrderStatusRegistered
	case client.PROCESSING:
		status = OrderStatusProcessing
	case client.INVALID:
		status = OrderStatusInvalid
	case client.PROCESSED:
		status = OrderStatusProcessed
	default:
		return OrderData{}, fmt.Errorf("%w: %s", ErrUnknownOrderStatus, orderData.Status)
	}

	var accrualPtr *decimal.Decimal
	if orderData.Accrual != nil {
		accrual := decimal.NewFromFloat32(*orderData.Accrual)
		accrualPtr = &accrual
	}

	return OrderData{
		OrderID: orderData.Order,
		Status:  status,
		Accrual: accrualPtr,
	}, nil
}
