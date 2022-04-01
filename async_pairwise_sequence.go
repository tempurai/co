package co

type AsyncAdjacentPairsSequence[R any] struct {
	*asyncSequence[[]R]

	previousIterator Iterator[R]
}

func NewAsyncAdjacentPairsSequence[R any](it AsyncSequenceable[R]) *AsyncAdjacentPairsSequence[R] {
	a := &AsyncAdjacentPairsSequence[R]{
		previousIterator: it.Iterator(),
	}
	a.asyncSequence = NewAsyncSequence(a.Iterator())
	return a
}

func (c *AsyncAdjacentPairsSequence[R]) Iterator() Iterator[[]R] {
	it := &asyncAdjacentPairsSequenceIterator[R]{
		AsyncAdjacentPairsSequence: c,
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[[]R](it)
	return it
}

type asyncAdjacentPairsSequenceIterator[R any] struct {
	*asyncSequenceIterator[[]R]

	*AsyncAdjacentPairsSequence[R]

	preProcessed bool
	previousData *data[R]
}

func (it *asyncAdjacentPairsSequenceIterator[R]) preflight() bool {
	defer func() { it.preProcessed = true }()

	if it.previousData == nil && it.previousIterator.preflight() {
		val, err := it.previousIterator.consume()
		it.previousData = NewDataWith(val, err)
		return true
	}
	return false
}

func (it *asyncAdjacentPairsSequenceIterator[R]) consume() ([]R, error) {
	if !it.preProcessed {
		it.preflight()
	}
	defer func() { it.preProcessed = false }()

	previousData := it.previousData

	val, err := it.previousIterator.next()
	if err != nil {
		it.previousData = nil
		return []R{previousData.value}, err
	}
	results := []R{previousData.value, val}

	it.previousData = NewDataWith(val, err)
	return results, nil
}

func (it *asyncAdjacentPairsSequenceIterator[R]) next() ([]R, error) {
	it.preflight()
	return it.consume()
}
