package handlers

import (
	"net/http"
	"time"

	"github.com/ASRafalsky/telemetry/internal/log"
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

func WithLogging(h http.Handler, logger *log.Logger) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		url := r.URL.String()
		method := r.Method

		lw := loggingResponseWriter{ResponseWriter: w}
		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		logger.Info("[Handler/Request]",
			log.StringField("url", url),
			log.StringField("method", method),
			log.DurationField("duration", duration),
		)
		logger.Info("[Handler/Response]",
			log.IntField("status", lw.responseData.status),
			log.IntField("size", lw.responseData.size),
		)
	}
	return http.HandlerFunc(logFn)
}
