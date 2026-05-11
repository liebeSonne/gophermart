package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/liebeSonne/gophermart/internal/model"
)

func NewUserBalanceRepository(
	client ContextClient,
) *UserBalanceRepository {
	return &UserBalanceRepository{
		client: client,
	}
}

type UserBalanceRepository struct {
	client ContextClient
}

func (r *UserBalanceRepository) Store(ctx context.Context, item model.UserBalance) error {
	const sqlQuery = `
		INSERT INTO "user_balance" (user_id, balance, withdrawn_sum) VALUES ($1, $2, $3)
		ON CONFLICT (user_id)
		DO UPDATE SET 
			balance = EXCLUDED.balance,
			withdrawn_sum = EXCLUDED.withdrawn_sum,
		 	updated_at = NOW()
	`

	_, err := r.client.Exec(ctx, sqlQuery, item.UserID, item.Balance, item.WithdrawnSum)
	if err != nil {
		return fmt.Errorf("error on insert user_blance: %w", err)
	}

	return nil
}

func (r *UserBalanceRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*model.UserBalance, error) {
	const sqlQuery = `
		SELECT user_id, balance, withdrawn_sum 
		FROM "user_balance"
		WHERE user_id = $1 
		LIMIT 1
	`

	var user model.UserBalance
	row := r.client.QueryRow(ctx, sqlQuery, userID)
	err := row.Scan(&user.UserID, &user.Balance, &user.WithdrawnSum)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error on scan row: %w", err)
	}

	return &user, nil
}

func (r *UserBalanceRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (model.UserBalance, error) {
	userBalancePtr, err := r.FindByUserID(ctx, userID)
	if err != nil {
		return model.UserBalance{}, err
	}

	if userBalancePtr == nil {
		return model.UserBalance{
			UserID:       userID,
			Balance:      decimal.Zero,
			WithdrawnSum: decimal.Zero,
		}, nil
	}

	return *userBalancePtr, nil
}
