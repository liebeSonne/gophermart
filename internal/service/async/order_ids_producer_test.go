package async

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/liebeSonne/gophermart/internal/model"
	"github.com/liebeSonne/gophermart/internal/provider"
)

func TestOrderIDsProducer_Produce(t *testing.T) {
	type when struct {
		findOrderIDs    [][]string
		findOrderIDsErr error
	}
	type on struct {
		channelSize uint
		cancelCtx   bool
	}
	type want struct {
		orderIDs []string
	}
	testCases := []struct {
		name string
		on   on
		when when
		want want
	}{
		{
			"ctx Done",
			on{channelSize: 10, cancelCtx: true},
			when{findOrderIDs: [][]string{{"1", "2", "3"}}},
			want{[]string{}},
		},
		{
			"empty select",
			on{channelSize: 10},
			when{findOrderIDs: [][]string{{}}},
			want{[]string{}},
		},
		{
			"one select",
			on{channelSize: 10},
			when{findOrderIDs: [][]string{{"1", "2", "3"}}},
			want{[]string{"1", "2", "3"}},
		},
		{
			"twice select",
			on{channelSize: 10},
			when{findOrderIDs: [][]string{{"1", "2", "3"}, {"4", "5", "6"}}},
			want{[]string{"1", "2", "3", "4", "5", "6"}},
		},
		{
			"select once more then buffer",
			on{channelSize: 5},
			when{findOrderIDs: [][]string{{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}}},
			want{[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}},
		},
		{
			"select twice more then buffer",
			on{channelSize: 5},
			when{findOrderIDs: [][]string{{"1", "2", "3", "4", "5", "6", "7"}, {"8", "9", "10", "11", "12", "13", "14"}}},
			want{[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14"}},
		},
		{
			"buffer 1",
			on{channelSize: 1},
			when{findOrderIDs: [][]string{{"1", "2", "3"}}},
			want{[]string{"1", "2", "3"}},
		},
		{
			"buffer 0",
			on{channelSize: 0},
			when{findOrderIDs: [][]string{{"1", "2", "3"}}},
			want{[]string{"1", "2", "3"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			userOrderProvider := provider.NewMockUserOrderProvider(t)
			selectIndex := 0
			userOrderProvider.EXPECT().FindOrderIDs(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ model.FindUserOrderSpecification) ([]string, error) {
				if len(tc.when.findOrderIDs)-1 >= selectIndex {
					ids := tc.when.findOrderIDs[selectIndex]
					selectIndex++
					return ids, tc.when.findOrderIDsErr
				}
				return nil, tc.when.findOrderIDsErr
			}).Maybe()

			l, _ := test.NewNullLogger()
			limit := uint(100)
			limitRetriesOnError := uint(0)
			waitingOnError := time.Millisecond * 10

			p := NewOrderIDsProducer(
				ctx,
				"name",
				tc.on.channelSize,
				&limit,
				limitRetriesOnError,
				waitingOnError,
				userOrderProvider,
				l,
			)

			p.Start()
			ch := p.Produce()

			if tc.on.cancelCtx {
				cancel()
			}

			orderIDs := make([]string, 0)
			timeout := time.After(2 * time.Second)

		loop:
			for {
				select {
				case orderID, ok := <-ch:
					if !ok {
						break loop
					}
					orderIDs = append(orderIDs, orderID)
				case <-timeout:
					t.Fatal("timed out waiting for orderIDs")
				}
			}
			assert.Equal(t, tc.want.orderIDs, orderIDs)
		})
	}
}
