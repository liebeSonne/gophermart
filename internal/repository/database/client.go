package database

import (
	"context"
	"fmt"
	"io"

	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client interface {
	Pool() *pgxpool.Pool

	io.Closer
}

func NewClient(
	ctx context.Context,
	dataSourceName string,
) (Client, error) {
	config, err := pgxpool.ParseConfig(dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.ParseConfig: %w", err)
	}

	config.AfterConnect = func(_ context.Context, conn *pgx.Conn) error {
		pgxdecimal.Register(conn.TypeMap())
		return nil
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("erro on pgxpool.New: %w", err)
	}

	return &client{
		pool: pool,
	}, nil
}

type client struct {
	pool *pgxpool.Pool
}

func (d *client) Pool() *pgxpool.Pool {
	return d.pool
}

func (d *client) Close() error {
	d.pool.Close()
	return nil
}
