package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/liebeSonne/gophermart/internal/handler"
)

func TestAddWithdrawnInput_Validate(t *testing.T) {
	userID1 := uuid.New()
	orderID1 := "12345678903"
	invalidOrderID1 := "111"
	amountPositive1 := decimal.NewFromFloat(10.5)
	amountNegative1 := decimal.NewFromFloat(-10.5)
	amountZero := decimal.Zero

	type on struct {
		input handler.AddWithdrawnInput
	}
	type want struct {
		err error
	}
	testCases := []struct {
		name string
		on   on
		want want
	}{
		{
			"positive amount",
			on{handler.AddWithdrawnInput{UserID: userID1, OrderID: orderID1, Amount: amountPositive1}},
			want{nil},
		},
		{
			"zero amount",
			on{handler.AddWithdrawnInput{UserID: userID1, OrderID: orderID1, Amount: amountZero}},
			want{nil},
		},
		{
			"negative amount",
			on{handler.AddWithdrawnInput{UserID: userID1, OrderID: orderID1, Amount: amountNegative1}},
			want{handler.ErrInvalidWithdrawnAmount},
		},
		{
			"empty order id",
			on{handler.AddWithdrawnInput{UserID: userID1, OrderID: "", Amount: amountPositive1}},
			want{handler.ErrInvalidOrderID},
		},
		{
			"invalid order id",
			on{handler.AddWithdrawnInput{UserID: userID1, OrderID: invalidOrderID1, Amount: amountPositive1}},
			want{handler.ErrInvalidOrderID},
		},
		{
			"userID is nil",
			on{handler.AddWithdrawnInput{UserID: uuid.Nil, OrderID: orderID1, Amount: amountPositive1}},
			want{handler.ErrInvalidUserID},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := addWithdrawnInputValidator{Input: tc.on.input}
			err := validator.Validate()

			if tc.want.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.want.err.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}
