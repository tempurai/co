package co

type AsyncPairwiseSequence[R any, T []R] struct {
	*asyncSequence[T]

	previousIterator Iterator[R]
}

func NewAsyncPairwiseSequence[R any, T []R](it AsyncSequenceable[R]) *AsyncPairwiseSequence[R, T] {
	a := &AsyncPairwiseSequence[R, T]{
		previousIterator: it.Iterator(),
	}
	a.asyncSequence = NewAsyncSequence[T](a)
	return a
}

func (c *AsyncPairwiseSequence[R, T]) Iterator() Iterator[T] {
	it := &asyncPairwiseSequenceIterator[R, T]{
		AsyncPairwiseSequence: c,
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[T](it)
	return it
}

type asyncPairwiseSequenceIterator[R any, T []R] struct {
	*asyncSequenceIterator[T]

	*AsyncPairwiseSequence[R, T]

	preProcessed bool
	previousData *data[R]
}

func (it *asyncPairwiseSequenceIterator[R, T]) preflight() bool {
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

func (it *asyncPairwiseSequenceIterator[R, T]) consume() (T, error) {
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

func (it *asyncPairwiseSequenceIterator[R, T]) next() (T, error) {
	it.preflight()
	return it.consume()
}
