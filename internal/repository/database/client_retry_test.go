package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/liebeSonne/gophermart/internal/repository"
)

func TestRetryContextClient_Query(t *testing.T) {
	transientError1 := &pgconn.PgError{Code: pgerrcode.CannotConnectNow}
	error1 := errors.New("error 1")

	type on struct {
		maxAttempts uint
		delay       time.Duration
	}
	type resultsData struct {
		err     error
		rowsErr error
	}
	type when struct {
		results []resultsData
	}
	type want struct {
		err   error
		times int
	}
	testCases := []struct {
		name string
		on   on
		when when
		want want
	}{
		{
			"retry error",
			on{3, time.Millisecond * 10},
			when{[]resultsData{
				{transientError1, nil},
				{transientError1, nil},
				{transientError1, nil},
			}},
			want{transientError1, 3},
		},
		{
			"not retry error",
			on{3, time.Millisecond * 10},
			when{[]resultsData{
				{transientError1, nil},
				{error1, nil},
			}},
			want{error1, 2},
		},
		{
			"rows error",
			on{3, time.Millisecond * 10},
			when{[]resultsData{
				{transientError1, nil},
				{nil, error1},
			}},
			want{error1, 2},
		},
		{
			"result from not first attempts",
			on{5, time.Millisecond * 10},
			when{[]resultsData{
				{transientError1, nil},
				{nil, nil},
			}},
			want{nil, 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			next := repository.NewMockContextClient(t)
			resultIndex := 0
			next.EXPECT().Query(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
				require.True(t, len(tc.when.results) > resultIndex)
				result := tc.when.results[resultIndex]
				resultIndex++
				if result.err != nil {
					return nil, result.err
				}
				rows := NewMockRows(t)
				rows.EXPECT().Err().Return(result.rowsErr).Maybe()
				rows.EXPECT().Close().Return().Maybe()
				return rows, nil
			}).Times(tc.want.times)

			c := NewRetryMiddlewareContextClient(next, tc.on.maxAttempts, tc.on.delay)
			r, err := c.Query(t.Context(), "", []string{})

			if tc.want.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.want.err.Error())
				return
			}

			require.NoError(t, err)
			if r != nil {
				r.Close()
			}
		})
	}
}

func TestRetryContextClient_Exec(t *testing.T) {
	transientError1 := &pgconn.PgError{Code: pgerrcode.CannotConnectNow}
	error1 := errors.New("error 1")

	type on struct {
		maxAttempts uint
		delay       time.Duration
	}
	type resultsData struct {
		err error
	}
	type when struct {
		results []resultsData
	}
	type want struct {
		err   error
		times int
	}
	testCases := []struct {
		name string
		on   on
		when when
		want want
	}{
		{
			"retry error",
			on{3, time.Millisecond * 10},
			when{[]resultsData{
				{transientError1},
				{transientError1},
				{transientError1},
			}},
			want{transientError1, 3},
		},
		{
			"not retry error",
			on{3, time.Millisecond * 10},
			when{[]resultsData{
				{transientError1},
				{error1},
			}},
			want{error1, 2},
		},
		{
			"result from not first attempts",
			on{5, time.Millisecond * 10},
			when{[]resultsData{
				{transientError1},
				{nil},
			}},
			want{nil, 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			next := repository.NewMockContextClient(t)
			resultIndex := 0
			next.EXPECT().Exec(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				require.True(t, len(tc.when.results) > resultIndex)
				result := tc.when.results[resultIndex]
				resultIndex++
				return pgconn.CommandTag{}, result.err
			}).Times(tc.want.times)

			c := NewRetryMiddlewareContextClient(next, tc.on.maxAttempts, tc.on.delay)
			_, err := c.Exec(t.Context(), "", []string{})

			if tc.want.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.want.err.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}
