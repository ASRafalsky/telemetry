package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFlags_Default(t *testing.T) {
	addr, loglevel, path := parseFlags()
	require.Equal(t, addr, ":8080")
	require.Equal(t, loglevel, "info")
	require.Empty(t, path)
}
