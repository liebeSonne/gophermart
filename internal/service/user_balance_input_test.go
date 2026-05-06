package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestAddWithdrawnInput_Validate(t *testing.T) {
	userID1 := uuid.New()
	orderID1 := "123"
	amountPositive1 := decimal.NewFromFloat(10.5)
	amountNegative1 := decimal.NewFromFloat(-10.5)
	amountZero := decimal.Zero

	type on struct {
		input AddWithdrawnInput
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
			on{AddWithdrawnInput{userID1, orderID1, amountPositive1}},
			want{nil},
		},
		{
			"zero amount",
			on{AddWithdrawnInput{userID1, orderID1, amountZero}},
			want{nil},
		},
		{
			"negative amount",
			on{AddWithdrawnInput{userID1, orderID1, amountNegative1}},
			want{ErrInvalidWithdrawnAmount},
		},
		{
			"empty order id",
			on{AddWithdrawnInput{userID1, "", amountPositive1}},
			want{ErrInvalidOrderID},
		},
		{
			"userID is nil",
			on{AddWithdrawnInput{uuid.Nil, orderID1, amountPositive1}},
			want{ErrInvalidUserID},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.on.input.Validate()

			if tc.want.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.want.err.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}
