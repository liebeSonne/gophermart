package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/liebeSonne/gophermart/internal/model"
	"github.com/liebeSonne/gophermart/internal/repository/uow"
)

var ErrUserBalanceIsNotEnough = errors.New("user balance is not enough")

type UserBalanceService interface {
	AddWithdrawn(ctx context.Context, input AddWithdrawnInput) error
}

func NewUserBalanceService(
	uowFactory uow.UnitOfWorkFactory,
) UserBalanceService {
	return &userBalanceService{
		uowFactory: uowFactory,
	}
}

type userBalanceService struct {
	uowFactory uow.UnitOfWorkFactory
}

func (s *userBalanceService) AddWithdrawn(ctx context.Context, input AddWithdrawnInput) error {
	err := input.Validate()
	if err != nil {
		return fmt.Errorf("invalid add withdrawn input: %w", err)
	}

	return s.uowFactory.ExecuteWithUnitOfWork(ctx, func(provider uow.RepositoryProvider) error {
		balanceRepository := provider.UserBalanceRepository()
		withdrawnRepository := provider.UserBalanceWithdrawnRepository()

		var balance model.UserBalance
		balance, err = balanceRepository.GetByUserID(ctx, input.UserID)
		if err != nil {
			return err
		}

		if balance.Balance.LessThan(input.Amount) {
			return ErrUserBalanceIsNotEnough
		}

		withdrawn := model.UserBalanceWithdrawn{
			ID:        withdrawnRepository.NextID(ctx),
			UserID:    input.UserID,
			OrderID:   input.OrderID,
			Amount:    input.Amount,
			CreatedAt: time.Now(),
		}

		balance.Balance = balance.Balance.Sub(input.Amount)
		balance.WithdrawnSum = balance.WithdrawnSum.Add(input.Amount)

		err = withdrawnRepository.Store(ctx, withdrawn)
		if err != nil {
			return fmt.Errorf("cannot store withdrawn amount: %w", err)
		}

		err = balanceRepository.Store(ctx, balance)
		if err != nil {
			return fmt.Errorf("cannot store balance: %w", err)
		}

		return nil
	})
}
