package handler

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"

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

func convertDecimalToFloat32(amount decimal.Decimal) (float32, error) {
	accrualFloat64, extract := amount.Float64()
	if !extract {
		return 0, fmt.Errorf("error on extract float64 from decimal amount (%v)", amount)
	}
	return float32(accrualFloat64), nil
}

func convertUserOrderToAPI(item model.UserOrder) (server.UserOrderData, error) {
	status, err := convertUserOrderStatusToAPI(item.Status)
	if err != nil {
		return server.UserOrderData{}, err
	}
	var accrualPtr *float32
	if item.Accrual != nil {
		accrual, err := convertDecimalToFloat32(*item.Accrual)
		if err != nil {
			return server.UserOrderData{}, fmt.Errorf("error on converting accrual value (%v): %w", *item.Accrual, err)
		}
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

func convertUserBalanceToAPI(item model.UserBalance) (server.GetUserBalanceResponse, error) {
	balance, err := convertDecimalToFloat32(item.Balance)
	if err != nil {
		return server.GetUserBalanceResponse{}, fmt.Errorf("error on converting balance value (%v): %w", item.Balance, err)
	}
	withdrawnSum, err := convertDecimalToFloat32(item.WithdrawnSum)
	if err != nil {
		return server.GetUserBalanceResponse{}, fmt.Errorf("error on converting withdrawn sum value (%v): %w", item.WithdrawnSum, err)
	}

	return server.GetUserBalanceResponse{
		Current:   balance,
		Withdrawn: withdrawnSum,
	}, nil
}

func convertUserBalanceWithdrawnToAPI(item model.UserBalanceWithdrawn) (server.WithdrawalData, error) {
	amount, err := convertDecimalToFloat32(item.Amount)
	if err != nil {
		return server.WithdrawalData{}, fmt.Errorf("error on converting amount value (%v): %w", item.Amount, err)
	}

	return server.WithdrawalData{
		Order:       item.OrderID,
		Sum:         amount,
		ProcessedAt: item.CreatedAt,
	}, nil
}

func convertUserBalanceWithdrawnItemsToAPI(items []model.UserBalanceWithdrawn) ([]server.WithdrawalData, error) {
	itemsData := make([]server.WithdrawalData, 0, len(items))
	for _, item := range items {
		itemData, err := convertUserBalanceWithdrawnToAPI(item)
		if err != nil {
			return server.GetUserWithdrawalsResponse{}, err
		}
		itemsData = append(itemsData, itemData)
	}

	return itemsData, nil
}
