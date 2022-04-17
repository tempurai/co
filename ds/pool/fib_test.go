package pool_test

import (
	"testing"

	"github.com/Jeffail/tunny"
	"tempura.ink/co"
)

func memoizeFib(n int) int {
	cache := make(map[int]int)
	result := make([]int, n)

	for i := 1; i <= n; i++ {
		result[i-1] = refinedExpensiveFib(i, cache)
	}

	return result[n-1]
}

func refinedExpensiveFib(n int, cache map[int]int) int {
	if n < 2 {
		cache[n] = n
		return n
	}
	if _, ok := cache[n-1]; !ok {
		cache[n-1] = refinedExpensiveFib(n-1, cache)
	}
	if _, ok := cache[n-2]; !ok {
		cache[n-2] = refinedExpensiveFib(n-2, cache)
	}
	return cache[n-1] + cache[n-2]
}

func BenchmarkFibSequence(b *testing.B) {
	for i := 1; i < b.N; i++ {
		memoizeFib(i)
	}
}

func BenchmarkFibWithAwaitAll(b *testing.B) {
	handlers := make([]func() (int, error), 0)
	for i := 1; i < b.N; i++ {
		func(idx int) {
			handlers = append(handlers, func() (int, error) {
				return memoizeFib(idx), nil
			})
		}(i)
	}

	co.AwaitAll(handlers...)
}

func BenchmarkFibWithTunny(b *testing.B) {
	pool := tunny.NewFunc(256, func(payload interface{}) interface{} {
		return memoizeFib(payload.(int))
	})
	defer pool.Close()

	b.ResetTimer()
	for i := 1; i < b.N; i++ {
		func(idx int) {
			pool.Process(idx)
		}(i)
	}
}
