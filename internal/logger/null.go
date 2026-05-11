package logger

func NewNullLogger() Logger {
	return &testLogger{}
}

type testLogger struct {
}

func (l *testLogger) Print(v ...interface{}) {
	_ = v
}

func (l *testLogger) Debugf(format string, args ...interface{}) {
	_ = format
	_ = args
}

func (l *testLogger) Infof(format string, args ...interface{}) {
	_ = format
	_ = args
}

func (l *testLogger) Warnf(format string, args ...interface{}) {
	_ = format
	_ = args
}

func (l *testLogger) Errorf(format string, args ...interface{}) {
	_ = format
	_ = args
}

func (l *testLogger) Debugw(msg string, keysAndValues ...interface{}) {
	_ = msg
	_ = keysAndValues
}

func (l *testLogger) Infow(msg string, keysAndValues ...interface{}) {
	_ = msg
	_ = keysAndValues
}

func (l *testLogger) Warnw(msg string, keysAndValues ...interface{}) {
	_ = msg
	_ = keysAndValues
}

func (l *testLogger) Errorw(msg string, keysAndValues ...interface{}) {
	_ = msg
	_ = keysAndValues
}
