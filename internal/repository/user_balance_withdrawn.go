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

type UserBalanceWithdrawnRepository interface {
	NextID(ctx context.Context) uuid.UUID
	Store(ctx context.Context, item model.UserBalanceWithdrawn) error
	FindByUserID(ctx context.Context, userID string) ([]model.UserBalanceWithdrawn, error)
}

func NewUserBalanceWithdrawnRepository(
	client database.ContextClient,
) UserBalanceWithdrawnRepository {
	return &userBalanceWithdrawnRepository{
		client: client,
	}
}

type userBalanceWithdrawnRepository struct {
	client database.ContextClient
}

func (r *userBalanceWithdrawnRepository) NextID(_ context.Context) uuid.UUID {
	return uuid.New()
}

func (r *userBalanceWithdrawnRepository) Store(ctx context.Context, item model.UserBalanceWithdrawn) error {
	const sqlQuery = `
		INSERT INTO "user_balance_withdrawn" (id, user_id, order_id, amount) VALUES ($1, $2, $3, $4)
	`

	_, err := r.client.Exec(ctx, sqlQuery, item.ID, item.UserID, item.OrderID, item.Amount)
	if err != nil {
		return fmt.Errorf("error on insert user_blance: %w", err)
	}

	return nil
}

func (r *userBalanceWithdrawnRepository) FindByUserID(ctx context.Context, userID string) ([]model.UserBalanceWithdrawn, error) {
	const sqlQuery = `
		SELECT id, user_id, order_id, amount, created_at
		FROM user_balance_withdrawn 
		WHERE user_id = $1 
		ORDER BY created_at 
	`

	rows, err := r.client.Query(ctx, sqlQuery, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error on query: %w", err)
	}

	items := make([]model.UserBalanceWithdrawn, 0)
	for rows.Next() {
		var item model.UserBalanceWithdrawn
		err := rows.Scan(&item.ID, &item.UserID, &item.OrderID, &item.Amount, &item.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error on scan row: %w", err)
		}
		items = append(items, item)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("error on scan rows: %w", rows.Err())
	}

	return items, nil
}
