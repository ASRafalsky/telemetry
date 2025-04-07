package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemStorage(t *testing.T) {
	ms := New[int, any]()

	tt := []struct {
		val any
		exp bool
	}{
		{
			val: 1,
			exp: true,
		},
		{
			val: 2.2,
			exp: true,
		},
		{
			val: true,
			exp: true,
		},
		{
			val: "value",
			exp: true,
		},
		{
			val: map[string]interface{}{
				"1": 1,
				"2": 2.2,
				"3": true,
			},
			exp: true,
		},
		{
			val: []string{"1", "2", "3"},
			exp: true,
		},
	}

	for i, tc := range tt {
		ms.Set(i, tc.val)
	}

	keySet := make([]int, 0)
	for i, tc := range tt {
		keySet = append(keySet, i)
		v, ok := ms.Get(i)
		require.True(t, ok)
		require.Equal(t, tc.val, v)
	}

	require.Equal(t, len(keySet), ms.Size())

	ms.Delete(0)
	require.Equal(t, len(keySet)-1, ms.Size())
	_, ok := ms.Get(0)
	require.False(t, ok)
}
