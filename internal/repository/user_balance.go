package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/liebeSonne/gophermart/internal/model"
	"github.com/liebeSonne/gophermart/internal/repository/database"
)

type UserBalanceRepository interface {
	Store(ctx context.Context, item model.UserBalance) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*model.UserBalance, error)
}

func NewUserBalanceRepository(
	client database.ContextClient,
) UserBalanceRepository {
	return &userBalanceRepository{
		client: client,
	}
}

type userBalanceRepository struct {
	client database.ContextClient
}

func (r *userBalanceRepository) Store(ctx context.Context, item model.UserBalance) error {
	const sqlQuery = `
		INSERT INTO "user_balance" (user_id, withdrawn_sum, withdrawn_sum) VALUES ($1, $2, $3)
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

func (r *userBalanceRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*model.UserBalance, error) {
	const sqlQuery = `
		SELECT user_id, withdrawn_sum, withdrawn_sum 
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
