package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type UserOrderStatus int

const (
	UserOrderStatusNew UserOrderStatus = iota
	UserOrderStatusProcessing
	UserOrderStatusInvalid
	UserOrderStatusProcessed
)

type UserOrder struct {
	ID        uuid.UUID        `db:"id"`
	UserID    uuid.UUID        `db:"user_id"`
	OrderID   string           `db:"order_id"`
	Status    UserOrderStatus  `db:"status"`
	Accrual   *decimal.Decimal `db:"accrual"`
	CreatedAt time.Time        `db:"created_at"`
	UpdatedAt time.Time        `db:"updated_at"`
	ExecuteAt time.Time        `db:"execute_at"`
	Retries   int              `db:"retries"`
}
