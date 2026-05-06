package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type UserBalanceWithdrawn struct {
	ID        uuid.UUID       `db:"id"`
	UserID    uuid.UUID       `db:"user_id"`
	OrderID   string          `db:"order_id"`
	Amount    decimal.Decimal `db:"amount"`
	CreatedAt time.Time       `db:"created_at"`
}
