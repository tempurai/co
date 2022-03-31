package co

type actionRace[R any] struct {
	*Action[*data[R]]

	it        Iterator[R]
	ignoreErr bool
}

func (a *actionRace[R]) run() {
	dataCh := make(chan *data[R])
	aBool := &AtomicBool{}

	for i := 0; a.it.hasNext(); i++ {
		go func(i int, fn dispatchFn[R]) {
			if aBool.Get() {
				return
			}

			val, err := fn()
			if err != nil && !a.ignoreErr {
				return
			}

			SafeSend(dataCh, NewDataWith(val, err))
			aBool.Set(true)
		}(i, a.it.dispatch())
	}

	a.listenProgressive(<-dataCh)
	close(dataCh)

	a.done()
}

func baseRace[R any](ignoreErr bool, co CoSequenceable[R]) *Action[*data[R]] {
	action := &actionRace[R]{
		Action:    NewAction[*data[R]](),
		it:        co.Iterator(),
		ignoreErr: ignoreErr,
	}

	SafeGo(action.run)
	return action.Action
}

func Race[R any](it CoSequenceable[R]) *Action[R] {
	action := baseRace(true, it)
	action.wait()

	return MapAction(action, func(t *data[R]) R {
		return t.value
	})
}

func Any[R any](it CoSequenceable[R]) *Action[*data[R]] {
	return baseRace(false, it)
}
