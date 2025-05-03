package middleware

import (
	"compress/gzip"
	"net/http"

	"github.com/ASRafalsky/telemetry/internal/compress"
)

func WithCompress(h http.HandlerFunc, log logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ow := w

		compressing := r.Header.Get("Accept-Encoding")
		switch compressing {
		case "gzip":
			cw := compress.NewCompressWriter(w, gzip.NewWriter(w), compressing)
			ow = cw
			defer func() {
				if err := cw.Close(); err != nil {
					log.Error("failed to close gzip writer", err.Error())
				}
			}()
		default:
		}

		switch r.Header.Get("Content-Encoding") {
		case "gzip":
			zr, err := gzip.NewReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = compress.NewCompressReader(r.Body, zr)
		default:
		}
		h.ServeHTTP(ow, r)
	}
}
