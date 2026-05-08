package async

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/liebeSonne/gophermart/internal/model"
	"github.com/liebeSonne/gophermart/internal/provider"
)

func TestOrderIDsProducer_Produce(t *testing.T) {
	timePast1 := time.Now().Add(-10 * time.Minute)

	type when struct {
		findOrderIDToExecuteAtMaps []map[string]time.Time
		findOrderIDsErr            error
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
			when{findOrderIDToExecuteAtMaps: []map[string]time.Time{{"1": timePast1, "2": timePast1, "3": timePast1}}},
			want{[]string{}},
		},
		{
			"empty select",
			on{channelSize: 10},
			when{findOrderIDToExecuteAtMaps: []map[string]time.Time{{}}},
			want{[]string{}},
		},
		{
			"one select",
			on{channelSize: 10},
			when{findOrderIDToExecuteAtMaps: []map[string]time.Time{{"1": timePast1, "2": timePast1, "3": timePast1}}},
			want{[]string{"1", "2", "3"}},
		},
		{
			"twice select",
			on{channelSize: 10},
			when{findOrderIDToExecuteAtMaps: []map[string]time.Time{{"1": timePast1, "2": timePast1, "3": timePast1}, {"4": timePast1, "5": timePast1, "6": timePast1}}},
			want{[]string{"1", "2", "3", "4", "5", "6"}},
		},
		{
			"select once more then buffer",
			on{channelSize: 5},
			when{findOrderIDToExecuteAtMaps: []map[string]time.Time{{"1": timePast1, "2": timePast1, "3": timePast1, "4": timePast1, "5": timePast1, "6": timePast1, "7": timePast1, "8": timePast1, "9": timePast1, "10": timePast1}}},
			want{[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}},
		},
		{
			"select twice more then buffer",
			on{channelSize: 5},
			when{findOrderIDToExecuteAtMaps: []map[string]time.Time{{"1": timePast1, "2": timePast1, "3": timePast1, "4": timePast1, "5": timePast1, "6": timePast1, "7": timePast1}, {"8": timePast1, "9": timePast1, "10": timePast1, "11": timePast1, "12": timePast1, "13": timePast1, "14": timePast1}}},
			want{[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14"}},
		},
		{
			"buffer 1",
			on{channelSize: 1},
			when{findOrderIDToExecuteAtMaps: []map[string]time.Time{{"1": timePast1, "2": timePast1, "3": timePast1}}},
			want{[]string{"1", "2", "3"}},
		},
		{
			"buffer 0",
			on{channelSize: 0},
			when{findOrderIDToExecuteAtMaps: []map[string]time.Time{{"1": timePast1, "2": timePast1, "3": timePast1}}},
			want{[]string{"1", "2", "3"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			retryProducer := NewMockProducer[string](t)
			retryProducer.EXPECT().Schedule(mock.Anything, mock.Anything).Return(nil).Maybe()

			userOrderProvider := provider.NewMockUserOrderProvider(t)
			selectIndex := 0
			userOrderProvider.EXPECT().FindOrderIDToExecuteAtMap(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ model.FindUserOrderSpecification) (map[string]time.Time, error) {
				if len(tc.when.findOrderIDToExecuteAtMaps)-1 >= selectIndex {
					resultMap := tc.when.findOrderIDToExecuteAtMaps[selectIndex]
					selectIndex++
					return resultMap, tc.when.findOrderIDsErr
				}
				return nil, tc.when.findOrderIDsErr
			}).Maybe()

			l, _ := test.NewNullLogger()
			limit := uint(100)
			limitRetriesOnError := uint(0)
			waitingOnError := time.Millisecond * 10

			p := NewOrderIDsProducer(
				"name",
				tc.on.channelSize,
				&limit,
				limitRetriesOnError,
				waitingOnError,
				userOrderProvider,
				retryProducer,
				l,
			)

			p.Start(ctx)
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
			sort.Strings(tc.want.orderIDs)
			sort.Strings(orderIDs)
			assert.Equal(t, tc.want.orderIDs, orderIDs)
		})
	}
}
