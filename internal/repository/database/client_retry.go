package database

import (
	"context"
	"errors"
	"time"

	"github.com/avast/retry-go"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/liebeSonne/gophermart/internal/repository"
)

func NewRetryMiddlewareContextClient(
	next repository.ContextClient,
	maxAttempts uint,
	delay time.Duration,
) repository.ContextClient {
	return &retryContextClient{
		next:        next,
		maxAttempts: maxAttempts,
		delay:       delay,
	}
}

type retryContextClient struct {
	next        repository.ContextClient
	maxAttempts uint
	delay       time.Duration
}

func (r *retryContextClient) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	var resultRows pgx.Rows

	err := retry.Do(
		func() error {
			rows, err := r.next.Query(ctx, sql, args...)
			if err != nil {
				return err
			}

			success := false
			defer func() {
				if !success {
					rows.Close()
				}
			}()

			if rows.Err() != nil {
				return rows.Err()
			}

			resultRows = rows
			success = true
			return nil
		},
		retry.Attempts(r.maxAttempts),
		retry.Delay(r.delay),
		retry.RetryIf(r.retryIf),
	)
	if err != nil {
		return nil, err
	}

	return resultRows, nil
}

func (r *retryContextClient) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return r.next.QueryRow(ctx, sql, args...)
}

func (r *retryContextClient) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	var resultCommandTag pgconn.CommandTag

	err := retry.Do(
		func() error {
			commandTag, err := r.next.Exec(ctx, sql, arguments...)
			if err != nil {
				return err
			}
			resultCommandTag = commandTag
			return nil
		},
		retry.Attempts(r.maxAttempts),
		retry.Delay(r.delay),
		retry.RetryIf(r.retryIf),
	)
	if err != nil {
		return pgconn.CommandTag{}, err
	}

	return resultCommandTag, nil
}

func (r *retryContextClient) retryIf(err error) bool {
	return r.isTransientError(err)
}

func (r *retryContextClient) isTransientError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.SerializationFailure,
			pgerrcode.DeadlockDetected,
			pgerrcode.AdminShutdown,
			pgerrcode.CannotConnectNow:
			return true
		}
	}
	return false
}
