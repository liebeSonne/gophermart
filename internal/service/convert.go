package service

import (
	"errors"
	"fmt"

	"github.com/liebeSonne/gophermart/internal/adapter"
	"github.com/liebeSonne/gophermart/internal/model"
)

var ErrUnknownAccrualOrderStatus = errors.New("unknown accrual order status")

func ConvertOrderStatus(status adapter.OrderStatus) (model.OrderStatus, error) {
	switch status {
	case adapter.OrderStatusRegistered:
		return model.OrderStatusNew, nil
	case adapter.OrderStatusProcessing:
		return model.OrderStatusProcessing, nil
	case adapter.OrderStatusInvalid:
		return model.OrderStatusInvalid, nil
	case adapter.OrderStatusProcessed:
		return model.OrderStatusProcessed, nil
	default:
		return -1, fmt.Errorf("%w: %d", ErrUnknownAccrualOrderStatus, status)
	}
}
