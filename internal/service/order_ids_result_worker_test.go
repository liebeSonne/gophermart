package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	ilogger "github.com/liebeSonne/gophermart/internal/logger"
	"github.com/liebeSonne/gophermart/internal/model"
	"github.com/liebeSonne/gophermart/internal/service/async"
)

func TestNewOrderIDResultWorker(t *testing.T) {
	error1 := errors.New("errro 1")
	orderID1 := "123"
	userID1 := uuid.New()
	unknownStatus := OrderStatus(-1)
	statusRegistered := OrderStatusRegistered
	statusProcessing := OrderStatusProcessing
	statusInvalid := OrderStatusInvalid
	statusProcessed := OrderStatusProcessed
	accrual1 := decimal.NewFromFloat(10.5)
	balance1 := decimal.NewFromFloat(100.55)

	retryDelay0 := time.Second * 33
	retryDelay1 := time.Second * 55
	retryDelay2 := time.Second * 77
	retryDelay3 := time.Second * 99
	retryDelay4 := time.Second * 111

	type on struct {
		result                       OrderIDWorkerResult
		retryDelay                   time.Duration
		minTooManyRequestsRetryDelay time.Duration
		maxTooManyRequestsRetryDelay time.Duration
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
		status  model.UserOrderStatus
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
			on{OrderIDWorkerResult{OrderID: orderID1}, retryDelay1, retryDelay2, retryDelay2},
			when{findUserIDErr: error1},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay1}},
		},
		{
			"error on find order",
			on{OrderIDWorkerResult{OrderID: orderID1}, retryDelay1, retryDelay2, retryDelay2},
			when{findUserID: &userID1, findOrderErr: error1},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay1}},
		},
		{
			"error on store order",
			on{OrderIDWorkerResult{OrderID: orderID1}, retryDelay1, retryDelay2, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.UserOrderStatusNew}, storeOrderErr: error1},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay1}},
		},
		{
			"not found user id by order",
			on{OrderIDWorkerResult{OrderID: orderID1}, retryDelay1, retryDelay2, retryDelay2},
			when{findUserID: nil},
			want{schedule: nil},
		},
		{
			"not found user order",
			on{OrderIDWorkerResult{OrderID: orderID1}, retryDelay1, retryDelay2, retryDelay2},
			when{findUserID: &userID1, findOrder: nil},
			want{schedule: nil},
		},
		{
			"result error",
			on{OrderIDWorkerResult{OrderID: orderID1, Err: error1}, retryDelay1, retryDelay2, retryDelay2},
			when{},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay1}},
		},
		{
			"result error to many requests",
			on{OrderIDWorkerResult{OrderID: orderID1, Err: ErrTooManyRequests}, retryDelay1, retryDelay2, retryDelay3},
			when{},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay2}},
		},
		{
			"result error to many requests with retry after more then default min retry delay",
			on{OrderIDWorkerResult{OrderID: orderID1, Err: NewErrTooManyRequestsRetryAfter(ErrTooManyRequests, retryDelay3)}, retryDelay1, retryDelay2, retryDelay4},
			when{},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay3}},
		},
		{
			"result error to many requests with retry after less then default min retry delay",
			on{OrderIDWorkerResult{OrderID: orderID1, Err: NewErrTooManyRequestsRetryAfter(ErrTooManyRequests, retryDelay0)}, retryDelay1, retryDelay2, retryDelay4},
			when{},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay2}},
		},
		{
			"result error to many requests with retry after more then default max retry delay",
			on{OrderIDWorkerResult{OrderID: orderID1, Err: NewErrTooManyRequestsRetryAfter(ErrTooManyRequests, retryDelay4)}, retryDelay1, retryDelay2, retryDelay3},
			when{},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay3}},
		},
		{
			"error on calculate new status",
			on{OrderIDWorkerResult{OrderID: orderID1, Status: &unknownStatus}, retryDelay1, retryDelay2, retryDelay2},
			when{},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay1}},
		},
		{
			"on already processed status",
			on{OrderIDWorkerResult{OrderID: orderID1, Status: &statusRegistered}, retryDelay1, retryDelay2, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.UserOrderStatusProcessed}},
			want{schedule: nil, storeOrder: nil},
		},
		{
			"new registered status",
			on{OrderIDWorkerResult{OrderID: orderID1, Status: &statusRegistered}, retryDelay1, retryDelay2, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.UserOrderStatusNew, Retries: 2, Accrual: nil}},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay1}, storeOrder: &storeOder{orderID: orderID1, status: model.UserOrderStatusNew, retries: 3, accrual: nil}},
		},
		{
			"new registered status with accrual",
			on{OrderIDWorkerResult{OrderID: orderID1, Status: &statusRegistered, Accrual: &accrual1}, retryDelay1, retryDelay2, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.UserOrderStatusNew, Retries: 2, Accrual: nil}},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay1}, storeOrder: &storeOder{orderID: orderID1, status: model.UserOrderStatusNew, retries: 3, accrual: nil}},
		},
		{
			"new processing status",
			on{OrderIDWorkerResult{OrderID: orderID1, Status: &statusProcessing}, retryDelay1, retryDelay2, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.UserOrderStatusNew, Retries: 2, Accrual: nil}},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay1}, storeOrder: &storeOder{orderID: orderID1, status: model.UserOrderStatusProcessing, retries: 3, accrual: nil}},
		},
		{
			"new processing status with accrual",
			on{OrderIDWorkerResult{OrderID: orderID1, Status: &statusProcessing, Accrual: &accrual1}, retryDelay1, retryDelay2, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.UserOrderStatusNew, Retries: 2, Accrual: nil}},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay1}, storeOrder: &storeOder{orderID: orderID1, status: model.UserOrderStatusProcessing, retries: 3, accrual: nil}},
		},
		{
			"new invalid status",
			on{OrderIDWorkerResult{OrderID: orderID1, Status: &statusInvalid}, retryDelay1, retryDelay2, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.UserOrderStatusNew, Retries: 2, Accrual: nil}},
			want{schedule: nil, storeOrder: &storeOder{orderID: orderID1, status: model.UserOrderStatusInvalid, retries: 3, accrual: nil}},
		},
		{
			"new invalid status with accrual",
			on{OrderIDWorkerResult{OrderID: orderID1, Status: &statusInvalid, Accrual: &accrual1}, retryDelay1, retryDelay2, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.UserOrderStatusNew, Retries: 2, Accrual: nil}},
			want{schedule: nil, storeOrder: &storeOder{orderID: orderID1, status: model.UserOrderStatusInvalid, retries: 3, accrual: nil}},
		},
		{
			"new processed status",
			on{OrderIDWorkerResult{OrderID: orderID1, Status: &statusProcessed}, retryDelay1, retryDelay2, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.UserOrderStatusNew, Retries: 2, Accrual: nil, UserID: userID1}},
			want{schedule: nil, storeOrder: &storeOder{orderID: orderID1, status: model.UserOrderStatusProcessed, retries: 3, accrual: nil}},
		},
		{
			"new processed status with accrual",
			on{OrderIDWorkerResult{OrderID: orderID1, Status: &statusProcessed, Accrual: &accrual1}, retryDelay1, retryDelay2, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.UserOrderStatusNew, Retries: 2, Accrual: nil}, getBalance: model.UserBalance{UserID: userID1, Balance: balance1}},
			want{schedule: nil, storeOrder: &storeOder{orderID: orderID1, status: model.UserOrderStatusProcessed, retries: 3, accrual: &accrual1}, storeBalance: &storeBalance{userID: userID1, balance: balance1.Add(accrual1)}},
		},
		{
			"new processed status with accrual = 0",
			on{OrderIDWorkerResult{OrderID: orderID1, Status: &statusProcessed, Accrual: &decimal.Zero}, retryDelay1, retryDelay2, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.UserOrderStatusNew, Retries: 2, Accrual: nil}},
			want{schedule: nil, storeOrder: &storeOder{orderID: orderID1, status: model.UserOrderStatusProcessed, retries: 3, accrual: &decimal.Zero}, storeBalance: nil},
		},
		{
			"error on store user balance",
			on{OrderIDWorkerResult{OrderID: orderID1, Status: &statusProcessed, Accrual: &accrual1}, retryDelay1, retryDelay2, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.UserOrderStatusNew}, getBalance: model.UserBalance{UserID: userID1, Balance: balance1}, storeBalanceErr: error1},
			want{schedule: &schedule{orderID: orderID1, delay: retryDelay1}},
		},
		{
			"error on get balance",
			on{OrderIDWorkerResult{OrderID: orderID1, Status: &statusProcessed, Accrual: &accrual1}, retryDelay1, retryDelay2, retryDelay2},
			when{findUserID: &userID1, findOrder: &model.UserOrder{OrderID: orderID1, Status: model.UserOrderStatusNew, Retries: 2, Accrual: nil}, getBalanceErr: error1},
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

			userOrderProvider := NewMockUserOrderProvider(t)
			userOrderProvider.EXPECT().FindUserIDByOrderID(mock.Anything, mock.Anything).Return(tc.when.findUserID, tc.when.findUserIDErr).Maybe()

			userOrderRepository := NewMockUserOrderRepository(t)
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

			userBalanceRepository := NewMockUserBalanceRepository(t)
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

			repositoryProvider := NewMockRepositoryProvider(t)
			repositoryProvider.EXPECT().UserOrderRepository().Return(userOrderRepository).Maybe()
			repositoryProvider.EXPECT().UserBalanceRepository().Return(userBalanceRepository).Maybe()

			uowFactory := NewMockUnitOfWorkFactory(t)
			uowFactory.EXPECT().ExecuteWithUnitOfWork(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, _ []string, fn func(provider RepositoryProvider) error) error {
				return fn(repositoryProvider)
			}).Maybe()

			l := ilogger.NewNullLogger()

			w := NewOrderIDResultWorker(
				ctx,
				"name",
				tc.on.retryDelay,
				tc.on.minTooManyRequestsRetryDelay,
				tc.on.maxTooManyRequestsRetryDelay,
				retryProducer,
				uowFactory,
				userOrderProvider,
				l,
			)

			_ = w.Handle(tc.on.result)
		})
	}
}

func TestOrderIDResultWorker_SleepingHandle(t *testing.T) {
	retryDelay := time.Second * 20
	minTooManyRequestsRetryDelay := time.Second * 10
	maxTooManyRequestsRetryDelay := time.Second * 40

	type on struct {
		result struct{}
	}
	type want struct {
		needDelay bool
		delayTime time.Duration
	}
	testCases := []struct {
		name string
		on   on
		want want
	}{
		{
			"on any result",
			on{},
			want{false, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			retryProducer := async.NewMockProducer[string](t)
			userOrderProvider := NewMockUserOrderProvider(t)
			uowFactory := NewMockUnitOfWorkFactory(t)

			l := ilogger.NewNullLogger()

			w := NewOrderIDResultWorker(
				ctx,
				"name",
				retryDelay,
				minTooManyRequestsRetryDelay,
				maxTooManyRequestsRetryDelay,
				retryProducer,
				uowFactory,
				userOrderProvider,
				l,
			)

			needDelay, delayTime := w.SleepingHandle(tc.on.result)

			assert.Equal(t, tc.want.needDelay, needDelay)
			assert.Equal(t, tc.want.delayTime, delayTime)
		})
	}
}
