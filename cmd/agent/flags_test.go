package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFlags_Default(t *testing.T) {
	addr, polling, report := parseFlags()
	require.Equal(t, ":8080", addr)
	require.Equal(t, 2, polling)
	require.Equal(t, 10, report)
}
