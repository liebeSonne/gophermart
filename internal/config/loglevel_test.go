//nolint:dupl
package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogFormat_String(t *testing.T) {
	type on struct {
		format LogFormat
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
			str := tc.on.format.String()
			require.Equal(t, tc.want.str, str)
		})
	}
}

func TestLogFormat_Validate(t *testing.T) {
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
			on{string(LogLevelError)},
			want{nil},
		},
		{
			"not valid",
			on{"invalid"},
			want{ErrInvalidLogLevel},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ll := LogLevel(tc.on.level)
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

func TestLogLevel_Set(t *testing.T) {
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
			on{string(LogLevelError)},
			want{nil},
		},
		{
			"not valid",
			on{"invalid"},
			want{ErrInvalidLogLevel},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ll := LogLevelInfo
			err := ll.Set(tc.on.level)

			if tc.want.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.want.err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestLogLevel_SetValue(t *testing.T) {
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
			on{string(LogLevelError)},
			want{nil},
		},
		{
			"not valid",
			on{"invalid"},
			want{ErrInvalidLogLevel},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ll := LogLevelInfo
			err := ll.SetValue(tc.on.level)

			if tc.want.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.want.err)
				return
			}

			require.NoError(t, err)
		})
	}
}
