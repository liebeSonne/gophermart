package handler

import (
	"errors"

	"github.com/liebeSonne/gophermart/api/server"
	"github.com/liebeSonne/gophermart/internal/model"
)

var orderStatusMap = map[model.OrderStatus]server.UserOrderDataStatus{
	model.OrderStatusNew:        server.NEW,
	model.OrderStatusProcessing: server.PROCESSING,
	model.OrderStatusInvalid:    server.INVALID,
	model.OrderStatusProcessed:  server.PROCESSED,
}

func convertUserOrderStatusToAPI(status model.OrderStatus) (server.UserOrderDataStatus, error) {
	result, ok := orderStatusMap[status]
	if !ok {
		return "", errors.New("unknown user order status")
	}
	return result, nil
}

func convertUserOrderToAPI(item model.UserOrder) (server.UserOrderData, error) {
	status, err := convertUserOrderStatusToAPI(item.Status)
	if err != nil {
		return server.UserOrderData{}, err
	}
	var accrualPtr *float32
	if item.Accrual != nil {
		accrualFloat64, extract := item.Accrual.Float64()
		if !extract {
			return server.UserOrderData{}, errors.New("error on extract accrual value")
		}
		accrualFloat32 := float32(accrualFloat64)
		accrualPtr = &accrualFloat32
	}

	return server.UserOrderData{
		Number:     item.OrderID,
		Status:     status,
		Accrual:    accrualPtr,
		UploadedAt: item.CreatedAt,
	}, nil
}

func convertUserOrdersToAPI(items []model.UserOrder) ([]server.UserOrderData, error) {
	itemsData := make([]server.UserOrderData, 0, len(items))
	for _, item := range items {
		itemData, err := convertUserOrderToAPI(item)
		if err != nil {
			return nil, err
		}
		itemsData = append(itemsData, itemData)
	}
	return itemsData, nil
}
