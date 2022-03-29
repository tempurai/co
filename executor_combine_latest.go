package co

import (
	"sync"
)

type combineLatestBasic struct {
	ces     []ConcurrentExecutor
	updated []int
	fn      func([]any, error, bool)
}

func NewCombineLatestBasic(ces []ConcurrentExecutor, fn func([]any, error, bool)) *combineLatestBasic {
	return &combineLatestBasic{
		ces:     ces,
		updated: make([]int, len(ces)),
		fn:      fn,
	}
}

func (c *combineLatestBasic) ifReachesToEnd() bool {
	for i := range c.ces {
		if c.updated[i] != c.ces[i].len() {
			return false
		}
	}
	return true
}

type combineLatestResult struct {
	index int
	data  any
	err   error
}

func (c *combineLatestBasic) exe() {
	resultChan := make(chan combineLatestResult)

	wg := sync.WaitGroup{}
	wg.Add(len(c.ces))

	for i := range c.ces {
		ce := c.ces[i]
		wg.Add(ce.len())

		go func(idx int, ce ConcurrentExecutor) {
			defer wg.Done()

			for j := 0; j < ce.len(); j++ {
				data, err := ce.exeAnyFnAt(j)
				SafeSend(resultChan, combineLatestResult{idx, data, err})
			}
		}(i, ce)
	}

	latestResults := make([]any, len(c.ces))
	arrayFilled := false
	go func() {
		for {
			select {
			case result := <-resultChan:
				latestResults[result.index] = result.data
				c.updated[result.index]++

				if !arrayFilled && !EvertGET(c.updated, 1) {
					continue
				}
				arrayFilled = true

				rte := c.ifReachesToEnd()
				c.fn(latestResults, result.err, rte)

				if rte {
					return
				}
			}
		}
	}()

	wg.Wait()
}

func CombineLatest[T1, T2 any](co1 *Concurrent[T1], co2 *Concurrent[T2], fn func(T1, T2, error, bool)) {
	NewCombineLatestBasic([]ConcurrentExecutor{co1, co2}, func(a []any, err error, b bool) {
		fn(CastOrNil[T1](a[0]), CastOrNil[T2](a[1]), err, b)
	}).exe()
}

func CombineLatest3[T1, T2, T3 any](co1 *Concurrent[T1], co2 *Concurrent[T2], co3 *Concurrent[T3], fn func(T1, T2, T3, error, bool)) {
	NewCombineLatestBasic([]ConcurrentExecutor{co1, co2, co3}, func(a []any, err error, b bool) {
		fn(CastOrNil[T1](a[0]), CastOrNil[T2](a[1]), CastOrNil[T3](a[2]), err, b)
	}).exe()
}
