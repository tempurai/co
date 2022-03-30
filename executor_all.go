package co

import (
	"sync"
)

type ActionAll[R any] struct {
	*Action[*data[R]]
	it ExecutableIterator[R]
}

func (a *ActionAll[R]) run() {
	wg := sync.WaitGroup{}
	sData := NewSequenceableData[R]()

	for i := 0; a.it.hasNext(); i++ {
		wg.Add(1)

		go func(i int, fn executableFn[R]) {
			defer wg.Done()

			val, err := fn()
			sData.setAt(i, val, err)

		}(i, a.it.exeFn())
	}

	wg.Wait()
	a.listenBulk(sData.GetAll())
	a.done()
}

func All[R any](co Concurrently[R]) *Action[*data[R]] {
	action := &ActionAll[R]{
		Action: NewAction[*data[R]](),
		it:     co.Iterator(),
	}

	SafeGo(action.run)
	return action.Action
}
