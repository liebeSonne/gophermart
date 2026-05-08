package service

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

var ErrInvalidOrderID = errors.New("invalid order ID")
var ErrInvalidUserID = errors.New("invalid user id")

type UploadUserOrderInput struct {
	OrderID string
	UserID  uuid.UUID
}

func (i *UploadUserOrderInput) Validate() error {
	if i.UserID == uuid.Nil {
		return ErrInvalidUserID
	}

	err := validateOrderID(i.OrderID)
	if err != nil {
		return err
	}

	return nil
}

func validateOrderID(orderID string) error {
	if orderID == "" {
		return ErrInvalidOrderID
	}

	isValid, err := validLuhn(orderID)
	if err != nil {
		return err
	}
	if !isValid {
		return fmt.Errorf("invalid lumn order ID (%v): %w", orderID, ErrInvalidOrderID)
	}

	return nil
}

func validLuhn(n string) (bool, error) {
	number, err := strconv.Atoi(n)
	if err != nil {
		return false, fmt.Errorf("failed conver number (%v) to int: %w", n, err)
	}
	return (number%10+checksum(number/10))%10 == 0, nil
}

func checksum(number int) int {
	var sum int

	for i := 0; number > 0; i++ {
		cur := number % 10
		if i%2 == 0 { // even
			cur *= 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}
		sum += cur
		number /= 10
	}

	return sum % 10
}
