package config

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	appID := "app"
	address1 := "127.0.0.1:1111"
	address2 := "127.0.0.2:2222"
	address3 := "127.0.0.3:3333"
	address4 := "127.0.0.4:4444"
	uri1 := "postgres://127.0.0.1:1111/database"
	uri2 := "postgres://127.0.0.1:2222/database"
	logLevel1 := LogLevelError
	logLevel2 := LogLevelWarn
	logFormat1 := LogFormatJSON
	logFormat2 := LogFormatText
	passwordSecretKey1 := "password secret key 1"
	authCookieTokenKey1 := "token_key_1"
	authSecretKey1 := "auth secret key 1"
	authTokenExpiresStr1 := "10m"
	authTokenExpiresDuration1 := time.Minute * 10

	type when struct {
		args []string
		envs map[string]string
	}
	type want struct {
		cfg Config
		err error
	}
	testCases := []struct {
		name string
		when when
		want want
	}{
		{
			"default",
			when{},
			want{
				Config{
					RunAddress:           DefaultRunAddress,
					DatabaseURI:          DefaultDatabaseURI,
					AccrualSystemAddress: DefaultAccrualSystemAddress,
					LogLevel:             DefaultLogLever,
					LogFormat:            DefaultLogFormat,
					PasswordSecretKey:    DefaultPasswordSecretKey,
					AuthCookieTokenKey:   DefaultAuthCookieTokenKey,
					AuthSecretKey:        DefaultAuthSecretKey,
					AuthTokenExpires:     DefaultAuthTokenExpires,
				},
				nil,
			},
		},
		{
			"from env",
			when{
				[]string{},
				map[string]string{
					RunAddressEnv:           address1,
					DatabaseURIEnv:          uri1,
					AccrualSystemAddressEnv: address2,
					LogLevelEnv:             string(logLevel1),
					LogFormatEnv:            string(logFormat1),
					PasswordSecretKeyEnv:    passwordSecretKey1,
					AuthCookieTokenKeyEnv:   authCookieTokenKey1,
					AuthSecretKeyEnv:        authSecretKey1,
					AuthTokenExpiresEnv:     authTokenExpiresStr1,
				},
			},
			want{
				Config{
					RunAddress:           address1,
					DatabaseURI:          uri1,
					AccrualSystemAddress: address2,
					LogLevel:             logLevel1,
					LogFormat:            logFormat1,
					PasswordSecretKey:    passwordSecretKey1,
					AuthCookieTokenKey:   authSecretKey1,
					AuthSecretKey:        authCookieTokenKey1,
					AuthTokenExpires:     authTokenExpiresDuration1,
				},
				nil,
			},
		},
		{
			"from env run address",
			when{
				[]string{},
				map[string]string{
					RunAddressEnv: address1,
				},
			},
			want{
				Config{
					RunAddress:           address1,
					DatabaseURI:          DefaultDatabaseURI,
					AccrualSystemAddress: DefaultAccrualSystemAddress,
					LogLevel:             DefaultLogLever,
					LogFormat:            DefaultLogFormat,
					PasswordSecretKey:    DefaultPasswordSecretKey,
					AuthCookieTokenKey:   DefaultAuthCookieTokenKey,
					AuthSecretKey:        DefaultAuthSecretKey,
					AuthTokenExpires:     DefaultAuthTokenExpires,
				},
				nil,
			},
		},
		{
			"from env database uri",
			when{
				[]string{},
				map[string]string{
					DatabaseURIEnv: uri1,
				},
			},
			want{
				Config{
					RunAddress:           DefaultRunAddress,
					DatabaseURI:          uri1,
					AccrualSystemAddress: DefaultAccrualSystemAddress,
					LogLevel:             DefaultLogLever,
					LogFormat:            DefaultLogFormat,
					PasswordSecretKey:    DefaultPasswordSecretKey,
					AuthCookieTokenKey:   DefaultAuthCookieTokenKey,
					AuthSecretKey:        DefaultAuthSecretKey,
					AuthTokenExpires:     DefaultAuthTokenExpires,
				},
				nil,
			},
		},
		{
			"from env accrual system address",
			when{
				[]string{},
				map[string]string{
					AccrualSystemAddressEnv: address2,
				},
			},
			want{
				Config{
					RunAddress:           DefaultRunAddress,
					DatabaseURI:          DefaultDatabaseURI,
					AccrualSystemAddress: address2,
					LogLevel:             DefaultLogLever,
					LogFormat:            DefaultLogFormat,
					PasswordSecretKey:    DefaultPasswordSecretKey,
					AuthCookieTokenKey:   DefaultAuthCookieTokenKey,
					AuthSecretKey:        DefaultAuthSecretKey,
					AuthTokenExpires:     DefaultAuthTokenExpires,
				},
				nil,
			},
		},
		{
			"from env log level",
			when{
				[]string{},
				map[string]string{
					LogLevelEnv: string(logLevel1),
				},
			},
			want{
				Config{
					RunAddress:           DefaultRunAddress,
					DatabaseURI:          DefaultDatabaseURI,
					AccrualSystemAddress: DefaultAccrualSystemAddress,
					LogLevel:             logLevel1,
					LogFormat:            DefaultLogFormat,
					PasswordSecretKey:    DefaultPasswordSecretKey,
					AuthCookieTokenKey:   DefaultAuthCookieTokenKey,
					AuthSecretKey:        DefaultAuthSecretKey,
					AuthTokenExpires:     DefaultAuthTokenExpires,
				},
				nil,
			},
		},
		{
			"from env log format",
			when{
				[]string{},
				map[string]string{
					LogFormatEnv: string(logFormat1),
				},
			},
			want{
				Config{
					RunAddress:           DefaultRunAddress,
					DatabaseURI:          DefaultDatabaseURI,
					AccrualSystemAddress: DefaultAccrualSystemAddress,
					LogLevel:             DefaultLogLever,
					LogFormat:            logFormat1,
					PasswordSecretKey:    DefaultPasswordSecretKey,
					AuthCookieTokenKey:   DefaultAuthCookieTokenKey,
					AuthSecretKey:        DefaultAuthSecretKey,
					AuthTokenExpires:     DefaultAuthTokenExpires,
				},
				nil,
			},
		},
		{
			"from env password secret key",
			when{
				[]string{},
				map[string]string{
					PasswordSecretKeyEnv: passwordSecretKey1,
				},
			},
			want{
				Config{
					RunAddress:           DefaultRunAddress,
					DatabaseURI:          DefaultDatabaseURI,
					AccrualSystemAddress: DefaultAccrualSystemAddress,
					LogLevel:             DefaultLogLever,
					LogFormat:            DefaultLogFormat,
					PasswordSecretKey:    passwordSecretKey1,
					AuthCookieTokenKey:   DefaultAuthCookieTokenKey,
					AuthSecretKey:        DefaultAuthSecretKey,
					AuthTokenExpires:     DefaultAuthTokenExpires,
				},
				nil,
			},
		},
		{
			"from env auth cookie token key",
			when{
				[]string{},
				map[string]string{
					PasswordSecretKeyEnv: authCookieTokenKey1,
				},
			},
			want{
				Config{
					RunAddress:           DefaultRunAddress,
					DatabaseURI:          DefaultDatabaseURI,
					AccrualSystemAddress: DefaultAccrualSystemAddress,
					LogLevel:             DefaultLogLever,
					LogFormat:            DefaultLogFormat,
					PasswordSecretKey:    DefaultPasswordSecretKey,
					AuthCookieTokenKey:   authCookieTokenKey1,
					AuthSecretKey:        DefaultAuthSecretKey,
					AuthTokenExpires:     DefaultAuthTokenExpires,
				},
				nil,
			},
		},
		{
			"from env auth cookie token key",
			when{
				[]string{},
				map[string]string{
					AuthSecretKeyEnv: authSecretKey1,
				},
			},
			want{
				Config{
					RunAddress:           DefaultRunAddress,
					DatabaseURI:          DefaultDatabaseURI,
					AccrualSystemAddress: DefaultAccrualSystemAddress,
					LogLevel:             DefaultLogLever,
					LogFormat:            DefaultLogFormat,
					PasswordSecretKey:    DefaultPasswordSecretKey,
					AuthCookieTokenKey:   DefaultAuthCookieTokenKey,
					AuthSecretKey:        authSecretKey1,
					AuthTokenExpires:     DefaultAuthTokenExpires,
				},
				nil,
			},
		},
		{
			"from env auth cookie token key",
			when{
				[]string{},
				map[string]string{
					AuthTokenExpiresEnv: authTokenExpiresStr1,
				},
			},
			want{
				Config{
					RunAddress:           DefaultRunAddress,
					DatabaseURI:          DefaultDatabaseURI,
					AccrualSystemAddress: DefaultAccrualSystemAddress,
					LogLevel:             DefaultLogLever,
					LogFormat:            DefaultLogFormat,
					PasswordSecretKey:    DefaultPasswordSecretKey,
					AuthCookieTokenKey:   DefaultAuthCookieTokenKey,
					AuthSecretKey:        DefaultAuthSecretKey,
					AuthTokenExpires:     authTokenExpiresDuration1,
				},
				nil,
			},
		},
		{
			"from flag",
			when{
				[]string{
					"-a", address1,
					"-d", uri1,
					"-r", address2,
					"-ll", string(logLevel1),
					"-lf", string(logFormat1),
				},
				map[string]string{},
			},
			want{
				Config{
					RunAddress:           address1,
					DatabaseURI:          uri1,
					AccrualSystemAddress: address2,
					LogLevel:             logLevel1,
					LogFormat:            logFormat1,
					PasswordSecretKey:    DefaultPasswordSecretKey,
					AuthCookieTokenKey:   DefaultAuthCookieTokenKey,
					AuthSecretKey:        DefaultAuthSecretKey,
					AuthTokenExpires:     DefaultAuthTokenExpires,
				},
				nil,
			},
		},
		{
			"from flag -a",
			when{
				[]string{
					"-a", address1,
				},
				map[string]string{},
			},
			want{
				Config{
					RunAddress:           address1,
					DatabaseURI:          DefaultDatabaseURI,
					AccrualSystemAddress: DefaultAccrualSystemAddress,
					LogLevel:             DefaultLogLever,
					LogFormat:            DefaultLogFormat,
					PasswordSecretKey:    DefaultPasswordSecretKey,
					AuthCookieTokenKey:   DefaultAuthCookieTokenKey,
					AuthSecretKey:        DefaultAuthSecretKey,
					AuthTokenExpires:     DefaultAuthTokenExpires,
				},
				nil,
			},
		},
		{
			"from flag -d",
			when{
				[]string{
					"-d", uri1,
				},
				map[string]string{},
			},
			want{
				Config{
					RunAddress:           DefaultRunAddress,
					DatabaseURI:          uri1,
					AccrualSystemAddress: DefaultAccrualSystemAddress,
					LogLevel:             DefaultLogLever,
					LogFormat:            DefaultLogFormat,
					PasswordSecretKey:    DefaultPasswordSecretKey,
					AuthCookieTokenKey:   DefaultAuthCookieTokenKey,
					AuthSecretKey:        DefaultAuthSecretKey,
					AuthTokenExpires:     DefaultAuthTokenExpires,
				},
				nil,
			},
		},
		{
			"from flag -r",
			when{
				[]string{
					"-r", address2,
				},
				map[string]string{},
			},
			want{
				Config{
					RunAddress:           DefaultRunAddress,
					DatabaseURI:          DefaultDatabaseURI,
					AccrualSystemAddress: address2,
					LogLevel:             DefaultLogLever,
					LogFormat:            DefaultLogFormat,
					PasswordSecretKey:    DefaultPasswordSecretKey,
					AuthCookieTokenKey:   DefaultAuthCookieTokenKey,
					AuthSecretKey:        DefaultAuthSecretKey,
					AuthTokenExpires:     DefaultAuthTokenExpires,
				},
				nil,
			},
		},
		{
			"from flag -ll",
			when{
				[]string{
					"-ll", string(logLevel1),
				},
				map[string]string{},
			},
			want{
				Config{
					RunAddress:           DefaultRunAddress,
					DatabaseURI:          DefaultDatabaseURI,
					AccrualSystemAddress: DefaultAccrualSystemAddress,
					LogLevel:             logLevel1,
					LogFormat:            DefaultLogFormat,
					PasswordSecretKey:    DefaultPasswordSecretKey,
					AuthCookieTokenKey:   DefaultAuthCookieTokenKey,
					AuthSecretKey:        DefaultAuthSecretKey,
					AuthTokenExpires:     DefaultAuthTokenExpires,
				},
				nil,
			},
		},
		{
			"from flag -lf",
			when{
				[]string{
					"-lf", string(logFormat1),
				},
				map[string]string{},
			},
			want{
				Config{
					RunAddress:           DefaultRunAddress,
					DatabaseURI:          DefaultDatabaseURI,
					AccrualSystemAddress: DefaultAccrualSystemAddress,
					LogLevel:             DefaultLogLever,
					LogFormat:            logFormat1,
					PasswordSecretKey:    DefaultPasswordSecretKey,
					AuthCookieTokenKey:   DefaultAuthCookieTokenKey,
					AuthSecretKey:        DefaultAuthSecretKey,
					AuthTokenExpires:     DefaultAuthTokenExpires,
				},
				nil,
			},
		},
		{
			"from env and flag",
			when{
				[]string{
					"-a", address1,
					"-d", uri1,
					"-r", address2,
					"-ll", string(logLevel1),
					"-lf", string(logFormat1),
				},
				map[string]string{
					RunAddressEnv:           address3,
					DatabaseURIEnv:          uri2,
					AccrualSystemAddressEnv: address4,
					LogLevelEnv:             string(logLevel2),
					LogFormatEnv:            string(logFormat2),
					PasswordSecretKeyEnv:    passwordSecretKey1,
					AuthCookieTokenKeyEnv:   authCookieTokenKey1,
					AuthSecretKeyEnv:        authSecretKey1,
					AuthTokenExpiresEnv:     authTokenExpiresStr1,
				},
			},
			want{
				Config{
					RunAddress:           address1,
					DatabaseURI:          uri1,
					AccrualSystemAddress: address2,
					LogLevel:             logLevel1,
					LogFormat:            logFormat1,
					PasswordSecretKey:    passwordSecretKey1,
					AuthCookieTokenKey:   authCookieTokenKey1,
					AuthSecretKey:        authSecretKey1,
					AuthTokenExpires:     authTokenExpiresDuration1,
				},
				nil,
			},
		},
		{
			"not all from env and flag",
			when{
				[]string{
					"-a", address1,
				},
				map[string]string{
					RunAddressEnv:           address3,
					DatabaseURIEnv:          uri2,
					AccrualSystemAddressEnv: address4,
					LogLevelEnv:             string(logLevel1),
					LogFormatEnv:            string(logFormat1),
					PasswordSecretKeyEnv:    passwordSecretKey1,
					AuthCookieTokenKeyEnv:   authCookieTokenKey1,
					AuthSecretKeyEnv:        authSecretKey1,
					AuthTokenExpiresEnv:     authTokenExpiresStr1,
				},
			},
			want{
				Config{
					RunAddress:           address1,
					DatabaseURI:          uri2,
					AccrualSystemAddress: address4,
					LogLevel:             logLevel1,
					LogFormat:            logFormat1,
					PasswordSecretKey:    passwordSecretKey1,
					AuthCookieTokenKey:   authCookieTokenKey1,
					AuthSecretKey:        authSecretKey1,
					AuthTokenExpires:     authTokenExpiresDuration1,
				},
				nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			oldArgs := os.Args
			oldEnv := os.Environ()

			args := make([]string, 0, len(tc.when.args)+1)
			args = append(args, "")
			args = append(args, tc.when.args...)
			os.Args = args

			os.Clearenv()
			for k, v := range tc.when.envs {
				t.Setenv(k, v)
			}
			t.Cleanup(func() {
				os.Args = oldArgs
				os.Clearenv()
				for _, pair := range oldEnv {
					kv := strings.SplitN(pair, "=", 2)
					_ = os.Setenv(kv[0], kv[1])
				}
			})

			cfg, err := LoadConfig(appID)

			if tc.want.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.want.err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want.cfg, cfg)
		})
	}
}
