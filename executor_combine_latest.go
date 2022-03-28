package co

import (
	"sync"
)

func combineLatestRunner(ces ...ConcurrentExecutor) ([]any, error) {
	wg := sync.WaitGroup{}
	wg.Add(len(ces))

	var globalError error = nil
	results := make([]any, len(ces))

	for i := range ces {
		go func(i int) {
			defer wg.Done()
			if globalError != nil {
				return
			}

			data, err := ces[i].latestFnToAny()
			if err != nil {
				globalError = err
				return
			}

			results = append(results, data)
		}(i)
	}

	wg.Wait()
	return results, globalError
}

func CombineLatest[T1, T2 any](co1 *Concurrent[T1], co2 *Concurrent[T2]) (T1, T2, error) {
	results, err := combineLatestRunner([]ConcurrentExecutor{co1, co2}...)

	return results[0].(T1), results[1].(T2), err
}

func CombineLatest3[T1, T2, T3 any](co1 *Concurrent[T1], co2 *Concurrent[T2], co3 *Concurrent[T3]) (T1, T2, T3, error) {
	results, err := combineLatestRunner([]ConcurrentExecutor{co1, co2}...)

	return results[0].(T1), results[1].(T2), results[2].(T3), err
}
