package cookie

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCookieService_SetAuthToken(t *testing.T) {
	tokenKey := "token_key"

	type on struct {
		tokenString string
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
			"not empty",
			on{"123"},
			want{nil},
		},
		{
			"empty",
			on{""},
			want{nil},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)

			cookieService := NewService(tokenKey)

			err := cookieService.SetAuthToken(tc.on.tokenString, w, r)

			if tc.want.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.want.err)
				return
			}

			require.NoError(t, err)

			res := w.Result()
			defer res.Body.Close()

			cookies := res.Cookies()
			require.Len(t, cookies, 1)
			require.Equal(t, tokenKey, cookies[0].Name)
			require.Equal(t, tc.on.tokenString, cookies[0].Value)

			cookies = r.Cookies()
			require.Len(t, cookies, 1)
			require.Equal(t, tokenKey, cookies[0].Name)
			require.Equal(t, tc.on.tokenString, cookies[0].Value)
		})
	}
}

func TestCookieService_GetAuthToken(t *testing.T) {
	tokenKey := "token_key"

	type when struct {
		cookies []http.Cookie
	}
	type want struct {
		err         error
		tokenString string
	}
	testCases := []struct {
		name string
		when when
		want want
	}{
		{
			"not exist cookies",
			when{[]http.Cookie{}},
			want{nil, ""},
		},
		{
			"empty",
			when{[]http.Cookie{{Name: tokenKey, Value: ""}}},
			want{nil, ""},
		},
		{
			"one cookies",
			when{[]http.Cookie{{Name: tokenKey, Value: "123"}}},
			want{nil, "123"},
		},
		{
			"no one cookies",
			when{[]http.Cookie{{Name: "else_key", Value: "333"}}},
			want{nil, ""},
		},
		{
			"many cookies",
			when{[]http.Cookie{{Name: tokenKey, Value: "123"}, {Name: "else_key", Value: "333"}}},
			want{nil, "123"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			for _, cookie := range tc.when.cookies {
				r.AddCookie(&cookie)
			}

			cookieService := NewService(tokenKey)

			tokenString, err := cookieService.GetAuthToken(r)

			if tc.want.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.want.err)
				return
			}

			require.NoError(t, err)

			require.Equal(t, tc.want.tokenString, tokenString)
		})
	}
}
