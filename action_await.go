package co

import (
	"sync"
)

type actionAwait[R any] struct {
	*Action[*data[R]]

	it Iterator[R]
}

func (a *actionAwait[R]) run() {
	wg := sync.WaitGroup{}
	sData := NewSequenceableData[R]()

	for i := 0; a.it.hasNext(); i++ {
		wg.Add(1)

		go func(i int, fn dispatchFn[R]) {
			defer wg.Done()

			val, err := fn()
			sData.setAt(i, val, err)

		}(i, a.it.dispatch())
	}

	wg.Wait()
	a.listenBulk(sData.GetAll())
	a.done()
}

func Await[R any](co CoSequenceable[R]) *Action[*data[R]] {
	action := &actionAwait[R]{
		Action: NewAction[*data[R]](),
		it:     co.Iterator(),
	}

	SafeGo(action.run)
	return action.Action
}
