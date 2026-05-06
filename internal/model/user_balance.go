package model

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type UserBalance struct {
	UserID       uuid.UUID       `db:"user_id"`
	Balance      decimal.Decimal `db:"balance"`
	WithdrawnSum decimal.Decimal `db:"withdrawn_sum"`
}
