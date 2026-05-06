package provider

import (
	"context"

	"github.com/google/uuid"

	"github.com/liebeSonne/gophermart/internal/model"
)

type UserBalanceWithDrawnProvider interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.UserBalanceWithdrawn, error)
}
