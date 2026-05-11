package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/liebeSonne/gophermart/internal/repository/uow"
)

func NewPool(
	pool *pgxpool.Pool,
) uow.Pool {
	return &poolImpl{
		pool: pool,
	}
}

type poolImpl struct {
	pool *pgxpool.Pool
}

func (p *poolImpl) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (uow.ContextClientTx, error) {
	tx, err := p.pool.BeginTx(ctx, txOptions)
	if err != nil {
		return nil, err
	}

	return tx, nil
}
