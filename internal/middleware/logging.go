package middleware

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/multierr"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData responseData
		err          error
	}

	loggingReadCloser struct {
		io.ReadCloser
		err error
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	if err != nil {
		r.err = err
	}
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func (r *loggingReadCloser) Read(b []byte) (int, error) {
	n, err := r.ReadCloser.Read(b)
	if err != nil && err != io.EOF {
		r.err = err
	}
	return n, err
}

func (r *loggingReadCloser) Close() error {
	err := r.ReadCloser.Close()
	if err != nil {
		r.err = multierr.Append(r.err, err)
	}
	return r.err
}

func WithLogging(h http.Handler, l logger) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		url := r.URL.String()
		method := r.Method

		lw := loggingResponseWriter{ResponseWriter: w}
		lr := loggingReadCloser{ReadCloser: r.Body}
		r.Body = &lr
		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		l.Info("[Handler/Request]",
			"url:", url,
			"method:", method,
			"duration:", duration.String(),
		)
		l.Info("[Handler/Response]",
			"status:", strconv.Itoa(lw.responseData.status),
			"size:", strconv.Itoa(lw.responseData.size))
		if lw.err != nil {
			l.Error("[Handler/Error]", lw.err.Error())
		}
		if lr.err != nil {
			l.Error("[Handler/Error]", lr.err.Error())
		}
	}
	return http.HandlerFunc(logFn)
}

type logger interface {
	Info(msg string, add ...string)
	Warn(msg string, add ...string)
	Error(msg string, add ...string)
	Debug(msg string, add ...string)
	Fatal(msg string, add ...string)
}
