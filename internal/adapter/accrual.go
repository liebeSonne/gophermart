package adapter

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/shopspring/decimal"

	"github.com/liebeSonne/gophermart/api/client"
	"github.com/liebeSonne/gophermart/internal/service"
)

func NewAccrualService(
	apiClient client.ClientWithResponsesInterface,
) service.AccrualService {
	return &accrualAdapter{
		client: apiClient,
	}
}

type accrualAdapter struct {
	client client.ClientWithResponsesInterface
}

func (a *accrualAdapter) GetOrders(ctx context.Context, orderID string) (service.OrderData, error) {
	resp, err := a.client.GetOrdersWithResponse(ctx, orderID)
	if err != nil {
		return service.OrderData{}, fmt.Errorf("unable to get order: %w", err)
	}

	statusCode := resp.StatusCode()
	switch statusCode {
	case http.StatusOK:
		orderData, err := a.convertOrderData(*resp.JSON200)
		if err != nil {
			return service.OrderData{}, fmt.Errorf("unable to convert order data: %w", err)
		}
		return orderData, nil
	case http.StatusNoContent:
		return service.OrderData{}, service.ErrOrderNotFound
	case http.StatusTooManyRequests:
		retryAfter := resp.HTTPResponse.Header.Get("Retry-After")
		if retryAfter != "" {
			seconds, err := strconv.Atoi(retryAfter)
			if err == nil {
				retryAfterDuration := time.Duration(seconds) * time.Second
				return service.OrderData{}, service.NewErrTooManyRequestsRetryAfter(service.ErrTooManyRequests, retryAfterDuration)
			}
		}
		return service.OrderData{}, service.ErrTooManyRequests
	case http.StatusInternalServerError:
		return service.OrderData{}, service.ErrServerError
	default:
		return service.OrderData{}, fmt.Errorf("unexpected status code (%d): %w", statusCode, service.ErrUnexpectedError)
	}
}

func (a *accrualAdapter) convertOrderData(orderData client.GetOrdersResponseData) (service.OrderData, error) {
	var status service.OrderStatus

	switch orderData.Status {
	case client.REGISTERED:
		status = service.OrderStatusRegistered
	case client.PROCESSING:
		status = service.OrderStatusProcessing
	case client.INVALID:
		status = service.OrderStatusInvalid
	case client.PROCESSED:
		status = service.OrderStatusProcessed
	default:
		return service.OrderData{}, fmt.Errorf("%w: %s", service.ErrUnknownOrderStatus, orderData.Status)
	}

	var accrualPtr *decimal.Decimal
	if orderData.Accrual != nil {
		accrual := decimal.NewFromFloat(*orderData.Accrual)
		accrualPtr = &accrual
	}

	return service.OrderData{
		OrderID: orderData.Order,
		Status:  status,
		Accrual: accrualPtr,
	}, nil
}
