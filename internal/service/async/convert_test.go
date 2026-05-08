package async

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/liebeSonne/gophermart/internal/adapter"
	"github.com/liebeSonne/gophermart/internal/model"
)

func Test_convertOrderStatus(t *testing.T) {
	invalidStatus := adapter.OrderStatus(-1)

	type on struct {
		status adapter.OrderStatus
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
			on{adapter.OrderStatusRegistered},
			want{status: model.OrderStatusNew},
		},
		{
			"processing",
			on{adapter.OrderStatusProcessing},
			want{status: model.OrderStatusProcessing},
		},
		{
			"invalid",
			on{adapter.OrderStatusInvalid},
			want{status: model.OrderStatusInvalid},
		},
		{
			"processed",
			on{adapter.OrderStatusProcessed},
			want{status: model.OrderStatusProcessed},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			status, err := convertOrderStatus(tc.on.status)

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
