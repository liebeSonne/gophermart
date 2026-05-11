package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/liebeSonne/gophermart/internal/handler"
)

// nolint: goconst
func TestCreateUserInput_Validate(t *testing.T) {
	login1 := "login 1"
	password1 := "password 1"

	type on struct {
		input handler.CreateUserInput
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
			on{handler.CreateUserInput{Login: login1, Password: password1}},
			want{nil},
		},
		{
			"empty login",
			on{handler.CreateUserInput{Login: "", Password: password1}},
			want{handler.ErrInvalidUserLogin},
		},
		{
			"empty password",
			on{handler.CreateUserInput{Login: login1, Password: ""}},
			want{handler.ErrInvalidUserPassword},
		},
		{
			"invalid login length",
			on{handler.CreateUserInput{Login: strings.Repeat("l", MaxUserLoginLength+1), Password: password1}},
			want{handler.ErrInvalidUserLogin},
		},
		{
			"invalid password length",
			on{handler.CreateUserInput{Login: login1, Password: strings.Repeat("l", MaxUserPasswordLength+1)}},
			want{handler.ErrInvalidUserPassword},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := createUserInputValidator{Input: tc.on.input}
			err := validator.Validate()

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
		input handler.LoginUserInput
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
			on{handler.LoginUserInput{Login: login1, Password: password1}},
			want{nil},
		},
		{
			"empty login",
			on{handler.LoginUserInput{Login: "", Password: password1}},
			want{handler.ErrInvalidUserLogin},
		},
		{
			"empty password",
			on{handler.LoginUserInput{Login: login1, Password: ""}},
			want{handler.ErrInvalidUserPassword},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := loginUserInputValidator{Input: tc.on.input}
			err := validator.Validate()

			if tc.want.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.want.err.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}
