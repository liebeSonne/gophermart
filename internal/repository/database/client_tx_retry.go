package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/liebeSonne/gophermart/internal/repository"
	"github.com/liebeSonne/gophermart/internal/repository/uow"
)

func NewRetryMiddlewareContextClientTx(
	next uow.ContextClientTx,
	maxAttempts uint,
	delay time.Duration,
) uow.ContextClientTx {
	return &retryContextClientTx{
		next:   next,
		client: NewRetryMiddlewareContextClient(next, maxAttempts, delay),
	}
}

type retryContextClientTx struct {
	client repository.ContextClient
	next   uow.ContextClientTx
}

func (r *retryContextClientTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return r.client.Query(ctx, sql, args...)
}

func (r *retryContextClientTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return r.client.QueryRow(ctx, sql, args...)
}

func (r *retryContextClientTx) Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error) {
	return r.client.Exec(ctx, sql, arguments...)
}

func (r *retryContextClientTx) Commit(ctx context.Context) error {
	return r.next.Commit(ctx)
}

func (r *retryContextClientTx) Rollback(ctx context.Context) error {
	return r.next.Rollback(ctx)
}
