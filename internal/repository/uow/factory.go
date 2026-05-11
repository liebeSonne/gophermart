package uow

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"

	"github.com/jackc/pgx/v5"
	"github.com/liebeSonne/gophermart/internal/service"
)

func NewUnitOfWorkFactory(
	pool Pool,
) service.UnitOfWorkFactory {
	return &unitOfWorkFactory{
		pool: pool,
	}
}

type unitOfWorkFactory struct {
	pool Pool
}

func (f *unitOfWorkFactory) ExecuteWithUnitOfWork(ctx context.Context, lockNames []string, fn func(provider service.RepositoryProvider) error) (err error) {
	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("error on begin transaction: %w", err)
	}

	defer func() {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			if !errors.Is(rollbackErr, pgx.ErrTxClosed) {
				err = errors.Join(err, fmt.Errorf("error on rollback transaction: %w", rollbackErr))
			}
		}
	}()

	p := NewRepositoryProvider(tx)

	for _, lockName := range lockNames {
		err = f.setLock(ctx, tx, lockName)
		if err != nil {
			err = fmt.Errorf("error on set lock: %w", err)
			return err
		}
	}

	err = fn(p)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		err = fmt.Errorf("error on commit transaction: %w", err)
		return err
	}

	return nil
}

func (f *unitOfWorkFactory) setLock(ctx context.Context, tx ContextClientTx, lockName string) error {
	lockID, err := f.getLockID(lockName)
	if err != nil {
		return err
	}

	const sqlQuery = `
		SELECT pg_advisory_xact_lock($1)
	`

	_, err = tx.Exec(ctx, sqlQuery, lockID)
	if err != nil {
		return err
	}

	return nil
}

func (f *unitOfWorkFactory) getLockID(name string) (int64, error) {
	h := fnv.New64a()
	_, err := h.Write([]byte(name))
	if err != nil {
		return 0, err
	}
	//nolint:gosec
	return int64(h.Sum64()), nil
}
