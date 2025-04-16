package handlers

import (
	"compress/gzip"
	"net/http"

	"github.com/ASRafalsky/telemetry/internal/compress"
)

func WithCompress(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ow := w

		compressing := r.Header.Get("Accept-Encoding")
		switch compressing {
		case "gzip":
			cw := compress.NewCompressWriter(w, gzip.NewWriter(w), compressing)
			ow = cw
			defer cw.Close()
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
			defer r.Body.Close()
		default:

		}
		h.ServeHTTP(ow, r)
	}
}
