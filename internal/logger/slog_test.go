package logger

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/liebeSonne/gophermart/internal/config"
)

func TestNewSlogLogger(t *testing.T) {
	testF := func(logger Logger) {
		logger.Debugf("msg-d1: %v, %v", "arg-d1-1", "arg-d1-2")
		logger.Infof("msg-i1: %v, %v", "arg-i1-1", "arg-i1-2")
		logger.Warnf("msg-w1: %v, %v", "arg-w1-1", "arg-w1-2")
		logger.Errorf("msg-e1: %v, %v", "arg-e1-1", "arg-e1-2")

		logger.Debugw("msg-d2", "arg-d2-1", "arg-d2-2")
		logger.Infow("msg-i2", "arg-i2-1", "arg-i2-2")
		logger.Warnw("msg-w2", "arg-w2-1", "arg-w2-2")
		logger.Errorw("msg-e2", "arg-e2-1", "arg-e2-2")
	}
	debugLevelValue := "DEBUG"
	infoLevelValue := "INFO"
	warnLevelValue := "WARN"
	errorLevelValue := "ERROR"

	messageKey := slog.MessageKey
	levelKey := slog.LevelKey

	debugFMsg := fmt.Sprintf(`{%q: %q, %q: "msg-d1: arg-d1-1, arg-d1-2"}`, levelKey, debugLevelValue, messageKey)
	infoFMsg := fmt.Sprintf(`{%q: %q, %q: "msg-i1: arg-i1-1, arg-i1-2"}`, levelKey, infoLevelValue, messageKey)
	warnFMsg := fmt.Sprintf(`{%q: %q, %q: "msg-w1: arg-w1-1, arg-w1-2"}`, levelKey, warnLevelValue, messageKey)
	errorFMsg := fmt.Sprintf(`{%q: %q, %q: "msg-e1: arg-e1-1, arg-e1-2"}`, levelKey, errorLevelValue, messageKey)

	debugWMsg := fmt.Sprintf(`{%q: %q, %q: "msg-d2", "arg-d2-1": "arg-d2-2"}`, levelKey, debugLevelValue, messageKey)
	infoWMsg := fmt.Sprintf(`{%q: %q, %q: "msg-i2", "arg-i2-1": "arg-i2-2"}`, levelKey, infoLevelValue, messageKey)
	warnWMsg := fmt.Sprintf(`{%q: %q, %q: "msg-w2", "arg-w2-1": "arg-w2-2"}`, levelKey, warnLevelValue, messageKey)
	errorWMsg := fmt.Sprintf(`{%q: %q, %q: "msg-e2", "arg-e2-1": "arg-e2-2"}`, levelKey, errorLevelValue, messageKey)

	type on struct {
		level  config.LogLevel
		format config.LogFormat
		f      func(Logger)
	}
	type want struct {
		messages []string
		err      error
	}
	testCases := []struct {
		name string
		on   on
		want want
	}{
		{
			"debug level",
			on{config.LogLevelDebug, config.LogFormatJSON, testF},
			want{
				[]string{
					debugFMsg,
					infoFMsg,
					warnFMsg,
					errorFMsg,

					debugWMsg,
					infoWMsg,
					warnWMsg,
					errorWMsg,
				},
				nil,
			},
		},
		{
			"info level",
			on{config.LogLevelInfo, config.LogFormatJSON, testF},
			want{
				[]string{
					infoFMsg,
					warnFMsg,
					errorFMsg,

					infoWMsg,
					warnWMsg,
					errorWMsg,
				},
				nil,
			},
		},
		{
			"warning level",
			on{config.LogLevelWarn, config.LogFormatJSON, testF},
			want{
				[]string{
					warnFMsg,
					errorFMsg,

					warnWMsg,
					errorWMsg,
				},
				nil,
			},
		},
		{
			"error level",
			on{config.LogLevelError, config.LogFormatJSON, testF},
			want{
				[]string{
					errorFMsg,

					errorWMsg,
				},
				nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			bufferedWriter := bufio.NewWriter(&buf)

			logger, err := NewSlogLogger(tc.on.level, tc.on.format, &buf)

			if tc.want.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.want.err)
				return
			}

			require.NoError(t, err)

			tc.on.f(logger)

			err = bufferedWriter.Flush()
			require.NoError(t, err)

			output := buf.String()

			messages := strings.Split(output, "\n")

			require.Equal(t, len(tc.want.messages), len(messages)-1)

			for i, msg := range tc.want.messages {
				if tc.on.format == config.LogFormatJSON {
					assertConstraintJSON(t, msg, messages[i])
				}
			}
		})
	}
}

func assertConstraintJSON(t *testing.T, expectedJSON, actualJSON string) {
	var actual, expected map[string]interface{}

	err := json.Unmarshal([]byte(expectedJSON), &expected)
	require.NoError(t, err, fmt.Sprintf("expectedJSON: %v", expectedJSON))

	err = json.Unmarshal([]byte(actualJSON), &actual)
	require.NoError(t, err, fmt.Sprintf("actualJSON: %v", actualJSON))

	for k, v := range expected {
		assert.Contains(t, actual, k, "Key %s should exist", k)
		assert.Equal(t, v, actual[k], "Value for key %s should match", k)
	}
}
