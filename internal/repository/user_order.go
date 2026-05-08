package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/liebeSonne/gophermart/internal/model"
	"github.com/liebeSonne/gophermart/internal/repository/database"
)

type UserOrderRepository interface {
	NextID(ctx context.Context) uuid.UUID
	Store(ctx context.Context, items []model.UserOrder) error
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.UserOrder, error)
	FindByOrderID(ctx context.Context, orderID string) (*model.UserOrder, error)
	FindOrderIDToExecuteAtMap(ctx context.Context, spec model.FindUserOrderSpecification) (map[string]time.Time, error)
	FindUserIDByOrderID(ctx context.Context, orderID string) (*uuid.UUID, error)
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
		INSERT INTO user_order (id, user_id, order_id, status, accrual, created_at, updated_at, execute_at, retries) VALUES %s
		ON CONFLICT (id)
		DO UPDATE SET 
			status = EXCLUDED.status,
			accrual = EXCLUDED.accrual,
		 	updated_at = EXCLUDED.updated_at,
		 	execute_at = EXCLUDED.execute_at,
		 	retries = EXCLUDED.retries
	`

	for chunkItems := range slices.Chunk(items, chunkSize) {
		values := make([]string, 0, len(chunkItems))
		args := make([]any, 0, len(chunkItems)*9)

		for i, item := range chunkItems {
			base := i * 9
			params := fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)", base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9)
			values = append(values, params)
			args = append(args, item.ID, item.UserID, item.OrderID, item.Status, item.Accrual, item.CreatedAt, item.UpdatedAt, item.ExecuteAt, item.Retries)
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
		SELECT id, user_id, order_id, status, accrual, created_at, updated_at, execute_at, retries
		FROM user_order 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
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
		err := rows.Scan(&item.ID, &item.UserID, &item.OrderID, &orderStatus, &item.Accrual, &item.CreatedAt, &item.UpdatedAt, &item.ExecuteAt, &item.Retries)
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
		SELECT id, user_id, order_id, status, accrual, created_at, updated_at, execute_at, retries
		FROM user_order 
		WHERE order_id = $1 
		LIMIT 1
	`

	var item model.UserOrder
	var orderStatus int

	row := r.client.QueryRow(ctx, sqlQuery, orderID)
	err := row.Scan(&item.ID, &item.UserID, &item.OrderID, &orderStatus, &item.Accrual, &item.CreatedAt, &item.UpdatedAt, &item.ExecuteAt, &item.Retries)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error on scan row: %w", err)
	}

	item.Status = model.OrderStatus(orderStatus)

	return &item, nil
}

func (r *userOrderRepository) FindOrderIDToExecuteAtMap(ctx context.Context, spec model.FindUserOrderSpecification) (map[string]time.Time, error) {
	const sqlQuery = `
		SELECT order_id, execute_at
		FROM user_order 
		WHERE 
		    updated_at <= $1
			AND status = ANY($2) 
		ORDER BY execute_at
		LIMIT $3 OFFSET $4
	`

	intStatuses := make([]int, 0, len(spec.Statuses))
	for _, status := range spec.Statuses {
		intStatuses = append(intStatuses, int(status))
	}

	rows, err := r.client.Query(ctx, sqlQuery, spec.BeforeUpdatedAt, intStatuses, spec.Limit, spec.Offset)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error on query: %w", err)
	}

	resultMap := make(map[string]time.Time)
	for rows.Next() {
		var item OrderDataShort
		err = rows.Scan(&item.OrderID, &item.ExecuteAt)
		if err != nil {
			return nil, fmt.Errorf("error on scan row: %w", err)
		}
		resultMap[item.OrderID] = item.ExecuteAt
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("error on scan rows: %w", rows.Err())
	}

	return resultMap, nil
}

func (r *userOrderRepository) FindUserIDByOrderID(ctx context.Context, orderID string) (*uuid.UUID, error) {
	const sqlQuery = `
		SELECT user_id
		FROM user_order 
		WHERE order_id = $1 
		LIMIT 1
	`

	var userID uuid.UUID
	row := r.client.QueryRow(ctx, sqlQuery, orderID)
	err := row.Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error on scan row: %w", err)
	}

	return &userID, nil
}

type OrderDataShort struct {
	OrderID   string    `db:"order_id"`
	ExecuteAt time.Time `db:"execute_at"`
}
