package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/liebeSonne/gophermart/internal/handler"
	"github.com/liebeSonne/gophermart/internal/model"
)

type UserBalanceRepository interface {
	Store(ctx context.Context, item model.UserBalance) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (model.UserBalance, error)
}

type UserBalanceWithdrawnRepository interface {
	NextID(ctx context.Context) uuid.UUID
	Store(ctx context.Context, item model.UserBalanceWithdrawn) error
}

func NewUserBalanceService(
	uowFactory UnitOfWorkFactory,
) *UserBalanceService {
	return &UserBalanceService{
		uowFactory: uowFactory,
	}
}

type UserBalanceService struct {
	uowFactory UnitOfWorkFactory
}

func (s *UserBalanceService) AddWithdrawn(ctx context.Context, input handler.AddWithdrawnInput) error {
	validator := addWithdrawnInputValidator{Input: input}
	err := validator.Validate()
	if err != nil {
		return fmt.Errorf("invalid add withdrawn input: %w", err)
	}

	lockName := MakeUserBalanceLockName(input.UserID)
	lockNames := []string{lockName}

	return s.uowFactory.ExecuteWithUnitOfWork(ctx, lockNames, func(provider RepositoryProvider) error {
		balanceRepository := provider.UserBalanceRepository()
		withdrawnRepository := provider.UserBalanceWithdrawnRepository()

		var balance model.UserBalance
		balance, err = balanceRepository.GetByUserID(ctx, input.UserID)
		if err != nil {
			return err
		}

		if balance.Balance.LessThan(input.Amount) {
			return handler.ErrUserBalanceIsNotEnough
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
