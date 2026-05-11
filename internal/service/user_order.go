package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/liebeSonne/gophermart/internal/handler"
	"github.com/liebeSonne/gophermart/internal/model"
	"github.com/liebeSonne/gophermart/internal/service/async"
)

type UserOrderRepository interface {
	NextID(ctx context.Context) uuid.UUID
	Store(ctx context.Context, items []model.UserOrder) error
	FindByOrderID(ctx context.Context, orderID string) (*model.UserOrder, error)
}

func NewUserOrderService(
	uowFactory UnitOfWorkFactory,
	requestProducer async.Producer[string],
) *UserOrderService {
	return &UserOrderService{
		uowFactory:      uowFactory,
		requestProducer: requestProducer,
	}
}

type UserOrderService struct {
	uowFactory      UnitOfWorkFactory
	requestProducer async.Producer[string]
}

func (u *UserOrderService) Upload(ctx context.Context, input handler.UploadUserOrderInput) (model.UserOrder, error) {
	validator := uploadUserOrderInputValidator{Input: input}
	err := validator.Validate()
	if err != nil {
		return model.UserOrder{}, fmt.Errorf("invalid upload user order input: %w", err)
	}

	var newUserOrder model.UserOrder

	lockName := MakeUserOrderLockName(input.OrderID)
	lockNames := []string{lockName}

	err = u.uowFactory.ExecuteWithUnitOfWork(ctx, lockNames, func(provider RepositoryProvider) error {
		repo := provider.UserOrderRepository()

		var userOrderPtr *model.UserOrder
		userOrderPtr, err = repo.FindByOrderID(ctx, input.OrderID)
		if err != nil {
			return err
		}

		if userOrderPtr != nil {
			if userOrderPtr.UserID == input.UserID {
				return handler.ErrUserOrderAlreadyUploadedByUser
			}
			return handler.ErrUserOrderAlreadyUploadedByOtherUser
		}

		newUserOrder = model.UserOrder{
			ID:        repo.NextID(ctx),
			OrderID:   input.OrderID,
			UserID:    input.UserID,
			Status:    model.UserOrderStatusNew,
			Accrual:   nil,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		return repo.Store(ctx, []model.UserOrder{newUserOrder})
	})
	if err != nil {
		return model.UserOrder{}, err
	}

	// Отправляем заказ из запроса на асинхронную обработку
	u.requestProducer.Add(newUserOrder.OrderID)

	return newUserOrder, nil
}
