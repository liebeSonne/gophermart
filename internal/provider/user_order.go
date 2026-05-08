package provider

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/liebeSonne/gophermart/internal/model"
)

type UserOrderProvider interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.UserOrder, error)
	FindOrderIDToExecuteAtMap(ctx context.Context, spec model.FindUserOrderSpecification) (map[string]time.Time, error)
	FindUserIDByOrderID(ctx context.Context, orderID string) (*uuid.UUID, error)
}
