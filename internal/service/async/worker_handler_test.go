package async

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerHandler_Handle(t *testing.T) {
	defaultResult := -1

	type when struct {
		results map[string]int
	}
	type on struct {
		values       []string
		countWorkers uint
		cancelCtx    bool
		waiting      time.Duration
	}
	type want struct {
		values []int
	}
	testCases := []struct {
		name string
		on   on
		when when
		want want
	}{
		{
			"ctx Done",
			on{[]string{"1", "2", "3"}, 2, true, time.Millisecond * 200},
			when{map[string]int{"1": 11, "2": 22, "3": 33}},
			want{[]int{}},
		},
		{
			"one workers",
			on{[]string{"1", "2", "3"}, 1, false, time.Millisecond * 200},
			when{map[string]int{"1": 11, "2": 22, "3": 33}},
			want{[]int{11, 22, 33}},
		},
		{
			"two workers",
			on{[]string{"1", "2", "3"}, 2, false, time.Millisecond * 200},
			when{map[string]int{"1": 11, "2": 22, "3": 33}},
			want{[]int{11, 22, 33}},
		},
		{
			"zero workers",
			on{[]string{"1", "2", "3"}, 0, false, time.Millisecond * 200},
			when{map[string]int{"1": 11, "2": 22, "3": 33}},
			want{[]int{}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			l, _ := test.NewNullLogger()

			handler := WorkerHandleFunc[string, int](func(value string, resulCh chan<- int) {
				result, ok := tc.when.results[value]
				if !ok {
					result = defaultResult
				}
				resulCh <- result
			})

			wh := NewWorkerHandler[string, int](ctx, "name", handler, l)

			inCh := testGenerateChWaiting[string](ctx, tc.on.values, 0)

			outCh := make(chan int)

			if tc.on.cancelCtx {
				cancel()
				time.Sleep(time.Millisecond * 10)
			}

			wh.Handle(inCh, outCh, tc.on.countWorkers)

			values := make([]int, 0)
			timeout := time.After(tc.on.waiting)

		loop:
			for {
				select {
				case value, ok := <-outCh:
					if !ok {
						break loop
					}
					values = append(values, value)
				case <-ctx.Done():
					break loop
				case <-timeout:
					break loop
				}
			}

			require.Len(t, values, len(tc.want.values))
			sort.Ints(values)
			sort.Ints(tc.want.values)
			assert.Equal(t, tc.want.values, values)
		})
	}
}
