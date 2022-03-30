package co

import (
	"sync"
)

func All[R any](co Concurrently[R]) *SequenceableData[R] {
	it := co.Iterator()

	seqData := NewSequenceableData[R]()
	wg := sync.WaitGroup{}

	for i := 0; it.hasNext(); i++ {
		wg.Add(1)

		go func(i int, fn executableFn[R]) {
			defer wg.Done()

			val, err := fn()
			seqData.setAt(i, val, err)
		}(i, it.exeFn())
	}

	wg.Wait()
	return seqData
}
