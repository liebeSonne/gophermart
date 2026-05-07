package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderStatus int

const (
	OrderStatusNew OrderStatus = iota
	OrderStatusProcessing
	OrderStatusInvalid
	OrderStatusProcessed
)

type UserOrder struct {
	ID        uuid.UUID        `db:"id"`
	UserID    uuid.UUID        `db:"user_id"`
	OrderID   string           `db:"order_id"`
	Status    OrderStatus      `db:"status"`
	Accrual   *decimal.Decimal `db:"accrual"`
	CreatedAt time.Time        `db:"created_at"`
	UpdatedAt time.Time        `db:"updated_at"`
}

type FindUserOrderSpecification struct {
	Statuses        []OrderStatus
	BeforeExecuteAt time.Time
	Limit           *uint
	Offset          *uint
}
