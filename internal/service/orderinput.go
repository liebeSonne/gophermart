package service

import (
	"errors"

	"github.com/google/uuid"
)

var ErrInvalidOrderID = errors.New("invalid order ID")

type UploadUserOrderInput struct {
	OrderID string
	UserID  uuid.UUID
}

func (i *UploadUserOrderInput) Validate() error {
	if i.OrderID == "" {
		return ErrInvalidOrderID
	}
	// TODO: добавить проверку валидности формата orderID по алгоритму Луна
	return nil
}
