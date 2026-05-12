package handler

import (
	"errors"

	"github.com/shopspring/decimal"

	"github.com/liebeSonne/gophermart/api/server"
	"github.com/liebeSonne/gophermart/internal/model"
)

var orderStatusMap = map[model.UserOrderStatus]server.UserOrderDataStatus{
	model.UserOrderStatusNew:        server.NEW,
	model.UserOrderStatusProcessing: server.PROCESSING,
	model.UserOrderStatusInvalid:    server.INVALID,
	model.UserOrderStatusProcessed:  server.PROCESSED,
}

func convertUserOrderStatusToAPI(status model.UserOrderStatus) (server.UserOrderDataStatus, error) {
	result, ok := orderStatusMap[status]
	if !ok {
		return "", errors.New("unknown user order status")
	}
	return result, nil
}

func convertDecimalToFloat64(amount decimal.Decimal) float64 {
	accrualFloat64, _ := amount.Float64()
	return accrualFloat64
}

func convertUserOrderToAPI(item model.UserOrder) (server.UserOrderData, error) {
	status, err := convertUserOrderStatusToAPI(item.Status)
	if err != nil {
		return server.UserOrderData{}, err
	}
	var accrualPtr *float64
	if item.Accrual != nil {
		accrual := convertDecimalToFloat64(*item.Accrual)
		accrualPtr = &accrual
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

func convertUserBalanceToAPI(item model.UserBalance) server.GetUserBalanceResponse {
	balance := convertDecimalToFloat64(item.Balance)
	withdrawnSum := convertDecimalToFloat64(item.WithdrawnSum)
	return server.GetUserBalanceResponse{
		Current:   balance,
		Withdrawn: withdrawnSum,
	}
}

func convertUserBalanceWithdrawnToAPI(item model.UserBalanceWithdrawn) server.WithdrawalData {
	amount := convertDecimalToFloat64(item.Amount)
	return server.WithdrawalData{
		Order:       item.OrderID,
		Sum:         amount,
		ProcessedAt: item.CreatedAt,
	}
}

func convertUserBalanceWithdrawnItemsToAPI(items []model.UserBalanceWithdrawn) []server.WithdrawalData {
	itemsData := make([]server.WithdrawalData, 0, len(items))
	for _, item := range items {
		itemData := convertUserBalanceWithdrawnToAPI(item)
		itemsData = append(itemsData, itemData)
	}
	return itemsData
}
