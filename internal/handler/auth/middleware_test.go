package auth

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/liebeSonne/gophermart/internal/auth"
)

func TestNewAuthMiddleware(t *testing.T) {
	defaultCode := http.StatusOK
	userID1 := uuid.New()

	type when struct {
		getTokenString string
		getTokenErr    error
		parseTokenData auth.Token
		parseTokenErr  error
	}
	type want struct {
		code          int
		contextUserID uuid.UUID
	}
	testCases := []struct {
		name string
		when when
		want want
	}{
		{
			"error on get cookies token",
			when{getTokenErr: errors.New("error")},
			want{http.StatusInternalServerError, uuid.Nil},
		},
		{
			"empty cookies token",
			when{getTokenString: ""},
			want{defaultCode, uuid.Nil},
		},
		{
			"error on parse token",
			when{getTokenString: "111", parseTokenErr: errors.New("error")},
			want{defaultCode, uuid.Nil},
		},
		{
			"context token",
			when{getTokenString: "111", parseTokenData: auth.Token{UserID: userID1.String()}},
			want{defaultCode, userID1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenService := NewMockTokenService(t)
			tokenService.EXPECT().Parse(mock.Anything).Return(tc.when.parseTokenData, tc.when.parseTokenErr).Maybe()

			cookieService := NewMockCookieService(t)
			cookieService.EXPECT().GetAuthToken(mock.Anything).Return(tc.when.getTokenString, tc.when.getTokenErr)

			contextUserID := uuid.Nil
			existContextUserID := false
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(defaultCode)
				_, err := w.Write([]byte("ok"))
				require.NoError(t, err)

				ctx := r.Context()
				contextUserID, existContextUserID = auth.GetUserIDFromContext(ctx)
			})

			l, _ := test.NewNullLogger()

			handler := NewAuthMiddleware(h, tokenService, cookieService, l)

			srv := httptest.NewServer(handler)
			defer srv.Close()

			client := resty.New()

			req := client.R()
			req.Method = http.MethodGet
			req.URL = srv.URL + "/"
			req.SetDoNotParseResponse(true)

			resp, err := req.Send()
			require.NoError(t, err)

			require.Equal(t, tc.want.code, resp.StatusCode(), fmt.Sprintf("expected status code %d but got %d with body: %s", tc.want.code, resp.StatusCode(), string(resp.Body())))

			require.Equal(t, tc.want.contextUserID != uuid.Nil, existContextUserID)
			if tc.want.contextUserID != uuid.Nil {
				require.Equal(t, tc.want.contextUserID, contextUserID)
			}
		})
	}
}
