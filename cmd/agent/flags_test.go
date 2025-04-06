package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFlags_Default(t *testing.T) {
	addr, polling, report := parseFlags()
	require.Equal(t, addr, "http://:8080")
	require.Equal(t, polling, 2)
	require.Equal(t, report, 10)
}
