package async

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestRequestProducer_Add(t *testing.T) {
	type on struct {
		channelSize         uint
		addValuesBeforeDone []string
		cancelCtx           bool
		addValuesAfterDone  []string
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
			on{10, []string{"1", "2", "3"}, true, []string{"4", "5", "6"}},
			want{[]string{"1", "2", "3"}},
		},
		{
			"ctx not Done",
			on{10, []string{"1", "2", "3"}, false, []string{"4", "5", "6"}},
			want{[]string{"1", "2", "3", "4", "5", "6"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			l, _ := test.NewNullLogger()

			p := NewRequestProducer(ctx, tc.on.channelSize, l)

			ch := p.Produce()

			for _, value := range tc.on.addValuesBeforeDone {
				p.Add(value)
			}

			if tc.on.cancelCtx {
				cancel()
				time.Sleep(time.Millisecond * 100)
			}

			for _, value := range tc.on.addValuesAfterDone {
				p.Add(value)
			}

			values := make([]string, 0)
			timeout := time.After(2 * time.Second)

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
