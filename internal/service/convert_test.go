package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/liebeSonne/gophermart/internal/model"
)

func Test_convertOrderStatus(t *testing.T) {
	invalidStatus := OrderStatus(-1)

	type on struct {
		status OrderStatus
	}
	type want struct {
		status model.OrderStatus
		err    error
	}
	testCases := []struct {
		name string
		on   on
		want want
	}{
		{
			"unknown status",
			on{invalidStatus},
			want{err: ErrUnknownAccrualOrderStatus},
		},
		{
			"registered",
			on{OrderStatusRegistered},
			want{status: model.OrderStatusNew},
		},
		{
			"processing",
			on{OrderStatusProcessing},
			want{status: model.OrderStatusProcessing},
		},
		{
			"invalid",
			on{OrderStatusInvalid},
			want{status: model.OrderStatusInvalid},
		},
		{
			"processed",
			on{OrderStatusProcessed},
			want{status: model.OrderStatusProcessed},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			status, err := ConvertOrderStatus(tc.on.status)

			if tc.want.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.want.err.Error())
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want.status, status)
		})
	}
}
