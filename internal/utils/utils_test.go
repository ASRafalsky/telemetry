package utils

import (
	"testing"
	"time"

	"github.com/gojek/heimdall/v7/httpclient"
	"github.com/stretchr/testify/require"
)

func TestGetClientAddr(t *testing.T) {
	addr, err := GetClientAddr(httpclient.NewClient(httpclient.WithHTTPTimeout(10 * time.Second)))
	require.NoError(t, err)
	require.NotEmpty(t, addr)
}

func TestGetAddr(t *testing.T) {
	addr, err := GetAddr("tcp")
	require.NoError(t, err)
	require.NotEmpty(t, addr)
}
