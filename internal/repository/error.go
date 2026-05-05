package repository

import "fmt"

type ErrConflictOrderID struct {
	OrderID string
	Err     error
}

func (e *ErrConflictOrderID) Error() string {
	return fmt.Sprintf("conflict orderID '%s': %s", e.OrderID, e.Err)
}

func (e *ErrConflictOrderID) Unwrap() error {
	return e.Err
}

func NewErrConflictOrderID(orderID string, err error) error {
	return &ErrConflictOrderID{OrderID: orderID, Err: err}
}
