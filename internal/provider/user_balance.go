package provider

import (
	"context"

	"github.com/google/uuid"

	"github.com/liebeSonne/gophermart/internal/model"
)

type UserBalanceProvider interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (model.UserBalance, error)
}
