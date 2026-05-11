package handler

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/liebeSonne/gophermart/internal/model"
)

var ErrUserOrderAlreadyUploadedByUser = errors.New("user order already uploaded by user")
var ErrUserOrderAlreadyUploadedByOtherUser = errors.New("user order already uploaded by other user")
var ErrInvalidOrderID = errors.New("invalid order ID")
var ErrInvalidUserID = errors.New("invalid user id")

type UploadUserOrderInput struct {
	OrderID string
	UserID  uuid.UUID
}

type UserOrderService interface {
	Upload(ctx context.Context, input UploadUserOrderInput) (model.UserOrder, error)
}

type UserOrderQueryService interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.UserOrder, error)
}
