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
	"github.com/liebeSonne/gophermart/internal/service/async"
)

func TestUserOrderService_Upload(t *testing.T) {
	orderID1 := "12345678903"
	invalidOrderID1 := "111"
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
		err            error
		produceOrderID *string
	}
	testCases := []struct {
		name string
		on   on
		when when
		want want
	}{
		{
			"empty order id",
			on{UploadUserOrderInput{"", userID1}},
			when{},
			want{err: ErrInvalidOrderID},
		},
		{
			"invalid order id",
			on{UploadUserOrderInput{invalidOrderID1, userID1}},
			when{},
			want{err: ErrInvalidOrderID},
		},
		{
			"err on find user order",
			on{UploadUserOrderInput{orderID1, userID1}},
			when{findOrderErr: error1},
			want{err: error1},
		},
		{
			"already uploaded by other user",
			on{UploadUserOrderInput{orderID1, userID1}},
			when{findOrder: &model.UserOrder{OrderID: orderID1, UserID: userID2}},
			want{err: ErrUserOrderAlreadyUploadedByOtherUser},
		},
		{
			"already uploaded by user",
			on{UploadUserOrderInput{orderID1, userID1}},
			when{findOrder: &model.UserOrder{OrderID: orderID1, UserID: userID1}},
			want{err: ErrUserOrderAlreadyUploadedByUser},
		},
		{
			"error on store",
			on{UploadUserOrderInput{orderID1, userID1}},
			when{storeErr: error1},
			want{err: error1},
		},
		{
			"success",
			on{UploadUserOrderInput{orderID1, userID1}},
			when{},
			want{err: nil, produceOrderID: &orderID1},
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
			uowFactory.EXPECT().ExecuteWithUnitOfWork(t.Context(), mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ []string, fn func(provider uow.RepositoryProvider) error) error {
				return fn(repositoryProvider)
			}).Maybe()

			requestProducer := async.NewMockProducer[string](t)
			requestProducer.EXPECT().Add(tc.on.input.OrderID).RunAndReturn(func(value string) {
				if tc.want.produceOrderID != nil {
					assert.Equal(t, *tc.want.produceOrderID, value)
				}
			}).Maybe()

			s := NewUserOrderService(uowFactory, requestProducer)

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
