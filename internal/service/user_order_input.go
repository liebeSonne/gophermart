package service

import (
	"errors"

	"github.com/google/uuid"
)

var ErrInvalidOrderID = errors.New("invalid order ID")
var ErrInvalidUserID = errors.New("invalid user id")

type UploadUserOrderInput struct {
	OrderID string
	UserID  uuid.UUID
}

func (i *UploadUserOrderInput) Validate() error {
	if i.UserID == uuid.Nil {
		return ErrInvalidUserID
	}

	err := validateOrderID(i.OrderID)
	if err != nil {
		return err
	}

	return nil
}

func validateOrderID(orderID string) error {
	if orderID == "" {
		return ErrInvalidOrderID
	}
	// TODO: добавить проверку валидности формата orderID по алгоритму Луна
	return nil
}
