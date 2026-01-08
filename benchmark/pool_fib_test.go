package pool_test

import (
	"sync"
	"testing"

	"github.com/Jeffail/tunny"
	"github.com/panjf2000/ants/v2"
	"github.com/tempurai/co"
	"github.com/tempurai/co/ds/pool"
)

var (
	FibTestParalles int = 512
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

func BenchmarkFibWithSequence(b *testing.B) {
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
	pool := tunny.NewFunc(FibTestParalles, func(payload interface{}) interface{} {
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

func BenchmarkFibWithAnts(b *testing.B) {
	defer ants.Release()

	var wg sync.WaitGroup
	p, _ := ants.NewPool(FibTestParalles)

	b.ResetTimer()
	for i := 1; i < b.N; i++ {
		wg.Add(1)
		p.Submit(func() {
			memoizeFib(i)
			wg.Done()
		})
	}
	wg.Wait()
}

func BenchmarkFibWithWorkPool(b *testing.B) {
	p := pool.NewWorkerPool[int](FibTestParalles)

	b.ResetTimer()
	for i := 1; i < b.N; i++ {
		func(idx int) {
			p.AddJob(func() int {
				return memoizeFib(idx)
			})
		}(i)
	}

	p.Wait()
}

func BenchmarkFibWithDispatchPool(b *testing.B) {
	p := pool.NewDispatchPool[int](FibTestParalles)

	b.ResetTimer()
	for i := 1; i < b.N; i++ {
		func(idx int) {
			p.AddJob(func() int {
				return memoizeFib(idx)
			})
		}(i)
	}

	p.Wait()
}
