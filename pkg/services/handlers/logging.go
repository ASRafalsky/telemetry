package handlers

import (
	"net/http"
	"strconv"
	"time"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func WithLogging(h http.Handler, l logger) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		url := r.URL.String()
		method := r.Method

		lw := loggingResponseWriter{ResponseWriter: w}
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
	}
	return http.HandlerFunc(logFn)
}

type logger interface {
	Info(msg ...string)
	Warn(msg ...string)
	Error(msg ...string)
	Debug(msg ...string)
	Fatal(msg ...string)
}
