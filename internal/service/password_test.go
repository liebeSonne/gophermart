package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordService_CreateHash(t *testing.T) {
	password1 := "password 1"
	password2 := "password 2"
	secretKey1 := []byte("secret1")
	secretKey2 := []byte("secret2")

	type when struct {
		secretKey1 []byte
		secretKey2 []byte
	}
	type on struct {
		password1 string
		password2 string
	}
	type want struct {
		err11 error
		err12 error
		err21 error
		err22 error
	}
	testCases := []struct {
		name string
		on   on
		when when
		want want
	}{
		{
			"equal secret key and not equal passwords",
			on{password1, password2},
			when{secretKey1, secretKey1},
			want{nil, nil, nil, nil},
		},
		{
			"not equal secret key and not equal passwords",
			on{password1, password2},
			when{secretKey1, secretKey2},
			want{nil, nil, nil, nil},
		},
		{
			"equal secret key and equal passwords",
			on{password1, password1},
			when{secretKey1, secretKey1},
			want{nil, nil, nil, nil},
		},
		{
			"not equal secret key and equal passwords",
			on{password1, password1},
			when{secretKey1, secretKey2},
			want{nil, nil, nil, nil},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s1 := NewPasswordService(tc.when.secretKey1)
			s2 := NewPasswordService(tc.when.secretKey2)

			passHash11, err11 := s1.CreateHash(t.Context(), tc.on.password1)
			passHash12, err12 := s1.CreateHash(t.Context(), tc.on.password2)

			passHash21, err21 := s2.CreateHash(t.Context(), tc.on.password1)
			passHash22, err22 := s2.CreateHash(t.Context(), tc.on.password2)

			asserPassHashErr(t, tc.want.err11, err11)
			asserPassHashErr(t, tc.want.err12, err12)
			asserPassHashErr(t, tc.want.err21, err21)
			asserPassHashErr(t, tc.want.err22, err22)

			assertPassHash(t, tc.on.password1, tc.on.password2, passHash11, passHash12, err11, err12)
			assertPassHash(t, tc.on.password1, tc.on.password2, passHash21, passHash22, err21, err22)
		})
	}
}

func asserPassHashErr(t *testing.T, expect, actual error) {
	if expect != nil {
		require.Error(t, actual)
		assert.ErrorContains(t, actual, expect.Error())
	} else {
		assert.NoError(t, actual)
	}
}

func assertPassHash(t *testing.T, pass1, pass2, hash1, hash2 string, err1, err2 error) {
	if err1 == nil && err2 == nil && pass1 == pass2 {
		assert.Equal(t, hash1, hash2)
	}
	if err1 == nil && err2 == nil && pass1 != pass2 {
		assert.NotEqual(t, hash1, hash2)
	}
}
