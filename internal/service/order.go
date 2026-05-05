package service

import (
	"context"
	"errors"
	"time"

	"github.com/liebeSonne/gophermart/internal/model"
	"github.com/liebeSonne/gophermart/internal/repository/uow"
)

var ErrUserOrderAlreadyUploadedByUser = errors.New("user order already uploaded by user")
var ErrUserOrderAlreadyUploadedByOtherUser = errors.New("user order already uploaded by other user")

type UserOrderService interface {
	Upload(ctx context.Context, input UploadUserOrderInput) (model.UserOrder, error)
}

func NewUserOrderService(
	uowFactory uow.UnitOfWorkFactory,
) UserOrderService {
	return &userOrderService{
		uowFactory: uowFactory,
	}
}

type userOrderService struct {
	uowFactory uow.UnitOfWorkFactory
}

func (u *userOrderService) Upload(ctx context.Context, input UploadUserOrderInput) (model.UserOrder, error) {
	err := input.Validate()
	if err != nil {
		return model.UserOrder{}, err
	}

	var newUserOrder model.UserOrder

	err = u.uowFactory.ExecuteWithUnitOfWork(ctx, func(provider uow.RepositoryProvider) error {
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

	return newUserOrder, nil
}
