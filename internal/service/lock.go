package service

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	usersLockName       = "user"
	userOrderLockName   = "user_order_%s"
	userBalanceLockName = "user_balance_%s"
)

func MakeUsersLockName() string {
	return usersLockName
}

func MakeUserOrderLockName(orderID string) string {
	return fmt.Sprintf(userOrderLockName, orderID)
}

func MakeUserBalanceLockName(userID uuid.UUID) string {
	return fmt.Sprintf(userBalanceLockName, userID.String())
}
