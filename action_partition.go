package co

type actionPartition[R any] struct {
	*Action[[]*data[R]]

	its   []Iterator[R]
	width int
}

func (a *actionPartition[R]) ifHasNext() bool {
	for i := range a.its {
		if a.its[i].hasNext() {
			return true
		}
	}
	return false
}

func (a *actionPartition[R]) run() {
	values := make([]*data[R], 0)

	for i := 0; a.ifHasNext(); i++ {
		for j := range a.its {
			val, err := a.its[j].next()
			values = append(values, &data[R]{val, err})

			if len(values) == a.width || !a.ifHasNext() {
				a.listenProgressive(values)
				values = make([]*data[R], 0)
			}
		}
	}

	a.done()
}

func Partition[R any](width int, cos ...AsyncSequenceable[R]) *Action[[]*data[R]] {
	action := &actionPartition[R]{
		its:   toAsyncIterators(cos...),
		width: width,
	}

	SafeGo(action.run)
	return action.Action
}
