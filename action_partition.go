package co

type actionPartition[R any] struct {
	*Action[[]*data[R]]

	its   []ExecutableIterator[R]
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
	seqData := NewSequenceableData[R]()

	for i := 0; a.ifHasNext(); i++ {
		for j := range a.its {
			data, err := a.its[j].exeNext()
			seqData.add(data, err)

			if seqData.len() == a.width || !a.ifHasNext() {
				a.listenProgressive(seqData.GetAll())
				seqData = NewSequenceableData[R]()
			}
		}
	}

	a.done()
}

func Partition[R any](width int, cos ...Concurrently[R]) *Action[[]*data[R]] {
	action := &actionPartition[R]{
		its:   toConcurrentIterators(cos...),
		width: width,
	}

	SafeGo(action.run)
	return action.Action
}
