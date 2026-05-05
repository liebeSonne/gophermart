package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/liebeSonne/gophermart/internal/model"
	"github.com/liebeSonne/gophermart/internal/repository"
	"github.com/liebeSonne/gophermart/internal/repository/uow"
)

func TestUserOrderService_Upload(t *testing.T) {
	orderID1 := "123"
	userID1 := uuid.New()
	userID2 := uuid.New()
	error1 := errors.New("error 1")

	type when struct {
		findOrder    *model.UserOrder
		findOrderErr error
		storeErr     error
	}
	type on struct {
		input UploadUserOrderInput
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
			"invalid order id",
			on{UploadUserOrderInput{"", userID1}},
			when{},
			want{ErrInvalidOrderID},
		},
		{
			"err on find user order",
			on{UploadUserOrderInput{orderID1, userID1}},
			when{findOrderErr: error1},
			want{error1},
		},
		{
			"already uploaded by other user",
			on{UploadUserOrderInput{orderID1, userID1}},
			when{findOrder: &model.UserOrder{OrderID: orderID1, UserID: userID2}},
			want{ErrUserOrderAlreadyUploadedByOtherUser},
		},
		{
			"already uploaded by user",
			on{UploadUserOrderInput{orderID1, userID1}},
			when{findOrder: &model.UserOrder{OrderID: orderID1, UserID: userID1}},
			want{ErrUserOrderAlreadyUploadedByUser},
		},
		{
			"error on store",
			on{UploadUserOrderInput{orderID1, userID1}},
			when{storeErr: error1},
			want{error1},
		},
		{
			"success",
			on{UploadUserOrderInput{orderID1, userID1}},
			when{},
			want{nil},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userOrderRepository := repository.NewMockUserOrderRepository(t)
			userOrderRepository.EXPECT().FindByOrderID(t.Context(), tc.on.input.OrderID).Return(tc.when.findOrder, tc.when.findOrderErr).Maybe()
			userOrderRepository.EXPECT().NextID(t.Context()).Return(uuid.New()).Maybe()
			userOrderRepository.EXPECT().Store(t.Context(), mock.Anything).Return(tc.when.storeErr).Maybe()

			repositoryProvider := uow.NewMockRepositoryProvider(t)
			repositoryProvider.EXPECT().UserOrderRepository().Return(userOrderRepository).Maybe()

			uowFactory := uow.NewMockUnitOfWorkFactory(t)
			uowFactory.EXPECT().ExecuteWithUnitOfWork(t.Context(), mock.Anything).RunAndReturn(func(_ context.Context, f func(provider uow.RepositoryProvider) error) error {
				return f(repositoryProvider)
			}).Maybe()

			s := NewUserOrderService(uowFactory)

			userOrder, err := s.Upload(t.Context(), tc.on.input)

			if tc.want.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.want.err.Error())
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.on.input.OrderID, userOrder.OrderID)
			assert.Equal(t, tc.on.input.UserID, userOrder.UserID)
		})
	}
}
