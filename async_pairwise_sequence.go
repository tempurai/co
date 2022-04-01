package co

type AsyncPairwiseSequence[R any] struct {
	*asyncSequence[[]R]

	previousIterator Iterator[R]
}

func NewAsyncPairwiseSequence[R any](it AsyncSequenceable[R]) *AsyncPairwiseSequence[R] {
	a := &AsyncPairwiseSequence[R]{
		previousIterator: it.Iterator(),
	}
	a.asyncSequence = NewAsyncSequence[[]R](a)
	return a
}

func (c *AsyncPairwiseSequence[R]) Iterator() Iterator[[]R] {
	it := &asyncPairwiseSequenceIterator[R]{
		AsyncPairwiseSequence: c,
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[[]R](it)
	return it
}

type asyncPairwiseSequenceIterator[R any] struct {
	*asyncSequenceIterator[[]R]

	*AsyncPairwiseSequence[R]

	preProcessed bool
	previousData *data[R]
}

func (it *asyncPairwiseSequenceIterator[R]) preflight() bool {
	defer func() { it.preProcessed = true }()

	if it.previousData != nil && it.previousIterator.preflight() {
		return true
	}
	if it.previousData == nil && it.previousIterator.preflight() {
		val, err := it.previousIterator.consume()
		it.previousData = NewDataWith(val, err)
		return true
	}

	return false
}

func (it *asyncPairwiseSequenceIterator[R]) consume() ([]R, error) {
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

func (it *asyncPairwiseSequenceIterator[R]) next() ([]R, error) {
	it.preflight()
	return it.consume()
}
