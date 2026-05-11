package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/liebeSonne/gophermart/internal/handler"
)

func TestUploadUserOrderInput_Validate(t *testing.T) {
	orderID1 := "12345678903"
	invalidOrderID1 := "111"
	notNumberOrderID1 := "abc"
	userID1 := uuid.New()

	type on struct {
		input handler.UploadUserOrderInput
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
			"valid",
			on{handler.UploadUserOrderInput{OrderID: orderID1, UserID: userID1}},
			want{nil},
		},
		{
			"empty order ID",
			on{handler.UploadUserOrderInput{OrderID: "", UserID: userID1}},
			want{handler.ErrInvalidOrderID},
		},
		{
			"invalid order ID",
			on{handler.UploadUserOrderInput{OrderID: invalidOrderID1, UserID: userID1}},
			want{handler.ErrInvalidOrderID},
		},
		{
			"not number order ID",
			on{handler.UploadUserOrderInput{OrderID: notNumberOrderID1, UserID: userID1}},
			want{handler.ErrInvalidOrderID},
		},
		{
			"userID is nil",
			on{handler.UploadUserOrderInput{OrderID: orderID1, UserID: uuid.Nil}},
			want{handler.ErrInvalidUserID},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := uploadUserOrderInputValidator{Input: tc.on.input}
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
