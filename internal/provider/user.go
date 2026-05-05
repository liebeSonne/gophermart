package provider

import (
	"context"

	"github.com/google/uuid"

	"github.com/liebeSonne/gophermart/internal/model"
)

type UserProvider interface {
	GetByID(ctx context.Context, userID uuid.UUID) (model.User, error)
	FindByLogin(ctx context.Context, login string) (*model.User, error)
}
