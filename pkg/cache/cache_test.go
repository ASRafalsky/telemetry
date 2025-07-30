package cache

import (
	"context"
	"fmt"
	"slices"
	"sync"
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

func Example() {
	const cnt = 10
	ms := New[int, int]()

	var keys []int
	for i := range cnt {
		keys = append(keys, i)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i, k := range keys {
			ms.Set(k, i)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for k := range keys {
			ms.Get(k)
		}
	}()

	wg.Wait()

	var evenKeyValueList []int
	_ = ms.ForEach(context.Background(), func(k int, v int) error {
		if k%2 == 0 {
			return nil
		}
		evenKeyValueList = append(evenKeyValueList, v)
		return nil
	})

	slices.Sort(evenKeyValueList)
	for _, k := range evenKeyValueList {
		fmt.Println(k)
	}

	fmt.Println("Get:")
	for _, k := range evenKeyValueList {
		if v, ok := ms.Get(k); ok {
			fmt.Println(k, v)
		}
	}

	// Drop each value for odd key.
	_ = ms.DropFn(context.Background(), func(k int, v int) (bool, error) {
		if k%2 != 0 {
			return false, nil
		}
		return true, nil
	})

	fmt.Println("Size after drop even:")
	fmt.Println(ms.Size())

	// Delete even values
	for _, k := range evenKeyValueList {
		ms.Delete(k)
	}
	fmt.Println("Size after delete all even:")
	fmt.Println(ms.Size())
	// Output:
	// 1
	// 3
	// 5
	// 7
	// 9
	// Get:
	// 1 1
	// 3 3
	// 5 5
	// 7 7
	// 9 9
	// Size after drop even:
	// 5
	// Size after delete all even:
	// 0
}
