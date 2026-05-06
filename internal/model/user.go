package model

import (
	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID `db:"id"`
	Login    string    `db:"login"`
	PassHash string    `db:"passhash"`
}
