package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/liebeSonne/gophermart/internal/model"
)

type UserOrderRepository interface {
	NextID(ctx context.Context) uuid.UUID
	Store(ctx context.Context, items []model.UserOrder) error
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.UserOrder, error)
}

func NewUserOrderRepository(
	pool *pgxpool.Pool,
) UserOrderRepository {
	return &userOrderRepository{
		pool: pool,
	}
}

type userOrderRepository struct {
	pool *pgxpool.Pool
}

func (r *userOrderRepository) NextID(_ context.Context) uuid.UUID {
	return uuid.New()
}

func (r *userOrderRepository) Store(ctx context.Context, items []model.UserOrder) (err error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("error on begin transaction: %w", err)
	}

	defer func() {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			err = errors.Join(err, fmt.Errorf("error on rollback transaction: %w", rollbackErr))
		}
	}()

	const sqlQuery = `
		INSERT INTO user_order (id, user_id, order_id, status, accrual) VALUES ($1, $2, $3, $4, $5)
		ON DUPLICATE KEY UPDATE
			status = VALUES(status),
			accrual = VALUES(accrual),
		 	updated_at = NOW()
	`

	stmtName := "insert_user_order"
	_, err = tx.Conn().Prepare(ctx, stmtName, sqlQuery)
	if err != nil {
		return fmt.Errorf("error on prepare statement: %w", err)
	}

	insertErrors := make([]error, 0)
	for _, userOrder := range items {
		_, err = tx.Exec(ctx, stmtName, userOrder.ID, userOrder.UserID, userOrder.OrderID, userOrder.Status, userOrder.Accrual)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgerrcode.UniqueViolation == pgErr.Code {
				err = NewErrConflictOrderID(userOrder.OrderID, err)
			}
			insertErrors = append(insertErrors, err)
		}
	}

	err = errors.Join(insertErrors...)
	if err != nil {
		return fmt.Errorf("error on insert: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("error on commit transaction: %w", err)
	}

	return nil
}

func (r *userOrderRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.UserOrder, error) {
	const sqlQuery = `
		SELECT id, user_id, order_id, status, accrual, create_at, updated_at
		FROM user_order 
		WHERE user_id = $1 
	`

	rows, err := r.pool.Query(ctx, sqlQuery, userID)
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
