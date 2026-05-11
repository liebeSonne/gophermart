package uow

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Pool interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (ContextClientTx, error)
}
