package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/liebeSonne/gophermart/internal/model"
	"github.com/liebeSonne/gophermart/internal/repository"
	"github.com/liebeSonne/gophermart/internal/repository/uow"
)

func TestUserBalanceService_AddWithdrawn(t *testing.T) {
	orderID1 := "12345678903"
	invalidOrderID1 := "111"
	userID1 := uuid.New()
	amountPositive1 := decimal.NewFromFloat(10.5)
	amountNegative1 := decimal.NewFromFloat(-10.5)
	error1 := errors.New("error 1")

	type when struct {
		getBalance        model.UserBalance
		getBalanceErr     error
		storeWithdrawnErr error
		storeBalanceErr   error
	}
	type on struct {
		input AddWithdrawnInput
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
			"invalid input userID",
			on{AddWithdrawnInput{uuid.Nil, orderID1, amountPositive1}},
			when{},
			want{ErrInvalidUserID},
		},
		{
			"empty input orderID",
			on{AddWithdrawnInput{userID1, "", amountPositive1}},
			when{},
			want{ErrInvalidOrderID},
		},
		{
			"invalid input orderID",
			on{AddWithdrawnInput{userID1, invalidOrderID1, amountPositive1}},
			when{},
			want{ErrInvalidOrderID},
		},
		{
			"invalid input amount",
			on{AddWithdrawnInput{userID1, orderID1, amountNegative1}},
			when{},
			want{ErrInvalidWithdrawnAmount},
		},
		{
			"error on get user balance",
			on{AddWithdrawnInput{userID1, orderID1, amountPositive1}},
			when{getBalanceErr: error1},
			want{error1},
		},
		{
			"balance in not enough",
			on{AddWithdrawnInput{userID1, orderID1, decimal.NewFromFloat(300.30)}},
			when{getBalance: model.UserBalance{UserID: userID1, Balance: decimal.NewFromFloat(100.10), WithdrawnSum: decimal.Zero}},
			want{ErrUserBalanceIsNotEnough},
		},
		{
			"error on store withdrawn",
			on{AddWithdrawnInput{userID1, orderID1, decimal.NewFromFloat(10.30)}},
			when{getBalance: model.UserBalance{UserID: userID1, Balance: decimal.NewFromFloat(100.10), WithdrawnSum: decimal.Zero}, storeWithdrawnErr: error1},
			want{error1},
		},
		{
			"error on store balance",
			on{AddWithdrawnInput{userID1, orderID1, decimal.NewFromFloat(10.30)}},
			when{getBalance: model.UserBalance{UserID: userID1, Balance: decimal.NewFromFloat(100.10), WithdrawnSum: decimal.Zero}, storeBalanceErr: error1},
			want{error1},
		},
		{
			"success on positive input amount",
			on{AddWithdrawnInput{userID1, orderID1, decimal.NewFromFloat(10.30)}},
			when{getBalance: model.UserBalance{UserID: userID1, Balance: decimal.NewFromFloat(100.10), WithdrawnSum: decimal.Zero}},
			want{nil},
		},
		{
			"success on positive input amount and positive withdrawn sum",
			on{AddWithdrawnInput{userID1, orderID1, decimal.NewFromFloat(10.30)}},
			when{getBalance: model.UserBalance{UserID: userID1, Balance: decimal.NewFromFloat(100.10), WithdrawnSum: decimal.NewFromFloat(300.30)}},
			want{nil},
		},
		{
			"success on zero input amount",
			on{AddWithdrawnInput{userID1, orderID1, decimal.Zero}},
			when{getBalance: model.UserBalance{UserID: userID1, Balance: decimal.NewFromFloat(100.10), WithdrawnSum: decimal.Zero}},
			want{nil},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userBalanceRepository := repository.NewMockUserBalanceRepository(t)
			userBalanceRepository.EXPECT().GetByUserID(t.Context(), mock.Anything).Return(tc.when.getBalance, tc.when.getBalanceErr).Maybe()
			userBalanceRepository.EXPECT().Store(t.Context(), mock.Anything).Return(tc.when.storeBalanceErr).Maybe()

			userBalanceWithdrawnRepository := repository.NewMockUserBalanceWithdrawnRepository(t)
			userBalanceWithdrawnRepository.EXPECT().NextID(t.Context()).Return(uuid.New()).Maybe()
			userBalanceWithdrawnRepository.EXPECT().Store(t.Context(), mock.Anything).Return(tc.when.storeWithdrawnErr).Maybe()

			repositoryProvider := uow.NewMockRepositoryProvider(t)
			repositoryProvider.EXPECT().UserBalanceRepository().Return(userBalanceRepository).Maybe()
			repositoryProvider.EXPECT().UserBalanceWithdrawnRepository().Return(userBalanceWithdrawnRepository).Maybe()

			uowFactory := uow.NewMockUnitOfWorkFactory(t)
			uowFactory.EXPECT().ExecuteWithUnitOfWork(t.Context(), mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ []string, fn func(provider uow.RepositoryProvider) error) error {
				return fn(repositoryProvider)
			}).Maybe()

			s := NewUserBalanceService(uowFactory)

			err := s.AddWithdrawn(t.Context(), tc.on.input)

			if tc.want.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.want.err.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}
