package main

import (
	"bytes"
	"net/http"
)

func newWriterWrapper(w http.ResponseWriter) *writerWrapper {
	return &writerWrapper{
		ResponseWriter: w,
		buffer:         bytes.NewBuffer(nil),
		headerMap:      make(http.Header),
	}
}

type writerWrapper struct {
	http.ResponseWriter
	// to record the status code
	status int
	// to handle immediate response
	buffer    *bytes.Buffer
	headerMap http.Header
}

func (w *writerWrapper) WriteHeader(code int) {
	w.status = code
}

func (w *writerWrapper) Header() http.Header {
	return w.headerMap
}

func (w *writerWrapper) Write(b []byte) (int, error) {
	return w.buffer.Write(b)
}

func (w *writerWrapper) RateLimited() bool {
	return w.status == http.StatusTooManyRequests
}

func (w *writerWrapper) Status() int {
	return w.status
}

func (w *writerWrapper) SendResponse() {
	w.ResponseWriter.WriteHeader(w.status)
	w.ResponseWriter.Write(w.buffer.Bytes())
}

func (w *writerWrapper) Reset() {
	w.buffer.Reset()
	w.status = 0
	w.headerMap = make(http.Header)
}
