package co

import (
	"sync"
)

type zipBasic struct {
	combineLatestBasic

	currentIndex int
}

func NewZipBasic(ces []ConcurrentExecutor, fn func([]any, error, bool)) *zipBasic {
	return &zipBasic{
		combineLatestBasic: *NewCombineLatestBasic(ces, fn),
		currentIndex:       1,
	}
}

type zipResult struct {
	combineLatestResult
}

func (z *zipBasic) ifReachesToIndexOrEnd(idx int) bool {
	for i := range z.ces {
		if z.updated[i] != idx && z.updated[i] != z.ces[i].len() {
			return false
		}
	}
	return true
}

func (z *zipBasic) exe() {
	resultChan := make(chan combineLatestResult)

	wg := sync.WaitGroup{}
	wg.Add(len(z.ces))

	cond := sync.Cond{}
	for i := range z.ces {
		ce := z.ces[i]
		wg.Add(ce.len())

		go func(idx int, ce ConcurrentExecutor) {
			defer wg.Done()

			for j := 0; j < ce.len(); j++ {
				cond.Wait()

				data, err := ce.exeFnAt(j)
				SafeSend(resultChan, combineLatestResult{idx, data, err})
			}
		}(i, ce)
	}

	latestResults := make([]any, len(z.ces))
	go func() {
		for {
			select {
			case result := <-resultChan:
				latestResults[result.index] = result.data
				z.updated[result.index]++

				if !EvertGET(z.updated, 1) {
					continue
				}
				if !z.ifReachesToIndexOrEnd(z.currentIndex) {
					continue
				}
				z.currentIndex++

				rte := z.ifReachesToEnd()
				z.fn(latestResults, result.err, rte)

				cond.Broadcast()
				if rte {
					return
				}
			}
		}
	}()

	wg.Wait()
}

func Zip[T1, T2 any](co1 *Concurrent[T1], co2 *Concurrent[T2], fn func(T1, T2, error, bool)) {
	NewZipBasic([]ConcurrentExecutor{co1, co2}, func(a []any, err error, b bool) {
		fn(CastOrNil[T1](a[0]), CastOrNil[T2](a[1]), err, b)
	}).exe()
}

func Zip3[T1, T2, T3 any](co1 *Concurrent[T1], co2 *Concurrent[T2], co3 *Concurrent[T3], fn func(T1, T2, T3, error, bool)) {
	NewZipBasic([]ConcurrentExecutor{co1, co2, co3}, func(a []any, err error, b bool) {
		fn(CastOrNil[T1](a[0]), CastOrNil[T2](a[1]), CastOrNil[T3](a[2]), err, b)
	}).exe()
}
