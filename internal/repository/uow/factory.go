package uow

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UnitOfWorkFactory interface {
	ExecuteWithUnitOfWork(ctx context.Context, f func(provider RepositoryProvider) error) error
}

func NewUnitOfWorkFactory(
	pool *pgxpool.Pool,
) UnitOfWorkFactory {
	return &unitOfWorkFactory{
		pool: pool,
	}
}

type unitOfWorkFactory struct {
	pool *pgxpool.Pool
}

func (f *unitOfWorkFactory) ExecuteWithUnitOfWork(ctx context.Context, fn func(provider RepositoryProvider) error) error {
	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("error on begin transaction: %w", err)
	}

	p := NewRepositoryProvider(tx)

	err = fn(p)

	if err != nil {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			err = errors.Join(err, fmt.Errorf("error on rollback transaction: %w", rollbackErr))
		}
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("error on commit transaction: %w", err)
	}

	return nil
}
