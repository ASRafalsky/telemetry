package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFlags_Default(t *testing.T) {
	addr := parseFlags()
	require.Equal(t, addr, ":8080")
}
