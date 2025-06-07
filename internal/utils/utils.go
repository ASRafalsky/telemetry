package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func SignCheck(t *testing.T, buf, key []byte, sign string) {
	t.Helper()

	if sign != "" && key != nil {
		h := hmac.New(sha256.New, key)
		_, err := h.Write(buf)
		require.NoError(t, err)
		require.Equal(t, hex.EncodeToString(h.Sum(nil)), sign)
	}
}
