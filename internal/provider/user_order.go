package provider

import (
	"context"

	"github.com/google/uuid"

	"github.com/liebeSonne/gophermart/internal/model"
)

type UserOrderProvider interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.UserOrder, error)
	FindOrderIDs(ctx context.Context, spec model.FindUserOrderSpecification) ([]string, error)
	FindUserIDByOrderID(ctx context.Context, orderID string) (*uuid.UUID, error)
}
