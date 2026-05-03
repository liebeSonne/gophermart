//nolint:dupl
package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogLevel_String(t *testing.T) {
	type on struct {
		level LogLevel
	}
	type want struct {
		str string
	}
	testCases := []struct {
		name string
		on   on
		want want
	}{
		{
			"string",
			on{"some string"},
			want{"some string"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			str := tc.on.level.String()
			require.Equal(t, tc.want.str, str)
		})
	}
}

func TestLogLevel_Validate(t *testing.T) {
	type on struct {
		level string
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
			on{string(LogFormatText)},
			want{nil},
		},
		{
			"not valid",
			on{"invalid"},
			want{ErrInvalidLogFormat},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ll := LogFormat(tc.on.level)
			err := ll.Validate()

			if tc.want.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.want.err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestLogFormat_Set(t *testing.T) {
	type on struct {
		format string
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
			on{string(LogFormatText)},
			want{nil},
		},
		{
			"not valid",
			on{"invalid"},
			want{ErrInvalidLogFormat},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lf := LogFormatJSON
			err := lf.Set(tc.on.format)

			if tc.want.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.want.err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestLogFormat_SetValue(t *testing.T) {
	type on struct {
		format string
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
			on{string(LogFormatText)},
			want{nil},
		},
		{
			"not valid",
			on{"invalid"},
			want{ErrInvalidLogFormat},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lf := LogFormatJSON
			err := lf.SetValue(tc.on.format)

			if tc.want.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.want.err)
				return
			}

			require.NoError(t, err)
		})
	}
}
