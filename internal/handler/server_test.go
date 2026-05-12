package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/liebeSonne/gophermart/internal/auth"
	ilogger "github.com/liebeSonne/gophermart/internal/logger"
	"github.com/liebeSonne/gophermart/internal/model"
)

//nolint:dupl
func TestServer_RegisterUser(t *testing.T) {
	cookieTokenKey := "access_token"

	type on struct {
		method string
		body   string
		header map[string]string
	}
	type when struct {
		createUser      model.User
		createUserErr   error
		createToken     string
		createTokenErr  error
		setAuthTokenErr error
	}
	type want struct {
		code   int
		header map[string]string
		cookie map[string]string
		body   *string
	}
	testCases := []struct {
		name string
		on   on
		when when
		want want
	}{
		{
			"error on decode body",
			on{
				method: http.MethodPost,
				body:   `"invalid body"`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			when{},
			want{code: http.StatusBadRequest},
		},
		{
			"error login exist on create user",
			on{
				method: http.MethodPost,
				body:   `{"login": "user1", "password": "123"}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			when{createUserErr: ErrUserLoginExists},
			want{code: http.StatusConflict},
		},
		{
			"error invalid login on create user",
			on{
				method: http.MethodPost,
				body:   `{"login": "user1", "password": "123"}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			when{createUserErr: ErrInvalidUserLogin},
			want{code: http.StatusBadRequest},
		},
		{
			"error invalid password on create user",
			on{
				method: http.MethodPost,
				body:   `{"login": "user1", "password": "123"}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			when{createUserErr: ErrInvalidUserPassword},
			want{code: http.StatusBadRequest},
		},
		{
			"unexpected error on create user",
			on{
				method: http.MethodPost,
				body:   `{"login": "user1", "password": "123"}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			when{createUserErr: errors.New("error 1")},
			want{code: http.StatusInternalServerError},
		},
		{
			"error on create token",
			on{
				method: http.MethodPost,
				body:   `{"login": "user1", "password": "123"}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			when{createUser: model.User{ID: uuid.New(), Login: "user1", PassHash: "passHash1"}, createTokenErr: errors.New("error 1")},
			want{code: http.StatusInternalServerError},
		},
		{
			"error on set auth token",
			on{
				method: http.MethodPost,
				body:   `{"login": "user1", "password": "123"}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			when{createUser: model.User{ID: uuid.New(), Login: "user1", PassHash: "passHash1"}, createToken: "token1", setAuthTokenErr: errors.New("error 1")},
			want{code: http.StatusInternalServerError},
		},
		{
			"success register user",
			on{
				method: http.MethodPost,
				body:   `{"login": "user1", "password": "123"}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			when{createUser: model.User{ID: uuid.New(), Login: "user1", PassHash: "passHash1"}, createToken: "token1"},
			want{code: http.StatusOK, header: map[string]string{"Content-Type": "application/json"}, cookie: map[string]string{cookieTokenKey: "token1"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userService := NewMockUserService(t)
			userService.EXPECT().Create(mock.Anything, mock.Anything).Return(tc.when.createUser, tc.when.createUserErr).Maybe()

			userOrderService := NewMockUserOrderService(t)
			userOrderQueryService := NewMockUserOrderQueryService(t)
			userBalanceQueryService := NewMockUserBalanceQueryService(t)
			userBalanceService := NewMockUserBalanceService(t)
			userBalanceWithDrawnQueryService := NewMockUserBalanceWithDrawnQueryService(t)

			tokenService := NewMockTokenService(t)
			tokenService.EXPECT().Create(mock.Anything).Return(tc.when.createToken, tc.when.createTokenErr).Maybe()

			cookieService := NewMockCookieService(t)
			cookieService.EXPECT().SetAuthToken(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(tokenString string, w http.ResponseWriter, r *http.Request) error {
				if tc.when.setAuthTokenErr != nil {
					return tc.when.setAuthTokenErr
				}
				cookie := &http.Cookie{
					Name:  cookieTokenKey,
					Value: tokenString,
				}
				http.SetCookie(w, cookie)
				r.AddCookie(cookie)
				return nil
			}).Maybe()

			l := ilogger.NewNullLogger()

			srv := NewServer(
				userService,
				userOrderService,
				userOrderQueryService,
				userBalanceQueryService,
				userBalanceService,
				userBalanceWithDrawnQueryService,
				tokenService,
				cookieService,
				l,
			)

			req := httptest.NewRequest(tc.on.method, "/api/user/register", strings.NewReader(tc.on.body))
			for k, v := range tc.on.header {
				req.Header.Set(k, v)
			}

			rr := httptest.NewRecorder()

			srv.RegisterUser(rr, req)

			require.Equal(t, tc.want.code, rr.Code)

			for k, v := range tc.want.header {
				assert.Equal(t, v, rr.Header().Get(k))
			}

			resp := rr.Result()
			cookies := resp.Cookies()
			for k, v := range tc.want.cookie {
				for _, c := range cookies {
					if c.Name == k {
						assert.Equal(t, v, c.Value)
					}
				}
			}

			if tc.want.body != nil {
				assert.Equal(t, *tc.want.body, rr.Body.String())
			}
		})
	}
}

//nolint:dupl
func TestServer_LoginUser(t *testing.T) {
	cookieTokenKey := "access_token"

	type on struct {
		method string
		body   string
		header map[string]string
	}
	type when struct {
		loginUser       model.User
		loginUserErr    error
		createToken     string
		createTokenErr  error
		setAuthTokenErr error
	}
	type want struct {
		code   int
		header map[string]string
		cookie map[string]string
		body   *string
	}
	testCases := []struct {
		name string
		on   on
		when when
		want want
	}{
		{
			"error on decode body",
			on{
				method: http.MethodPost,
				body:   `"invalid body"`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			when{},
			want{code: http.StatusBadRequest},
		},
		{
			"error invalid login on login user",
			on{
				method: http.MethodPost,
				body:   `{"login": "user1", "password": "123"}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			when{loginUserErr: ErrInvalidUserLogin},
			want{code: http.StatusBadRequest},
		},
		{
			"error invalid password on login user",
			on{
				method: http.MethodPost,
				body:   `{"login": "user1", "password": "123"}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			when{loginUserErr: ErrInvalidUserPassword},
			want{code: http.StatusBadRequest},
		},
		{
			"error user not found on login user",
			on{
				method: http.MethodPost,
				body:   `{"login": "user1", "password": "123"}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			when{loginUserErr: ErrUserNotFound},
			want{code: http.StatusUnauthorized},
		},
		{
			"error not valid login on login user",
			on{
				method: http.MethodPost,
				body:   `{"login": "user1", "password": "123"}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			when{loginUserErr: ErrNotValidUserLoginPassword},
			want{code: http.StatusUnauthorized},
		},
		{
			"unexpected error on create user",
			on{
				method: http.MethodPost,
				body:   `{"login": "user1", "password": "123"}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			when{loginUserErr: errors.New("error 1")},
			want{code: http.StatusInternalServerError},
		},
		{
			"error on create token",
			on{
				method: http.MethodPost,
				body:   `{"login": "user1", "password": "123"}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			when{loginUser: model.User{ID: uuid.New(), Login: "user1", PassHash: "passHash1"}, createTokenErr: errors.New("error 1")},
			want{code: http.StatusInternalServerError},
		},
		{
			"error on set auth token",
			on{
				method: http.MethodPost,
				body:   `{"login": "user1", "password": "123"}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			when{loginUser: model.User{ID: uuid.New(), Login: "user1", PassHash: "passHash1"}, createToken: "token1", setAuthTokenErr: errors.New("error 1")},
			want{code: http.StatusInternalServerError},
		},
		{
			"success login user",
			on{
				method: http.MethodPost,
				body:   `{"login": "user1", "password": "123"}`,
				header: map[string]string{"Content-Type": "application/json"},
			},
			when{loginUser: model.User{ID: uuid.New(), Login: "user1", PassHash: "passHash1"}, createToken: "token1"},
			want{code: http.StatusOK, header: map[string]string{"Content-Type": "application/json"}, cookie: map[string]string{cookieTokenKey: "token1"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userService := NewMockUserService(t)
			userService.EXPECT().Login(mock.Anything, mock.Anything).Return(tc.when.loginUser, tc.when.loginUserErr).Maybe()

			userOrderService := NewMockUserOrderService(t)
			userOrderQueryService := NewMockUserOrderQueryService(t)
			userBalanceQueryService := NewMockUserBalanceQueryService(t)
			userBalanceService := NewMockUserBalanceService(t)
			userBalanceWithDrawnQueryService := NewMockUserBalanceWithDrawnQueryService(t)

			tokenService := NewMockTokenService(t)
			tokenService.EXPECT().Create(mock.Anything).Return(tc.when.createToken, tc.when.createTokenErr).Maybe()

			cookieService := NewMockCookieService(t)
			cookieService.EXPECT().SetAuthToken(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(tokenString string, w http.ResponseWriter, r *http.Request) error {
				if tc.when.setAuthTokenErr != nil {
					return tc.when.setAuthTokenErr
				}
				cookie := &http.Cookie{
					Name:  cookieTokenKey,
					Value: tokenString,
				}
				http.SetCookie(w, cookie)
				r.AddCookie(cookie)
				return nil
			}).Maybe()

			l := ilogger.NewNullLogger()

			srv := NewServer(
				userService,
				userOrderService,
				userOrderQueryService,
				userBalanceQueryService,
				userBalanceService,
				userBalanceWithDrawnQueryService,
				tokenService,
				cookieService,
				l,
			)

			req := httptest.NewRequest(tc.on.method, "/api/user/login", strings.NewReader(tc.on.body))
			for k, v := range tc.on.header {
				req.Header.Set(k, v)
			}

			rr := httptest.NewRecorder()

			srv.LoginUser(rr, req)

			require.Equal(t, tc.want.code, rr.Code)

			for k, v := range tc.want.header {
				assert.Equal(t, v, rr.Header().Get(k))
			}

			resp := rr.Result()
			cookies := resp.Cookies()
			for k, v := range tc.want.cookie {
				for _, c := range cookies {
					if c.Name == k {
						assert.Equal(t, v, c.Value)
					}
				}
			}

			if tc.want.body != nil {
				assert.Equal(t, *tc.want.body, rr.Body.String())
			}
		})
	}
}

//nolint:dupl
func TestServer_UploadUserOrders(t *testing.T) {
	type on struct {
		method   string
		body     string
		bodyErr  error
		header   map[string]string
		ctxToken *auth.Token
	}
	type when struct {
		uploadErr error
	}
	type want struct {
		code int
	}
	testCases := []struct {
		name string
		on   on
		when when
		want want
	}{
		{
			"error on read body",
			on{
				method:  http.MethodPost,
				bodyErr: errors.New("error 1"),
				header:  map[string]string{"Content-Type": "text/plain"},
			},
			when{},
			want{code: http.StatusBadRequest},
		},
		{
			"no context token",
			on{
				method:   http.MethodPost,
				body:     `12345678903`,
				header:   map[string]string{"Content-Type": "application/json"},
				ctxToken: nil,
			},
			when{},
			want{code: http.StatusUnauthorized},
		},
		{
			"no user in context token",
			on{
				method:   http.MethodPost,
				body:     `12345678903`,
				header:   map[string]string{"Content-Type": "application/json"},
				ctxToken: &auth.Token{},
			},
			when{},
			want{code: http.StatusUnauthorized},
		},
		{
			"invalid user id in context token",
			on{
				method:   http.MethodPost,
				body:     `12345678903`,
				header:   map[string]string{"Content-Type": "application/json"},
				ctxToken: &auth.Token{UserID: "123"},
			},
			when{},
			want{code: http.StatusUnauthorized},
		},
		{
			"error invalid order id on upload",
			on{
				method:   http.MethodPost,
				body:     `12345678903`,
				header:   map[string]string{"Content-Type": "application/json"},
				ctxToken: &auth.Token{UserID: uuid.New().String()},
			},
			when{uploadErr: ErrInvalidOrderID},
			want{code: http.StatusUnprocessableEntity},
		},
		{
			"error order id already uploaded by other user on upload",
			on{
				method:   http.MethodPost,
				body:     `12345678903`,
				header:   map[string]string{"Content-Type": "application/json"},
				ctxToken: &auth.Token{UserID: uuid.New().String()},
			},
			when{uploadErr: ErrUserOrderAlreadyUploadedByOtherUser},
			want{code: http.StatusConflict},
		},
		{
			"error order id already uploaded by this user on upload",
			on{
				method:   http.MethodPost,
				body:     `12345678903`,
				header:   map[string]string{"Content-Type": "application/json"},
				ctxToken: &auth.Token{UserID: uuid.New().String()},
			},
			when{uploadErr: ErrUserOrderAlreadyUploadedByUser},
			want{code: http.StatusOK},
		},
		{
			"unexpected error on upload",
			on{
				method:   http.MethodPost,
				body:     `12345678903`,
				header:   map[string]string{"Content-Type": "application/json"},
				ctxToken: &auth.Token{UserID: uuid.New().String()},
			},
			when{uploadErr: errors.New("error 1")},
			want{code: http.StatusInternalServerError},
		},
		{
			"success upload",
			on{
				method:   http.MethodPost,
				body:     `12345678903`,
				header:   map[string]string{"Content-Type": "application/json"},
				ctxToken: &auth.Token{UserID: uuid.New().String()},
			},
			when{},
			want{code: http.StatusAccepted},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userService := NewMockUserService(t)
			userOrderService := NewMockUserOrderService(t)
			userOrderService.EXPECT().Upload(mock.Anything, mock.Anything).Return(model.UserOrder{}, tc.when.uploadErr).Maybe()

			userOrderQueryService := NewMockUserOrderQueryService(t)
			userBalanceQueryService := NewMockUserBalanceQueryService(t)
			userBalanceService := NewMockUserBalanceService(t)
			userBalanceWithDrawnQueryService := NewMockUserBalanceWithDrawnQueryService(t)
			tokenService := NewMockTokenService(t)
			cookieService := NewMockCookieService(t)
			l := ilogger.NewNullLogger()

			srv := NewServer(
				userService,
				userOrderService,
				userOrderQueryService,
				userBalanceQueryService,
				userBalanceService,
				userBalanceWithDrawnQueryService,
				tokenService,
				cookieService,
				l,
			)

			var body io.Reader
			if tc.on.bodyErr != nil {
				body = &testErrorReader{tc.on.bodyErr}
			} else {
				body = strings.NewReader(tc.on.body)
			}
			req := httptest.NewRequest(tc.on.method, "/api/user/orders", body)
			for k, v := range tc.on.header {
				req.Header.Set(k, v)
			}
			if tc.on.ctxToken != nil {
				ctx := auth.CreateTokenContext(req.Context(), *tc.on.ctxToken)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			srv.UploadUserOrders(rr, req)

			require.Equal(t, tc.want.code, rr.Code)
		})
	}
}

//nolint:dupl
func TestServer_GetUserOrders(t *testing.T) {
	userID1 := uuid.New()
	invalidUserOrderStatus := model.UserOrderStatus(-1)
	orderItemID1 := uuid.New()
	orderItemID2 := uuid.New()
	time1 := time.Now().Add(-1 * time.Hour).Truncate(time.Second)
	time2 := time.Now().Add(-1 * time.Minute).Truncate(time.Second)
	orderID1 := "12345678903"
	orderID2 := "2377225624"
	accrual1 := decimal.NewFromFloat(10.5)

	type on struct {
		method   string
		ctxToken *auth.Token
	}
	type when struct {
		findOrders    []model.UserOrder
		findOrdersErr error
	}
	type want struct {
		code   int
		header map[string]string
		body   string
	}
	testCases := []struct {
		name string
		on   on
		when when
		want want
	}{
		{
			"no context token",
			on{method: http.MethodGet, ctxToken: nil},
			when{},
			want{code: http.StatusUnauthorized},
		},
		{
			"no user in context token",
			on{method: http.MethodGet, ctxToken: &auth.Token{}},
			when{},
			want{code: http.StatusUnauthorized},
		},
		{
			"invalid user id in context token",
			on{method: http.MethodGet, ctxToken: &auth.Token{UserID: "123"}},
			when{},
			want{code: http.StatusUnauthorized},
		},
		{
			"unexpected error on find user orders",
			on{method: http.MethodGet, ctxToken: &auth.Token{UserID: uuid.New().String()}},
			when{findOrdersErr: errors.New("error 1")},
			want{code: http.StatusInternalServerError},
		},
		{
			"empty find user orders",
			on{method: http.MethodGet, ctxToken: &auth.Token{UserID: uuid.New().String()}},
			when{findOrders: []model.UserOrder{}},
			want{code: http.StatusNoContent, header: map[string]string{"Content-Type": "application/json"}},
		},
		{
			"success find user orders",
			on{method: http.MethodGet, ctxToken: &auth.Token{UserID: userID1.String()}},
			when{findOrders: []model.UserOrder{
				{ID: orderItemID1, UserID: userID1, OrderID: orderID1, Status: model.UserOrderStatusProcessed, Accrual: &accrual1, CreatedAt: time1, UpdatedAt: time2},
				{ID: orderItemID2, UserID: userID1, OrderID: orderID2, Status: model.UserOrderStatusNew, Accrual: nil, CreatedAt: time2, UpdatedAt: time2},
			}},
			want{
				code:   http.StatusOK,
				header: map[string]string{"Content-Type": "application/json"},
				body:   fmt.Sprintf(`[{"number": %q, "status": "PROCESSED", "accrual": %s, "uploaded_at": %q}, {"number": %q, "status": "NEW", "uploaded_at": %q}]`, orderID1, accrual1.String(), time1.Format(time.RFC3339), orderID2, time2.Format(time.RFC3339)),
			},
		},
		{
			"error on convert find user orders",
			on{method: http.MethodGet, ctxToken: &auth.Token{UserID: userID1.String()}},
			when{findOrders: []model.UserOrder{
				{ID: orderItemID1, UserID: userID1, OrderID: orderID1, Status: invalidUserOrderStatus, Accrual: nil, CreatedAt: time1, UpdatedAt: time2},
			}},
			want{code: http.StatusInternalServerError},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userService := NewMockUserService(t)
			userOrderService := NewMockUserOrderService(t)
			userOrderQueryService := NewMockUserOrderQueryService(t)
			userOrderQueryService.EXPECT().FindByUserID(mock.Anything, mock.Anything).Return(tc.when.findOrders, tc.when.findOrdersErr).Maybe()

			userBalanceQueryService := NewMockUserBalanceQueryService(t)
			userBalanceService := NewMockUserBalanceService(t)
			userBalanceWithDrawnQueryService := NewMockUserBalanceWithDrawnQueryService(t)
			tokenService := NewMockTokenService(t)
			cookieService := NewMockCookieService(t)
			l := ilogger.NewNullLogger()

			srv := NewServer(
				userService,
				userOrderService,
				userOrderQueryService,
				userBalanceQueryService,
				userBalanceService,
				userBalanceWithDrawnQueryService,
				tokenService,
				cookieService,
				l,
			)

			req := httptest.NewRequest(tc.on.method, "/api/user/orders", http.NoBody)
			if tc.on.ctxToken != nil {
				ctx := auth.CreateTokenContext(req.Context(), *tc.on.ctxToken)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			srv.GetUserOrders(rr, req)

			require.Equal(t, tc.want.code, rr.Code)

			for k, v := range tc.want.header {
				assert.Equal(t, v, rr.Header().Get(k))
			}

			if tc.want.body != "" {
				assert.JSONEq(t, tc.want.body, rr.Body.String())
			}
		})
	}
}

//nolint:dupl
func TestServer_GetUserBalance(t *testing.T) {
	userID1 := uuid.New()

	type on struct {
		method   string
		ctxToken *auth.Token
	}
	type when struct {
		getBalance    model.UserBalance
		getBalanceErr error
	}
	type want struct {
		code   int
		header map[string]string
		body   string
	}
	testCases := []struct {
		name string
		on   on
		when when
		want want
	}{
		{
			"no context token",
			on{method: http.MethodGet, ctxToken: nil},
			when{},
			want{code: http.StatusUnauthorized},
		},
		{
			"no user in context token",
			on{method: http.MethodGet, ctxToken: &auth.Token{}},
			when{},
			want{code: http.StatusUnauthorized},
		},
		{
			"invalid user id in context token",
			on{method: http.MethodGet, ctxToken: &auth.Token{UserID: "123"}},
			when{},
			want{code: http.StatusUnauthorized},
		},
		{
			"unexpected error on get user balance",
			on{method: http.MethodGet, ctxToken: &auth.Token{UserID: uuid.New().String()}},
			when{getBalanceErr: errors.New("error 1")},
			want{code: http.StatusInternalServerError},
		},
		{
			"success get user balance",
			on{method: http.MethodGet, ctxToken: &auth.Token{UserID: userID1.String()}},
			when{getBalance: model.UserBalance{UserID: userID1, Balance: decimal.NewFromFloat(10.5), WithdrawnSum: decimal.NewFromFloat(100.2)}},
			want{code: http.StatusOK, header: map[string]string{"Content-Type": "application/json"}, body: `{"current": 10.5, "withdrawn": 100.2}`},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userService := NewMockUserService(t)
			userOrderService := NewMockUserOrderService(t)
			userOrderQueryService := NewMockUserOrderQueryService(t)
			userBalanceQueryService := NewMockUserBalanceQueryService(t)
			userBalanceQueryService.EXPECT().GetByUserID(mock.Anything, mock.Anything).Return(tc.when.getBalance, tc.when.getBalanceErr).Maybe()

			userBalanceService := NewMockUserBalanceService(t)
			userBalanceWithDrawnQueryService := NewMockUserBalanceWithDrawnQueryService(t)
			tokenService := NewMockTokenService(t)
			cookieService := NewMockCookieService(t)
			l := ilogger.NewNullLogger()

			srv := NewServer(
				userService,
				userOrderService,
				userOrderQueryService,
				userBalanceQueryService,
				userBalanceService,
				userBalanceWithDrawnQueryService,
				tokenService,
				cookieService,
				l,
			)

			req := httptest.NewRequest(tc.on.method, "/api/user/balance", http.NoBody)
			if tc.on.ctxToken != nil {
				ctx := auth.CreateTokenContext(req.Context(), *tc.on.ctxToken)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			srv.GetUserBalance(rr, req)

			require.Equal(t, tc.want.code, rr.Code)

			for k, v := range tc.want.header {
				assert.Equal(t, v, rr.Header().Get(k))
			}

			if tc.want.body != "" {
				assert.JSONEq(t, tc.want.body, rr.Body.String())
			}
		})
	}
}

//nolint:dupl
func TestServer_WithdrawUserBalance(t *testing.T) {
	userID1 := uuid.New()

	type on struct {
		method   string
		body     string
		bodyErr  error
		ctxToken *auth.Token
	}
	type when struct {
		addWithdrawnErr error
	}
	type want struct {
		code int
	}
	testCases := []struct {
		name string
		on   on
		when when
		want want
	}{
		{
			"decode error",
			on{method: http.MethodPost, bodyErr: errors.New("error 1")},
			when{},
			want{code: http.StatusBadRequest},
		},
		{
			"invalid body",
			on{method: http.MethodPost, body: `invalid body`},
			when{},
			want{code: http.StatusBadRequest},
		},
		{
			"no context token",
			on{method: http.MethodPost, ctxToken: nil, body: `{"order": "2377225624", "sum": 751}`},
			when{},
			want{code: http.StatusUnauthorized},
		},
		{
			"no user in context token",
			on{method: http.MethodPost, ctxToken: &auth.Token{}, body: `{"order": "2377225624", "sum": 751}`},
			when{},
			want{code: http.StatusUnauthorized},
		},
		{
			"invalid user id in context token",
			on{method: http.MethodPost, ctxToken: &auth.Token{UserID: "123"}, body: `{"order": "2377225624", "sum": 751}`},
			when{},
			want{code: http.StatusUnauthorized},
		},
		{
			"error invalid order id on add withdraw",
			on{method: http.MethodPost, ctxToken: &auth.Token{UserID: uuid.New().String()}, body: `{"order": "2377225624", "sum": 751}`},
			when{addWithdrawnErr: ErrInvalidOrderID},
			want{code: http.StatusUnprocessableEntity},
		},
		{
			"error user balance not enough on add withdraw",
			on{method: http.MethodPost, ctxToken: &auth.Token{UserID: uuid.New().String()}, body: `{"order": "2377225624", "sum": 751}`},
			when{addWithdrawnErr: ErrUserBalanceIsNotEnough},
			want{code: http.StatusPaymentRequired},
		},
		{
			"unexpected error on add withdraw",
			on{method: http.MethodPost, ctxToken: &auth.Token{UserID: uuid.New().String()}, body: `{"order": "2377225624", "sum": 751}`},
			when{addWithdrawnErr: errors.New("error 1")},
			want{code: http.StatusInternalServerError},
		},
		{
			"success on add withdraw",
			on{method: http.MethodPost, ctxToken: &auth.Token{UserID: userID1.String()}, body: `{"order": "2377225624", "sum": 751}`},
			when{},
			want{code: http.StatusOK},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userService := NewMockUserService(t)
			userOrderService := NewMockUserOrderService(t)
			userOrderQueryService := NewMockUserOrderQueryService(t)
			userBalanceQueryService := NewMockUserBalanceQueryService(t)
			userBalanceService := NewMockUserBalanceService(t)
			userBalanceService.EXPECT().AddWithdrawn(mock.Anything, mock.Anything).Return(tc.when.addWithdrawnErr).Maybe()

			userBalanceWithDrawnQueryService := NewMockUserBalanceWithDrawnQueryService(t)
			tokenService := NewMockTokenService(t)
			cookieService := NewMockCookieService(t)
			l := ilogger.NewNullLogger()

			srv := NewServer(
				userService,
				userOrderService,
				userOrderQueryService,
				userBalanceQueryService,
				userBalanceService,
				userBalanceWithDrawnQueryService,
				tokenService,
				cookieService,
				l,
			)

			var body io.Reader
			if tc.on.bodyErr != nil {
				body = &testErrorReader{tc.on.bodyErr}
			} else {
				body = strings.NewReader(tc.on.body)
			}
			req := httptest.NewRequest(tc.on.method, "/api/user/balance/withdraw", body)
			if tc.on.ctxToken != nil {
				ctx := auth.CreateTokenContext(req.Context(), *tc.on.ctxToken)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			srv.WithdrawUserBalance(rr, req)

			require.Equal(t, tc.want.code, rr.Code)
		})
	}
}

//nolint:dupl
func TestServer_GetUserWithdrawals(t *testing.T) {
	userID1 := uuid.New()
	itemID1 := uuid.New()
	itemID2 := uuid.New()
	time1 := time.Now().Add(-1 * time.Hour).Truncate(time.Second)
	time2 := time.Now().Add(-1 * time.Minute).Truncate(time.Second)
	orderID1 := "12345678903"
	orderID2 := "2377225624"

	type on struct {
		method   string
		ctxToken *auth.Token
	}
	type when struct {
		find    []model.UserBalanceWithdrawn
		findErr error
	}
	type want struct {
		code   int
		header map[string]string
		body   string
	}
	testCases := []struct {
		name string
		on   on
		when when
		want want
	}{
		{
			"no context token",
			on{method: http.MethodGet, ctxToken: nil},
			when{},
			want{code: http.StatusUnauthorized},
		},
		{
			"no user in context token",
			on{method: http.MethodGet, ctxToken: &auth.Token{}},
			when{},
			want{code: http.StatusUnauthorized},
		},
		{
			"invalid user id in context token",
			on{method: http.MethodGet, ctxToken: &auth.Token{UserID: "123"}},
			when{},
			want{code: http.StatusUnauthorized},
		},
		{
			"unexpected error on find withdrawn",
			on{method: http.MethodGet, ctxToken: &auth.Token{UserID: uuid.New().String()}},
			when{findErr: errors.New("error 1")},
			want{code: http.StatusInternalServerError},
		},
		{
			"empty user withdrawn",
			on{method: http.MethodGet, ctxToken: &auth.Token{UserID: userID1.String()}},
			when{find: []model.UserBalanceWithdrawn{}},
			want{code: http.StatusNoContent, header: map[string]string{"Content-Type": "application/json"}},
		},
		{
			"success find user withdrawn",
			on{method: http.MethodGet, ctxToken: &auth.Token{UserID: userID1.String()}},
			when{find: []model.UserBalanceWithdrawn{
				{ID: itemID1, UserID: userID1, OrderID: orderID1, Amount: decimal.NewFromFloat(10.5), CreatedAt: time1},
				{ID: itemID2, UserID: userID1, OrderID: orderID2, Amount: decimal.NewFromFloat(100.2), CreatedAt: time2},
			}},
			want{
				code:   http.StatusOK,
				header: map[string]string{"Content-Type": "application/json"},
				body:   fmt.Sprintf(`[{"order": %q, "sum": 10.5, "processed_at": %q}, {"order": %q, "sum": 100.2, "processed_at": %q}]`, orderID1, time1.Format(time.RFC3339), orderID2, time2.Format(time.RFC3339)),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userService := NewMockUserService(t)
			userOrderService := NewMockUserOrderService(t)
			userOrderQueryService := NewMockUserOrderQueryService(t)
			userBalanceQueryService := NewMockUserBalanceQueryService(t)
			userBalanceService := NewMockUserBalanceService(t)
			userBalanceWithDrawnQueryService := NewMockUserBalanceWithDrawnQueryService(t)
			userBalanceWithDrawnQueryService.EXPECT().FindByUserID(mock.Anything, mock.Anything).Return(tc.when.find, tc.when.findErr).Maybe()

			tokenService := NewMockTokenService(t)
			cookieService := NewMockCookieService(t)
			l := ilogger.NewNullLogger()

			srv := NewServer(
				userService,
				userOrderService,
				userOrderQueryService,
				userBalanceQueryService,
				userBalanceService,
				userBalanceWithDrawnQueryService,
				tokenService,
				cookieService,
				l,
			)

			req := httptest.NewRequest(tc.on.method, "/api/user/withdrawals", http.NoBody)
			if tc.on.ctxToken != nil {
				ctx := auth.CreateTokenContext(req.Context(), *tc.on.ctxToken)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			srv.GetUserWithdrawals(rr, req)

			require.Equal(t, tc.want.code, rr.Code)

			for k, v := range tc.want.header {
				assert.Equal(t, v, rr.Header().Get(k))
			}

			if tc.want.body != "" {
				assert.JSONEq(t, tc.want.body, rr.Body.String())
			}
		})
	}
}
