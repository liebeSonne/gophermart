package async

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFanIn(t *testing.T) {
	type on struct {
		inputs    [][]string
		cancelCtx bool
		waiting   time.Duration
	}
	type want struct {
		outputs []string
	}
	testCases := []struct {
		name string
		on   on
		want want
	}{
		{
			"ctx Done",
			on{[][]string{{"1", "2", "3"}, {"4", "5", "6"}}, true, time.Millisecond * 500},
			want{[]string{}},
		},
		{
			"ctx not Done",
			on{[][]string{{"1", "2", "3"}, {"4", "5", "6"}}, false, time.Millisecond * 500},
			want{[]string{"1", "2", "3", "4", "5", "6"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			if tc.on.cancelCtx {
				cancel()
			}

			inputsCh := make([]<-chan string, 0, len(tc.on.inputs))
			for _, inputs := range tc.on.inputs {
				ch := testGenerateChWaiting(ctx, inputs, 0)
				inputsCh = append(inputsCh, ch)
			}

			outputCh := FanIn[string](ctx, inputsCh...)

			values := make([]string, 0)
			timeout := time.After(tc.on.waiting)

		loop:
			for {
				select {
				case value, ok := <-outputCh:
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

			require.Len(t, values, len(tc.want.outputs))
			sort.Strings(values)
			sort.Strings(tc.want.outputs)
			assert.Equal(t, tc.want.outputs, values)
		})
	}
}

func testGenerateChWaiting[T any](ctx context.Context, items []T, waiting time.Duration) <-chan T {
	outCh := make(chan T, len(items))
	go func() {
		defer close(outCh)
		for _, value := range items {
			time.Sleep(waiting)
			select {
			case <-ctx.Done():
				return
			case outCh <- value:
			}
		}
	}()
	return outCh
}
