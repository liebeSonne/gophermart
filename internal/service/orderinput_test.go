package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUploadUserOrderInput_Validate(t *testing.T) {
	orderID1 := "123"
	userID1 := uuid.New()

	type on struct {
		input UploadUserOrderInput
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
			on{UploadUserOrderInput{orderID1, userID1}},
			want{nil},
		},
		{
			"empty order ID",
			on{UploadUserOrderInput{"", userID1}},
			want{ErrInvalidOrderID},
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
