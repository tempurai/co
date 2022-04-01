package co

type AsyncMergedSequence[R any] struct {
	*asyncSequence[R]

	aSequenceables []AsyncSequenceable[R]
}

func NewAsyncMergedSequence[R any](as ...AsyncSequenceable[R]) *AsyncMergedSequence[R] {
	a := &AsyncMergedSequence[R]{
		aSequenceables: as,
	}
	a.asyncSequence = NewAsyncSequence(a.Iterator())
	return a
}

func (c *AsyncMergedSequence[R]) Iterator() Iterator[R] {
	it := &asyncMergedSequenceIterator[R]{}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	for i := range c.aSequenceables {
		it.its = append(it.its, c.aSequenceables[i].Iterator())
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

func (it *asyncMergedSequenceIterator[R]) nextAvailableIndex() (int, bool) {
	for range it.its {
		idx := it.nextIndex()
		if it.its[idx].preflight() {
			return idx, true
		}
	}
	return 0, false
}

func (it *asyncMergedSequenceIterator[R]) preflight() bool {
	defer func() { it.preProcessed = true }()

	i, ok := it.nextAvailableIndex()
	if !ok {
		return false
	}

	val, err := it.its[i].consume()
	it.previousData = NewDataWith(val, err)

	return true
}

func (it *asyncMergedSequenceIterator[R]) consume() (R, error) {
	if !it.preProcessed {
		it.preflight()
	}
	defer func() { it.preProcessed = false }()

	return it.previousData.value, it.previousData.err
}
