package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTokenServiceImpl_CreateAndParse(t *testing.T) {
	secretKey := "secret123"
	tokenExpiry := time.Hour * 3
	userID1 := "123"

	type on struct {
		token Token
	}
	type want struct {
		createErr error
		parseErr  error
	}
	testCases := []struct {
		name string
		on   on
		want want
	}{
		{
			"create jwt token with user",
			on{Token{UserID: userID1}},
			want{nil, nil},
		},
		{
			"create jwt token with empty user",
			on{Token{UserID: ""}},
			want{nil, nil},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenService := NewTokenService(secretKey, tokenExpiry)

			tokenString, err := tokenService.Create(tc.on.token)

			if tc.want.createErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.want.createErr)
				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, tokenString)

			tokenData, err := tokenService.Parse(tokenString)

			if tc.want.parseErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.want.parseErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.on.token.UserID, tokenData.UserID)
		})
	}
}
