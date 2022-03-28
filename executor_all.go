package co

import "sync"

func All[R any](co *Concurrent[R]) *SequenceableData[R] {
	wg := sync.WaitGroup{}
	wg.Add(len(co.executors))

	for i := range co.executors {
		go func(i int) {
			defer wg.Done()
			if co.executors[i].isExecuted() {
				return
			}
			co.executors[i].exe()
		}(i)
	}

	wg.Wait()
	sData := NewSequenceableData[R]()
	sData.executorSwap(co.executors)

	return sData
}
