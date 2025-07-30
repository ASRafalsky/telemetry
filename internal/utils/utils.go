package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// SignCheck compares sha256 hmac with key for data from buf with sign.
// For tests.
func SignCheck(t *testing.T, buf, key []byte, sign string) {
	t.Helper()

	if sign != "" && key != nil {
		h := hmac.New(sha256.New, key)
		_, err := h.Write(buf)
		require.NoError(t, err)
		require.Equal(t, hex.EncodeToString(h.Sum(nil)), sign)
	}
}

// PrintBuildInfo printsuild info.
func PrintBuildInfo(version, date, commit string) {
	if version == "" {
		version = "N/A"
	}
	if date == "" {
		date = "N/A"
	}
	if commit == "" {
		commit = "N/A"
	}
	fmt.Printf("Build version: %s\n", version)
	fmt.Printf("Build date: %s\n", date)
	fmt.Printf("Build commit: %s\n", commit)
}
