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

func TestProducer_Add(t *testing.T) {
	type on struct {
		channelSize            uint
		addValuesBeforeCtxDone []string
		cancelCtx              bool
		addValuesAfterCtxDone  []string
		waiting                time.Duration
	}
	type want struct {
		values []string
	}
	testCases := []struct {
		name string
		on   on
		want want
	}{
		{
			"ctx Done",
			on{10, []string{"1", "2", "3"}, true, []string{"4", "5", "6"}, time.Millisecond * 200},
			want{[]string{"1", "2", "3"}},
		},
		{
			"ctx not Done",
			on{10, []string{"1", "2", "3"}, false, []string{"4", "5", "6"}, time.Millisecond * 200},
			want{[]string{"1", "2", "3", "4", "5", "6"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			l, _ := test.NewNullLogger()

			p := NewProducer[string](ctx, "name", tc.on.channelSize, l)

			ch := p.Produce()

			for _, value := range tc.on.addValuesBeforeCtxDone {
				p.Add(value)
			}

			if tc.on.cancelCtx {
				cancel()
				time.Sleep(time.Millisecond * 10)
			}

			for _, value := range tc.on.addValuesAfterCtxDone {
				p.Add(value)
			}

			values := make([]string, 0)
			timeout := time.After(tc.on.waiting)

		loop:
			for {
				select {
				case value, ok := <-ch:
					if !ok {
						break loop
					}
					values = append(values, value)
				case <-timeout:
					break loop
				}
			}

			assert.Equal(t, tc.want.values, values)
		})
	}
}

//nolint:gocognit
func TestProducer_Schedule(t *testing.T) {
	type ctxDone struct {
		beforeCtxDone time.Duration
		cancelCtx     bool
		afterCtxDone  time.Duration
	}
	type scheduleData struct {
		value  string
		delay  time.Duration
		cancel bool
	}
	type on struct {
		channelSize      uint
		addBeforeCtxDone []scheduleData
		ctxDone          ctxDone
		addAfterCtxDone  []scheduleData
		waiting          time.Duration
	}
	type want struct {
		values []string
	}
	testCases := []struct {
		name string
		on   on
		want want
	}{
		{
			"ctx Done",
			on{
				10,
				[]scheduleData{
					{"1", time.Millisecond * 100, false},
					{"2", time.Millisecond * 200, false},
					{"3", time.Millisecond * 300, false},
				},
				ctxDone{
					time.Millisecond * 150,
					true,
					time.Millisecond * 100,
				},
				[]scheduleData{
					{"4", time.Millisecond * 110, false},
					{"5", time.Millisecond * 210, false},
					{"6", time.Millisecond * 310, false},
					{"7", time.Second * 2, false},
				},
				time.Second * 1,
			},
			want{[]string{"1"}},
		},
		{
			"ctx not Done",
			on{
				10,
				[]scheduleData{
					{"1", time.Millisecond * 100, false},
					{"2", time.Millisecond * 200, false},
					{"3", time.Millisecond * 300, false},
				},
				ctxDone{},
				[]scheduleData{
					{"4", time.Millisecond * 110, false},
					{"5", time.Millisecond * 210, false},
					{"6", time.Millisecond * 310, false},
					{"7", time.Second * 2, false},
				},
				time.Second * 1,
			},
			want{[]string{"1", "4", "2", "5", "3", "6"}},
		},
		{
			"cancel schedule",
			on{
				10,
				[]scheduleData{
					{"1", time.Millisecond * 100, true},
					{"2", time.Millisecond * 200, false},
					{"3", time.Millisecond * 300, true},
				},
				ctxDone{},
				[]scheduleData{
					{"4", time.Millisecond * 110, false},
					{"5", time.Millisecond * 210, true},
					{"6", time.Millisecond * 310, false},
					{"7", time.Second * 2, false},
					{"8", time.Second * 2, true},
				},
				time.Second * 1,
			},
			want{[]string{"4", "2", "6"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			l, _ := test.NewNullLogger()

			p := NewProducer[string](ctx, "name", tc.on.channelSize, l)
			ch := p.Produce()

			for _, schedule := range tc.on.addBeforeCtxDone {
				scheduleCancel := p.Schedule(schedule.value, schedule.delay)
				if schedule.cancel {
					scheduleCancel()
				}
			}

			if tc.on.ctxDone.cancelCtx {
				time.Sleep(tc.on.ctxDone.beforeCtxDone)
				cancel()
				time.Sleep(tc.on.ctxDone.afterCtxDone)
			}

			for _, schedule := range tc.on.addAfterCtxDone {
				scheduleCancel := p.Schedule(schedule.value, schedule.delay)
				if schedule.cancel {
					scheduleCancel()
				}
			}

			values := make([]string, 0)
			timeout := time.After(tc.on.waiting)

		loop:
			for {
				select {
				case value, ok := <-ch:
					if !ok {
						break loop
					}
					values = append(values, value)
				case <-timeout:
					break loop
				}
			}

			assert.Equal(t, tc.want.values, values)
		})
	}
}

func TestProducer_Setup(t *testing.T) {
	type on struct {
		channelSize uint
		values      []string
		cancelCtx   bool
		cancelSetup bool
		waiting     time.Duration
		waitingCh   time.Duration
	}
	type want struct {
		values []string
	}
	testCases := []struct {
		name string
		on   on
		want want
	}{
		{
			"ctx Done",
			on{
				10,
				[]string{"1", "2", "3"},
				true,
				false,
				time.Millisecond * 300,
				time.Millisecond * 0,
			},
			want{[]string{}},
		},
		{
			"cancel setup",
			on{
				10,
				[]string{"1", "2", "3"},
				false,
				true,
				time.Millisecond * 300,
				time.Millisecond * 10,
			},
			want{[]string{}},
		},
		{
			"ctx not Done",
			on{
				10,
				[]string{"1", "2", "3"},
				false,
				false,
				time.Millisecond * 300,
				time.Millisecond * 0,
			},
			want{[]string{"1", "2", "3"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			l, _ := test.NewNullLogger()

			p := NewProducer[string](ctx, "name", uint(len(tc.on.values)), l)

			setupCh := testGenerateChWaiting(ctx, tc.on.values, tc.on.waitingCh)

			if tc.on.cancelCtx {
				cancel()
			}
			cancelSetup := p.Setup(setupCh)
			if tc.on.cancelSetup {
				cancelSetup()
				time.Sleep(time.Millisecond * 10)
			}

			ch := p.Produce()

			values := make([]string, 0)
			timeout := time.After(tc.on.waiting)

		loop:
			for {
				select {
				case value, ok := <-ch:
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
			sort.Strings(values)
			sort.Strings(tc.want.values)
			assert.Equal(t, tc.want.values, values)
		})
	}
}
