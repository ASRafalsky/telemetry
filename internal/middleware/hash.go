package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"net/http"
)

type hashWriter struct {
	w       http.ResponseWriter
	hw      hasher
	hashAlg string
}

type hasher interface {
	io.Writer
	Sum(b []byte) []byte
}

func newHashWriter(w http.ResponseWriter, hash hash.Hash, hashAlg string) *hashWriter {
	return &hashWriter{
		w:       w,
		hw:      hash,
		hashAlg: hashAlg,
	}
}

func (c *hashWriter) Header() http.Header {
	return c.w.Header()
}

func (c *hashWriter) Write(p []byte) (int, error) {
	if _, err := c.hw.Write(p); err != nil {
		return 0, err
	}
	return c.w.Write(p)
}

func (c *hashWriter) WriteHeader(statusCode int) {
	c.w.Header().Set(c.hashAlg, hex.EncodeToString(c.hw.Sum(nil)))
	c.w.WriteHeader(statusCode)
}

func WithSign(h http.HandlerFunc, key []byte, log logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sign := r.Header.Get("HashSHA256")
		var signHash hash.Hash
		if len(key) == 0 && sign != "" {
			w.WriteHeader(http.StatusBadRequest)
		}
		if len(key) > 0 {
			signHash = hmac.New(sha256.New, key)
			w = newHashWriter(w, signHash, "HashSHA256")
		}

		if sign != "" {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))
			if _, err := signHash.Write(body); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Error("Failed to write hash", err.Error())
				return
			}
			hexSign, err := hex.DecodeString(sign)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if !hmac.Equal(signHash.Sum(nil), hexSign) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		h.ServeHTTP(w, r)
	}
}
