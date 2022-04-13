package co

type AsyncMergedSequence[R any] struct {
	*asyncSequence[R]

	aSequenceables []AsyncSequenceable[R]
}

func NewAsyncMergedSequence[R any](as ...AsyncSequenceable[R]) *AsyncMergedSequence[R] {
	a := &AsyncMergedSequence[R]{
		aSequenceables: as,
	}
	a.asyncSequence = NewAsyncSequence[R](a)
	return a
}

func (a *AsyncMergedSequence[R]) iterator() Iterator[R] {
	it := &asyncMergedSequenceIterator[R]{}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	for i := range a.aSequenceables {
		it.its = append(it.its, a.aSequenceables[i].iterator())
	}
	return it
}

type asyncMergedSequenceIterator[R any] struct {
	*asyncSequenceIterator[R]

	its          []Iterator[R]
	currentIndex int

	previousData *data[R]
	preProcessed bool
}

func (it *asyncMergedSequenceIterator[R]) nextIndex() int {
	defer func() {
		if it.currentIndex+1 >= len(it.its) {
			it.currentIndex = 0
		} else {
			it.currentIndex++
		}
	}()
	return it.currentIndex
}

func (it *asyncMergedSequenceIterator[R]) next() *Optional[R] {
	for range it.its {
		op := it.its[it.nextIndex()].next()
		if !op.valid {
			continue
		}
		return op
	}
	return NewOptionalEmpty[R]()
}
