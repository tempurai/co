package co

type actionRace[R any] struct {
	*Action[*data[R]]

	it        executableListIterator[R]
	ignoreErr bool
}

func (a *actionRace[R]) run() {
	dataCh := make(chan *data[R])
	aBool := &AtomicBool{}

	for i := 0; a.it.preflight(); i++ {
		go func(i int) {
			if aBool.Get() {
				return
			}

			val, err := a.it.exeAt(i)
			if (err != nil && !a.ignoreErr) || aBool.Get() {
				return
			}

			SafeSend(dataCh, NewDataWith(val, err))
			aBool.Set(true)
		}(i)
	}

	a.listen(<-dataCh)
	close(dataCh)
	a.done()
}

func baseRace[R any](ignoreErr bool, list *executablesList[R]) *Action[*data[R]] {
	action := &actionRace[R]{
		Action:    NewAction[*data[R]](),
		it:        list.Iterator(),
		ignoreErr: ignoreErr,
	}

	SafeGo(action.run)
	return action.Action
}

func Race[R any](list *executablesList[R]) *Action[R] {
	action := baseRace(true, list)
	// wait first data to complete
	action.PeakData()

	return MapAction(action, func(t *data[R]) R {
		return t.value
	})
}

func Any[R any](list *executablesList[R]) *Action[*data[R]] {
	return baseRace(false, list)
}

func AwaitRace[R any](fns ...func() (R, error)) R {
	return Race(NewExecutablesList[R]().AddExecutable(fns...)).PeakData()
}

func AwaitAny[R any](fns ...func() (R, error)) *data[R] {
	return Any(NewExecutablesList[R]().AddExecutable(fns...)).PeakData()
}
