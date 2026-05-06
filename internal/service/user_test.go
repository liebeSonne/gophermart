package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/liebeSonne/gophermart/internal/model"
	"github.com/liebeSonne/gophermart/internal/provider"
	"github.com/liebeSonne/gophermart/internal/repository"
	"github.com/liebeSonne/gophermart/internal/repository/uow"
)

// nolint:goconst
func TestUserService_Create(t *testing.T) {
	login1 := "login 1"
	password1 := "password 1"
	passHash1 := "pass hash 1"
	passHash2 := "pass hash 2"
	user1 := model.User{Login: login1, PassHash: passHash2}
	error1 := errors.New("error 1")

	type when struct {
		findUser    *model.User
		findUserErr error
		passHash    string
		passHashErr error
		storeErr    error
	}
	type on struct {
		input CreateUserInput
	}
	type want struct {
		err error
	}
	testCases := []struct {
		name string
		on   on
		when when
		want want
	}{
		{
			"not found user with login",
			on{CreateUserInput{login1, password1}},
			when{nil, nil, passHash1, nil, nil},
			want{nil},
		},
		{
			"exist user with login",
			on{CreateUserInput{login1, password1}},
			when{&user1, nil, passHash1, nil, nil},
			want{ErrUserLoginExists},
		},
		{
			"empty user login",
			on{CreateUserInput{"", password1}},
			when{nil, nil, passHash1, nil, nil},
			want{ErrInvalidUserLogin},
		},
		{
			"invalid user login length",
			on{CreateUserInput{strings.Repeat("l", MaxUserLoginLength+1), password1}},
			when{nil, nil, passHash1, nil, nil},
			want{ErrInvalidUserLogin},
		},
		{
			"empty user password",
			on{CreateUserInput{login1, ""}},
			when{nil, nil, passHash1, nil, nil},
			want{ErrInvalidUserPassword},
		},
		{
			"invalid user password length",
			on{CreateUserInput{login1, strings.Repeat("p", MaxUserPasswordLength+1)}},
			when{nil, nil, passHash1, nil, nil},
			want{ErrInvalidUserPassword},
		},
		{
			"error on find user",
			on{CreateUserInput{login1, password1}},
			when{nil, error1, passHash1, nil, nil},
			want{error1},
		},
		{
			"error on generate password hash",
			on{CreateUserInput{login1, password1}},
			when{nil, nil, passHash1, error1, nil},
			want{error1},
		},
		{
			"error on store user",
			on{CreateUserInput{login1, password1}},
			when{nil, nil, passHash1, nil, error1},
			want{error1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userRepository := repository.NewMockUserRepository(t)
			userRepository.EXPECT().FindByLogin(t.Context(), tc.on.input.Login).Return(tc.when.findUser, tc.when.findUserErr).Maybe()
			userRepository.EXPECT().NextID(t.Context()).Return(uuid.New()).Maybe()
			userRepository.EXPECT().Store(t.Context(), mock.Anything).Return(tc.when.storeErr).Maybe()

			repositoryProvider := uow.NewMockRepositoryProvider(t)
			repositoryProvider.EXPECT().UserRepository().Return(userRepository).Maybe()

			uowFactory := uow.NewMockUnitOfWorkFactory(t)
			uowFactory.EXPECT().ExecuteWithUnitOfWork(t.Context(), mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ []string, fn func(provider uow.RepositoryProvider) error) error {
				return fn(repositoryProvider)
			}).Maybe()

			userProvider := provider.NewMockUserProvider(t)

			passwordService := NewMockPasswordService(t)
			passwordService.EXPECT().CreateHash(t.Context(), tc.on.input.Password).Return(tc.when.passHash, tc.when.passHashErr).Maybe()

			s := NewUserService(uowFactory, passwordService, userProvider)

			user, err := s.Create(t.Context(), tc.on.input)

			if tc.want.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.want.err.Error())
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.on.input.Login, user.Login)
		})
	}
}

func TestUserService_CheckPassword(t *testing.T) {
	login1 := "login 1"
	password1 := "password 1"
	passHash1 := "pass hash 1"
	user1 := model.User{Login: login1, PassHash: passHash1}
	error1 := errors.New("error 1")

	type when struct {
		findUser     *model.User
		findUserErr  error
		checkPass    bool
		checkPassErr error
	}
	type on struct {
		input LoginUserInput
	}
	type want struct {
		err error
	}
	testCases := []struct {
		name string
		on   on
		when when
		want want
	}{
		{
			"success login",
			on{LoginUserInput{login1, password1}},
			when{&user1, nil, true, nil},
			want{nil},
		},
		{
			"not valid user login password",
			on{LoginUserInput{login1, password1}},
			when{&user1, nil, false, nil},
			want{ErrNotValidUserLoginPassword},
		},
		{
			"empty user login",
			on{LoginUserInput{"", password1}},
			when{nil, nil, false, nil},
			want{ErrInvalidUserLogin},
		},
		{
			"empty user password",
			on{LoginUserInput{login1, ""}},
			when{nil, nil, false, nil},
			want{ErrInvalidUserPassword},
		},
		{
			"error on find user",
			on{LoginUserInput{login1, password1}},
			when{nil, error1, false, nil},
			want{error1},
		},
		{
			"user not found",
			on{LoginUserInput{login1, password1}},
			when{nil, nil, false, nil},
			want{ErrUserNotFound},
		},
		{
			"error on check password hash",
			on{LoginUserInput{login1, password1}},
			when{&user1, nil, false, error1},
			want{error1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uowFactory := uow.NewMockUnitOfWorkFactory(t)

			userProvider := provider.NewMockUserProvider(t)
			userProvider.EXPECT().FindByLogin(t.Context(), tc.on.input.Login).Return(tc.when.findUser, tc.when.findUserErr).Maybe()

			passwordService := NewMockPasswordService(t)
			passwordService.EXPECT().CheckHash(t.Context(), tc.on.input.Password, mock.Anything).Return(tc.when.checkPass, tc.when.checkPassErr).Maybe()

			s := NewUserService(uowFactory, passwordService, userProvider)

			user, err := s.Login(t.Context(), tc.on.input)

			if tc.want.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.want.err.Error())
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.on.input.Login, user.Login)
		})
	}
}
