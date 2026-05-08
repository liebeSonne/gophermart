package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/liebeSonne/gophermart/internal/adapter"
	"github.com/liebeSonne/gophermart/internal/model"
	"github.com/liebeSonne/gophermart/internal/provider"
	"github.com/liebeSonne/gophermart/internal/repository"
	"github.com/liebeSonne/gophermart/internal/repository/uow"
	"github.com/liebeSonne/gophermart/internal/service/async"
)

func TestNewOrderIDResultWorker(t *testing.T) {
	error1 := errors.New("errro 1")
	orderID1 := "123"
	userID1 := uuid.New()
	unknownStatus := adapter.OrderStatus(-1)
	statusRegistered := adapter.OrderStatusRegistered
	statusProcessing := adapter.OrderStatusProcessing
	statusInvalid := adapter.OrderStatusInvalid
	statusProcessed := adapter.OrderStatusProcessed
	accrual1 := decimal.NewFromFloat(10.5)
	balance1 := decimal.NewFromFloat(100.55)

	retryDelay0 := time.Second * 33
	retryDelay1 := time.Second * 55
	retryDelay2 := time.Second * 77
	retryDelay3 := time.Second * 99

	type on struct {
		result              async.OrderIDWorkerResult
		retryDelay          time.Duration
		tooManyRetriesDelay time.Duration
	}
	type when struct {
		findOrder       *model.UserOrder
		findOrderErr    error
		storeOrderErr   error
		findUserID      *uuid.UUID
		findUserIDErr   error
		getBalance      model.UserBalance
		getBalanceErr   error
		storeBalanceErr error
	}
	type schedule struct {
		orderID string
		delay   time.Duration
	}
	type storeOder struct {
		orderID string
		status  model.OrderStatus
		retries int
		accrual *decimal.Decimal
	}
	type storeBalance struct {
		userID  uuid.UUID
		balance decimal.Decimal
	}
	type want struct {
		schedule     *schedule
		storeOrder   *storeOder
		storeBalance *storeBalance
	}
	testCases := []struct {
		name string
		on   on
		when when
		want want
	}{
		{
			"error on find user id by order",
			on{async.OrderIDWorkerResult{OrderID: orderID1}, retryDelay1, retryDelay2},
			when{findUserIDErr: error1},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay1}},
		},
		{
			"error on find order",
			on{async.OrderIDWorkerResult{OrderID: orderID1}, retryDelay1, retryDelay2},
			when{findUserID: &userID1, findOrderErr: error1},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay1}},
		},
		{
			"error on store order",
			on{async.OrderIDWorkerResult{OrderID: orderID1}, retryDelay1, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.OrderStatusNew}, storeOrderErr: error1},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay1}},
		},
		{
			"not found user id by order",
			on{async.OrderIDWorkerResult{OrderID: orderID1}, retryDelay1, retryDelay2},
			when{findUserID: nil},
			want{schedule: nil},
		},
		{
			"not found user order",
			on{async.OrderIDWorkerResult{OrderID: orderID1}, retryDelay1, retryDelay2},
			when{findUserID: &userID1, findOrder: nil},
			want{schedule: nil},
		},
		{
			"result error",
			on{async.OrderIDWorkerResult{OrderID: orderID1, Err: error1}, retryDelay1, retryDelay2},
			when{},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay1}},
		},
		{
			"result error to many retries",
			on{async.OrderIDWorkerResult{OrderID: orderID1, Err: adapter.ErrTooManyRetries}, retryDelay1, retryDelay2},
			when{},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay2}},
		},
		{
			"result error to many retries with retry after more then default retry delay",
			on{async.OrderIDWorkerResult{OrderID: orderID1, Err: adapter.NewErrTooManyRetriesRetryAfter(adapter.ErrTooManyRetries, retryDelay3)}, retryDelay1, retryDelay2},
			when{},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay3}},
		},
		{
			"result error to many retries with retry after less then default retry delay",
			on{async.OrderIDWorkerResult{OrderID: orderID1, Err: adapter.NewErrTooManyRetriesRetryAfter(adapter.ErrTooManyRetries, retryDelay0)}, retryDelay1, retryDelay2},
			when{},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay2}},
		},
		{
			"error on calculate new status",
			on{async.OrderIDWorkerResult{OrderID: orderID1, Status: &unknownStatus}, retryDelay1, retryDelay2},
			when{},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay1}},
		},
		{
			"on already processed status",
			on{async.OrderIDWorkerResult{OrderID: orderID1, Status: &statusRegistered}, retryDelay1, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.OrderStatusProcessed}},
			want{schedule: nil, storeOrder: nil},
		},
		{
			"new registered status",
			on{async.OrderIDWorkerResult{OrderID: orderID1, Status: &statusRegistered}, retryDelay1, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.OrderStatusNew, Retries: 2, Accrual: nil}},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay1}, storeOrder: &storeOder{orderID: orderID1, status: model.OrderStatusNew, retries: 3, accrual: nil}},
		},
		{
			"new registered status with accrual",
			on{async.OrderIDWorkerResult{OrderID: orderID1, Status: &statusRegistered, Accrual: &accrual1}, retryDelay1, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.OrderStatusNew, Retries: 2, Accrual: nil}},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay1}, storeOrder: &storeOder{orderID: orderID1, status: model.OrderStatusNew, retries: 3, accrual: nil}},
		},
		{
			"new processing status",
			on{async.OrderIDWorkerResult{OrderID: orderID1, Status: &statusProcessing}, retryDelay1, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.OrderStatusNew, Retries: 2, Accrual: nil}},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay1}, storeOrder: &storeOder{orderID: orderID1, status: model.OrderStatusProcessing, retries: 3, accrual: nil}},
		},
		{
			"new processing status with accrual",
			on{async.OrderIDWorkerResult{OrderID: orderID1, Status: &statusProcessing, Accrual: &accrual1}, retryDelay1, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.OrderStatusNew, Retries: 2, Accrual: nil}},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay1}, storeOrder: &storeOder{orderID: orderID1, status: model.OrderStatusProcessing, retries: 3, accrual: nil}},
		},
		{
			"new invalid status",
			on{async.OrderIDWorkerResult{OrderID: orderID1, Status: &statusInvalid}, retryDelay1, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.OrderStatusNew, Retries: 2, Accrual: nil}},
			want{schedule: nil, storeOrder: &storeOder{orderID: orderID1, status: model.OrderStatusInvalid, retries: 3, accrual: nil}},
		},
		{
			"new invalid status with accrual",
			on{async.OrderIDWorkerResult{OrderID: orderID1, Status: &statusInvalid, Accrual: &accrual1}, retryDelay1, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.OrderStatusNew, Retries: 2, Accrual: nil}},
			want{schedule: nil, storeOrder: &storeOder{orderID: orderID1, status: model.OrderStatusInvalid, retries: 3, accrual: nil}},
		},
		{
			"new processed status",
			on{async.OrderIDWorkerResult{OrderID: orderID1, Status: &statusProcessed}, retryDelay1, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.OrderStatusNew, Retries: 2, Accrual: nil, UserID: userID1}},
			want{schedule: nil, storeOrder: &storeOder{orderID: orderID1, status: model.OrderStatusProcessed, retries: 3, accrual: nil}},
		},
		{
			"new processed status with accrual",
			on{async.OrderIDWorkerResult{OrderID: orderID1, Status: &statusProcessed, Accrual: &accrual1}, retryDelay1, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.OrderStatusNew, Retries: 2, Accrual: nil}, getBalance: model.UserBalance{UserID: userID1, Balance: balance1}},
			want{schedule: nil, storeOrder: &storeOder{orderID: orderID1, status: model.OrderStatusProcessed, retries: 3, accrual: &accrual1}, storeBalance: &storeBalance{userID: userID1, balance: balance1.Add(accrual1)}},
		},
		{
			"new processed status with accrual = 0",
			on{async.OrderIDWorkerResult{OrderID: orderID1, Status: &statusProcessed, Accrual: &decimal.Zero}, retryDelay1, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.OrderStatusNew, Retries: 2, Accrual: nil}},
			want{schedule: nil, storeOrder: &storeOder{orderID: orderID1, status: model.OrderStatusProcessed, retries: 3, accrual: &decimal.Zero}, storeBalance: nil},
		},
		{
			"error on store user balance",
			on{async.OrderIDWorkerResult{OrderID: orderID1, Status: &statusProcessed, Accrual: &accrual1}, retryDelay1, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.OrderStatusNew}, getBalance: model.UserBalance{UserID: userID1, Balance: balance1}, storeBalanceErr: error1},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay1}},
		},
		{
			"error on get balance",
			on{async.OrderIDWorkerResult{OrderID: orderID1, Status: &statusProcessed, Accrual: &accrual1}, retryDelay1, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.OrderStatusNew, Retries: 2, Accrual: nil}, getBalanceErr: error1},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay1}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()

			retryProducer := async.NewMockProducer[string](t)
			if tc.want.schedule != nil {
				retryProducer.EXPECT().Schedule(tc.want.schedule.orderID, tc.want.schedule.delay).Return(nil).Once()
			}

			userOrderProvider := provider.NewMockUserOrderProvider(t)
			userOrderProvider.EXPECT().FindUserIDByOrderID(mock.Anything, mock.Anything).Return(tc.when.findUserID, tc.when.findUserIDErr).Maybe()

			userOrderRepository := repository.NewMockUserOrderRepository(t)
			userOrderRepository.EXPECT().FindByOrderID(mock.Anything, mock.Anything).Return(tc.when.findOrder, tc.when.findOrderErr).Maybe()
			userOrderRepository.EXPECT().Store(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, items []model.UserOrder) error {
				if tc.when.storeOrderErr != nil {
					return tc.when.storeOrderErr
				}
				if tc.want.storeOrder != nil {
					require.Len(t, items, 1)
					item := items[0]
					assert.Equal(t, tc.want.storeOrder.orderID, item.OrderID)
					assert.Equal(t, tc.want.storeOrder.status, item.Status)
					assert.Equal(t, tc.want.storeOrder.retries, item.Retries)
					assert.Equal(t, tc.want.storeOrder.accrual, item.Accrual)
				}
				return nil
			}).Maybe()

			userBalanceRepository := repository.NewMockUserBalanceRepository(t)
			userBalanceRepository.EXPECT().GetByUserID(mock.Anything, mock.Anything).Return(tc.when.getBalance, tc.when.getBalanceErr).Maybe()
			userBalanceRepository.EXPECT().Store(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, userBalance model.UserBalance) error {
				if tc.when.storeBalanceErr != nil {
					return tc.when.storeBalanceErr
				}
				if tc.want.storeBalance != nil {
					assert.Equal(t, tc.want.storeBalance.userID, userBalance.UserID)
					assert.Equal(t, tc.want.storeBalance.balance, userBalance.Balance)
				}
				return nil
			}).Maybe()

			repositoryProvider := uow.NewMockRepositoryProvider(t)
			repositoryProvider.EXPECT().UserOrderRepository().Return(userOrderRepository).Maybe()
			repositoryProvider.EXPECT().UserBalanceRepository().Return(userBalanceRepository).Maybe()

			uowFactory := uow.NewMockUnitOfWorkFactory(t)
			uowFactory.EXPECT().ExecuteWithUnitOfWork(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ []string, fn func(provider uow.RepositoryProvider) error) error {
				return fn(repositoryProvider)
			}).Maybe()

			l, _ := test.NewNullLogger()

			w := NewOrderIDResultWorker(
				ctx,
				"name",
				tc.on.retryDelay,
				tc.on.tooManyRetriesDelay,
				retryProducer,
				uowFactory,
				userOrderProvider,
				l,
			)

			outCh := make(chan struct{})

			w.Handle(tc.on.result, outCh)
		})
	}
}
