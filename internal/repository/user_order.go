package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/liebeSonne/gophermart/internal/model"
	"github.com/liebeSonne/gophermart/internal/repository/database"
)

type UserOrderRepository interface {
	NextID(ctx context.Context) uuid.UUID
	Store(ctx context.Context, items []model.UserOrder) error
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.UserOrder, error)
	FindByOrderID(ctx context.Context, orderID string) (*model.UserOrder, error)
}

func NewUserOrderRepository(
	client database.ContextClient,
) UserOrderRepository {
	return &userOrderRepository{
		client: client,
	}
}

type userOrderRepository struct {
	client database.ContextClient
}

func (r *userOrderRepository) NextID(_ context.Context) uuid.UUID {
	return uuid.New()
}

func (r *userOrderRepository) Store(ctx context.Context, items []model.UserOrder) error {
	const sqlQuery = `
		INSERT INTO user_order (id, user_id, order_id, status, accrual) VALUES %s
		ON CONFLICT (id)
		DO UPDATE SET 
			status = EXCLUDED.status,
			accrual = EXCLUDED.accrual,
		 	updated_at = NOW()
	`

	for chunkItems := range slices.Chunk(items, chunkSize) {
		values := make([]string, 0, len(chunkItems))
		args := make([]any, 0, len(chunkItems)*5)

		for i, item := range chunkItems {
			base := i * 5
			params := fmt.Sprintf("($%d,$%d,$%d,$%d,$%d)", base+1, base+2, base+3, base+4, base+5)
			values = append(values, params)
			args = append(args, item.ID, item.UserID, item.OrderID, item.Status, item.Accrual)
		}

		query := fmt.Sprintf(sqlQuery, strings.Join(values, ","))
		_, err := r.client.Exec(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("error on executing statement: %w", err)
		}
	}

	return nil
}

func (r *userOrderRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.UserOrder, error) {
	const sqlQuery = `
		SELECT id, user_id, order_id, status, accrual, created_at, updated_at
		FROM user_order 
		WHERE user_id = $1 
	`

	rows, err := r.client.Query(ctx, sqlQuery, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error on query: %w", err)
	}

	items := make([]model.UserOrder, 0)
	for rows.Next() {
		var item model.UserOrder
		var orderStatus int
		err := rows.Scan(&item.ID, &item.UserID, &item.OrderID, &orderStatus, &item.Accrual, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("error on scan row: %w", err)
		}
		item.Status = model.OrderStatus(orderStatus)
		items = append(items, item)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("error on scan rows: %w", rows.Err())
	}

	return items, nil
}

func (r *userOrderRepository) FindByOrderID(ctx context.Context, orderID string) (*model.UserOrder, error) {
	const sqlQuery = `
		SELECT id, user_id, order_id, status, accrual, created_at, updated_at
		FROM user_order 
		WHERE order_id = $1 
		LIMIT 1
	`

	var item model.UserOrder
	var orderStatus int

	row := r.client.QueryRow(ctx, sqlQuery, orderID)
	err := row.Scan(&item.ID, &item.UserID, &item.OrderID, &orderStatus, &item.Accrual, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error on scan row: %w", err)
	}

	item.Status = model.OrderStatus(orderStatus)

	return &item, nil
}
