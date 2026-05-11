package service

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/liebeSonne/gophermart/internal/handler"
)

type addWithdrawnInputValidator struct {
	Input handler.AddWithdrawnInput
}

func (i *addWithdrawnInputValidator) Validate() error {
	if i.Input.UserID == uuid.Nil {
		return handler.ErrInvalidUserID
	}

	err := validateOrderID(i.Input.OrderID)
	if err != nil {
		return err
	}

	if i.Input.Amount.LessThan(decimal.Zero) {
		return handler.ErrInvalidWithdrawnAmount
	}
	return nil
}
