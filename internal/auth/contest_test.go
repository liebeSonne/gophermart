package auth

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetUserIDFromContext(t *testing.T) {
	userID := uuid.New()

	type on struct {
		ctxToken *Token
	}
	type want struct {
		userID uuid.UUID
		ok     bool
	}
	testCases := []struct {
		name string
		on   on
		want want
	}{
		{
			"no context token",
			on{ctxToken: nil},
			want{uuid.UUID{}, false},
		},
		{
			"no user in context token",
			on{ctxToken: &Token{}},
			want{uuid.UUID{}, false},
		},
		{
			"invalid user id in context token",
			on{ctxToken: &Token{UserID: "123"}},
			want{uuid.UUID{}, false},
		},
		{
			"success user id from context",
			on{ctxToken: &Token{UserID: userID.String()}},
			want{userID, true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			if tc.on.ctxToken != nil {
				ctx = CreateTokenContext(ctx, *tc.on.ctxToken)
			}

			id, ok := GetUserIDFromContext(ctx)

			assert.Equal(t, tc.want.userID, id)
			assert.Equal(t, tc.want.ok, ok)
		})
	}
}
