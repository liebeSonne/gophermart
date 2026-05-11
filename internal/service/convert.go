package service

import (
	"errors"
	"fmt"

	"github.com/liebeSonne/gophermart/internal/model"
)

var ErrUnknownAccrualOrderStatus = errors.New("unknown accrual order status")

func ConvertOrderStatus(status OrderStatus) (model.OrderStatus, error) {
	switch status {
	case OrderStatusRegistered:
		return model.OrderStatusNew, nil
	case OrderStatusProcessing:
		return model.OrderStatusProcessing, nil
	case OrderStatusInvalid:
		return model.OrderStatusInvalid, nil
	case OrderStatusProcessed:
		return model.OrderStatusProcessed, nil
	default:
		return -1, fmt.Errorf("%w: %d", ErrUnknownAccrualOrderStatus, status)
	}
}
