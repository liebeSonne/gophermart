package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// nolint: goconst
func TestCreateUserInput_Validate(t *testing.T) {
	login1 := "login 1"
	password1 := "password 1"

	type on struct {
		input CreateUserInput
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
			on{CreateUserInput{login1, password1}},
			want{nil},
		},
		{
			"empty login",
			on{CreateUserInput{"", password1}},
			want{ErrInvalidUserLogin},
		},
		{
			"empty password",
			on{CreateUserInput{login1, ""}},
			want{ErrInvalidUserPassword},
		},
		{
			"invalid login length",
			on{CreateUserInput{strings.Repeat("l", MaxUserLoginLength+1), password1}},
			want{ErrInvalidUserLogin},
		},
		{
			"invalid password length",
			on{CreateUserInput{login1, strings.Repeat("l", MaxUserPasswordLength+1)}},
			want{ErrInvalidUserPassword},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.on.input.Validate()

			if tc.want.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.want.err.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}

// nolint: goconst
func TestLoginUserInput_Validate(t *testing.T) {
	login1 := "login 1"
	password1 := "password 1"

	type on struct {
		input LoginUserInput
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
			on{LoginUserInput{login1, password1}},
			want{nil},
		},
		{
			"empty login",
			on{LoginUserInput{"", password1}},
			want{ErrInvalidUserLogin},
		},
		{
			"empty password",
			on{LoginUserInput{login1, ""}},
			want{ErrInvalidUserPassword},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.on.input.Validate()

			if tc.want.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.want.err.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}
