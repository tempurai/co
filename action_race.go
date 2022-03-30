package co

type actionRace[R any] struct {
	*Action[*data[R]]

	it        ExecutableIterator[R]
	ignoreErr bool
}

func (a *actionRace[R]) run() {
	dataCh := make(chan *data[R])
	aBool := &AtomicBool{}

	for i := 0; a.it.hasNext(); i++ {
		go func(i int, fn executableFn[R]) {
			if aBool.Get() {
				return
			}

			val, err := fn()
			if err != nil && !a.ignoreErr {
				return
			}

			SafeSend(dataCh, NewDataWith[R](val, err))
			aBool.Set(true)
		}(i, a.it.exeFn())
	}

	a.listenProgressive(<-dataCh)
	close(dataCh)

	a.done()
}

func baseRace[R any](ignoreErr bool, co Concurrently[R]) *Action[*data[R]] {
	action := &actionRace[R]{
		Action:    NewAction[*data[R]](),
		it:        co.Iterator(),
		ignoreErr: ignoreErr,
	}

	SafeGo(action.run)
	return action.Action
}

func Race[R any](it Concurrently[R]) *Action[R] {
	action := baseRace(true, it)
	action.wait()

	return MapAction[*data[R], R](action, func(t *data[R]) R {
		return t.value
	})
}

func Any[R any](it Concurrently[R]) *Action[*data[R]] {
	return baseRace(false, it)
}
