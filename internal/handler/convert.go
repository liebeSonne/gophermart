package handler

import (
	"errors"
	"fmt"

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

func convertDecimalToFloat64(amount decimal.Decimal) (float64, error) {
	accrualFloat64, _ := amount.Float64()
	return accrualFloat64, nil
}

func convertUserOrderToAPI(item model.UserOrder) (server.UserOrderData, error) {
	status, err := convertUserOrderStatusToAPI(item.Status)
	if err != nil {
		return server.UserOrderData{}, err
	}
	var accrualPtr *float64
	if item.Accrual != nil {
		accrual, err := convertDecimalToFloat64(*item.Accrual)
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
	balance, err := convertDecimalToFloat64(item.Balance)
	if err != nil {
		return server.GetUserBalanceResponse{}, fmt.Errorf("error on converting balance value (%v): %w", item.Balance, err)
	}
	withdrawnSum, err := convertDecimalToFloat64(item.WithdrawnSum)
	if err != nil {
		return server.GetUserBalanceResponse{}, fmt.Errorf("error on converting withdrawn sum value (%v): %w", item.WithdrawnSum, err)
	}

	return server.GetUserBalanceResponse{
		Current:   balance,
		Withdrawn: withdrawnSum,
	}, nil
}

func convertUserBalanceWithdrawnToAPI(item model.UserBalanceWithdrawn) (server.WithdrawalData, error) {
	amount, err := convertDecimalToFloat64(item.Amount)
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
