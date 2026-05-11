package handler

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/liebeSonne/gophermart/internal/model"
)

var ErrInvalidWithdrawnAmount = errors.New("invalid withdrawn amount")
var ErrUserBalanceIsNotEnough = errors.New("user balance is not enough")

type AddWithdrawnInput struct {
	UserID  uuid.UUID
	OrderID string
	Amount  decimal.Decimal
}

type UserBalanceService interface {
	AddWithdrawn(ctx context.Context, input AddWithdrawnInput) error
}

type UserBalanceQueryService interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (model.UserBalance, error)
}

type UserBalanceWithDrawnQueryService interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.UserBalanceWithdrawn, error)
}
