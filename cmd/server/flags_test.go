package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFlags_Default(t *testing.T) {
	address, logLevel, logpPath, dumpPath, storePeriod, restore := parseFlags()
	require.Equal(t, address, ":8080")
	require.Equal(t, logLevel, "info")
	require.Empty(t, logpPath)
	require.Equal(t, "./dump/dump", dumpPath)
	require.Equal(t, 300, storePeriod)
	require.False(t, restore)
}
