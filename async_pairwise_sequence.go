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

	previousData *R
}

func (it *asyncPairwiseSequenceIterator[R, T]) preflight() {
	if it.previousData != nil {
		return
	}

	for op, err := it.previousIterator.next(); op.valid; op, err = it.previousIterator.next() {
		if err != nil {
			continue
		}
		it.previousData = &op.data
		break
	}
}

func (it *asyncPairwiseSequenceIterator[R, T]) next() (*Optional[T], error) {
	it.preflight()

	previousData := *it.previousData
	for op, err := it.previousIterator.next(); op.valid; op, err = it.previousIterator.next() {
		if err != nil {
			return NewOptionalEmpty[T](), nil
		}
		it.previousData = &op.data
		return OptionalOf(T{previousData, op.data}), nil
	}

	return NewOptionalEmpty[T](), nil
}
