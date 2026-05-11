package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/liebeSonne/gophermart/internal/repository/uow"
)

func NewPoolRetryMiddleware(
	next uow.Pool,
	maxAttempts uint,
	delay time.Duration,
) uow.Pool {
	return &poolRetryMiddleware{
		next:        next,
		maxAttempts: maxAttempts,
		delay:       delay,
	}
}

type poolRetryMiddleware struct {
	next        uow.Pool
	maxAttempts uint
	delay       time.Duration
}

func (r *poolRetryMiddleware) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (uow.ContextClientTx, error) {
	tx, err := r.next.BeginTx(ctx, txOptions)
	if err != nil {
		return nil, err
	}
	return NewRetryMiddlewareContextClientTx(tx, r.maxAttempts, r.delay), nil
}
