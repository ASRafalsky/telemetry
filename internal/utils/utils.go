package utils

import (
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
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

func ParsePublicKey(path string) (*rsa.PublicKey, error) {
	if len(path) == 0 {
		return nil, nil
	}

	pubKeyRaw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key from the file: %s, %w", path, err)
	}

	block, _ := pem.Decode(pubKeyRaw)

	pubKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key from the file: %s, %w", path, err)
	}

	return pubKey, nil
}

func ParsePrivateKey(path string) (*rsa.PrivateKey, error) {
	if len(path) == 0 {
		return nil, nil
	}

	privKeyRaw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key from the file: %s, %w", path, err)
	}

	block, _ := pem.Decode(privKeyRaw)
	return x509.ParsePKCS1PrivateKey(block.Bytes)
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
