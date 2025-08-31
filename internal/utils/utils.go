package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/gojek/heimdall/v7/httpclient"
	"github.com/mailru/easyjson"
	"github.com/stretchr/testify/require"

	"github.com/ASRafalsky/telemetry/internal/transport"
	"github.com/ASRafalsky/telemetry/internal/types"
)

// SignCheckTest compares sha256 hmac with key for data from buf with sign.
// For tests.
func SignCheckTest(t *testing.T, buf, key []byte, sign string) {
	t.Helper()
	ok, err := SignCheck(buf, key, sign)
	require.NoError(t, err)
	require.True(t, ok, fmt.Sprintf("bufSz: %d, keySz: %d, sign: %s", len(buf), len(key), sign))
}

// SignCheck compares sha256 hmac with key for data from buf with sign.
func SignCheck(buf, key []byte, sign string) (bool, error) {
	if sign != "" && key != nil && len(buf) > 0 {
		h := hmac.New(sha256.New, key)
		_, err := h.Write(buf)
		if err != nil {
			return false, err
		}
		return hex.EncodeToString(h.Sum(nil)) == sign, nil
	}
	return false, nil
}

// Sign returns signature string for data signed with key.
func Sign(data []byte, key []byte) (string, error) {
	if len(key) == 0 {
		return "", nil
	}
	h := hmac.New(sha256.New, key)
	if _, err := h.Write(data); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func Encrypt(key *rsa.PublicKey, data []byte) ([]byte, error) {
	if key == nil {
		return data, nil
	}
	cipherdata, err := rsa.EncryptPKCS1v15(rand.Reader, key, data)
	if err != nil {
		return nil, err
	}
	return cipherdata, nil
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

// GetClientAddr returns local addr if it is possible.
func GetClientAddr(client *httpclient.Client) (string, error) {
	var clientAddr string
	// Create test server. Yes, I know. Change my mind.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientAddr = r.RemoteAddr
		_ = r.Body.Close()
	}))
	defer srv.Close()
	r, err := client.Get(srv.URL, http.Header{})
	if err != nil {
		return "", err
	}
	defer func() {
		_ = r.Body.Close()
	}()
	idx := strings.LastIndex(clientAddr, ":")
	if idx == -1 {
		return "", fmt.Errorf("invalid address: %s", clientAddr)
	}
	return clientAddr[:idx], nil
}

// GetAddr is another way to get client addr. Which method is correct?
func GetAddr(network string) (string, error) {
	list, err := net.Listen(network, ":0")
	if err != nil {
		return "", err
	}
	defer func() {
		_ = list.Close()
	}()

	conn, err := net.Dial(network, list.Addr().String())
	if err != nil {
		return "", err
	}
	addr := conn.LocalAddr().String()
	idx := strings.LastIndex(addr, ":")
	if idx == -1 {
		return "", fmt.Errorf("invalid address: %s", addr)
	}
	return strings.Trim(addr[:idx], "[]"), nil
}

func MetricsMap(t *testing.T) (mMap map[string][]byte, vKeys, dKeys []string) {
	values := []float64{123.123, 456, 789.987}
	mMap = make(map[string][]byte)
	for n, value := range values {
		vKey := "value" + strconv.Itoa(n)
		vKeys = append(vKeys, vKey)
		buf, err := easyjson.Marshal(transport.Metrics{
			MType: types.GaugeType,
			ID:    vKey,
			Value: &value,
		})
		require.NoError(t, err)
		mMap[types.GaugeType+vKey] = buf
		dKey := "delta" + strconv.Itoa(n)
		dKeys = append(dKeys, dKey)
		delta := int64(n)
		buf, err = easyjson.Marshal(transport.Metrics{
			MType: types.CounterType,
			ID:    dKey,
			Delta: &delta,
		})
		require.NoError(t, err)
		mMap[types.CounterType+dKey] = buf
	}
	return
}
