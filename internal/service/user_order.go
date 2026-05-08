package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/liebeSonne/gophermart/internal/model"
	"github.com/liebeSonne/gophermart/internal/repository/uow"
	"github.com/liebeSonne/gophermart/internal/service/async"
)

var ErrUserOrderAlreadyUploadedByUser = errors.New("user order already uploaded by user")
var ErrUserOrderAlreadyUploadedByOtherUser = errors.New("user order already uploaded by other user")

type UserOrderService interface {
	Upload(ctx context.Context, input UploadUserOrderInput) (model.UserOrder, error)
}

func NewUserOrderService(
	uowFactory uow.UnitOfWorkFactory,
	requestProducer async.Producer[string],
) UserOrderService {
	return &userOrderService{
		uowFactory:      uowFactory,
		requestProducer: requestProducer,
	}
}

type userOrderService struct {
	uowFactory      uow.UnitOfWorkFactory
	requestProducer async.Producer[string]
}

func (u *userOrderService) Upload(ctx context.Context, input UploadUserOrderInput) (model.UserOrder, error) {
	err := input.Validate()
	if err != nil {
		return model.UserOrder{}, fmt.Errorf("invalid upload user order input: %w", err)
	}

	var newUserOrder model.UserOrder

	lockName := MakeUserOrderLockName(input.OrderID)
	lockNames := []string{lockName}

	err = u.uowFactory.ExecuteWithUnitOfWork(ctx, lockNames, func(provider uow.RepositoryProvider) error {
		repo := provider.UserOrderRepository()

		var userOrderPtr *model.UserOrder
		userOrderPtr, err = repo.FindByOrderID(ctx, input.OrderID)
		if err != nil {
			return err
		}

		if userOrderPtr != nil {
			if userOrderPtr.UserID == input.UserID {
				return ErrUserOrderAlreadyUploadedByUser
			}
			return ErrUserOrderAlreadyUploadedByOtherUser
		}

		newUserOrder = model.UserOrder{
			ID:        repo.NextID(ctx),
			OrderID:   input.OrderID,
			UserID:    input.UserID,
			Status:    model.OrderStatusNew,
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
