package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var ErrInvalidWithdrawnAmount = errors.New("invalid withdrawn amount")

type AddWithdrawnInput struct {
	UserID  uuid.UUID
	OrderID string
	Amount  decimal.Decimal
}

func (i *AddWithdrawnInput) Validate() error {
	if i.UserID == uuid.Nil {
		return ErrInvalidUserID
	}

	err := validateOrderID(i.OrderID)
	if err != nil {
		return err
	}

	if i.Amount.LessThan(decimal.Zero) {
		return ErrInvalidWithdrawnAmount
	}
	return nil
}
