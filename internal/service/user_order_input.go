package service

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/google/uuid"

	"github.com/liebeSonne/gophermart/internal/handler"
)

type uploadUserOrderInputValidator struct {
	Input handler.UploadUserOrderInput
}

func (i *uploadUserOrderInputValidator) Validate() error {
	if i.Input.UserID == uuid.Nil {
		return handler.ErrInvalidUserID
	}

	err := validateOrderID(i.Input.OrderID)
	if err != nil {
		return err
	}

	return nil
}

func validateOrderID(orderID string) error {
	if orderID == "" {
		return handler.ErrInvalidOrderID
	}

	isValid, err := validLuhn(orderID)
	if err != nil {
		return errors.Join(handler.ErrInvalidOrderID, fmt.Errorf("error validating luhn: %w", err))
	}
	if !isValid {
		return fmt.Errorf("invalid lumn order ID (%v): %w", orderID, handler.ErrInvalidOrderID)
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
