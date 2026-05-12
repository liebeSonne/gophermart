package handler

// Объект, вызывающий ошибку при прочтении
type testErrorReader struct {
	err error
}

func (e *testErrorReader) Read(_ []byte) (n int, err error) {
	return 0, e.err
}
