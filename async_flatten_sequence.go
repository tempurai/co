package co

type AsyncFlattenSequence[R any, T []R] struct {
	*asyncSequence[R]

	previousIterator Iterator[T]
}

func NewAsyncFlattenSequence[R any, T []R](it AsyncSequenceable[T]) *AsyncFlattenSequence[R, T] {
	a := &AsyncFlattenSequence[R, T]{
		previousIterator: it.iterator(),
	}
	a.asyncSequence = NewAsyncSequence[R](a)
	return a
}

func (c *AsyncFlattenSequence[R, T]) iterator() Iterator[R] {
	it := &asyncFlattenSequenceIterator[R, T]{
		AsyncFlattenSequence: c,
		bufferedData:         make(T, 0),
	}
	it.asyncSequenceIterator = NewAsyncSequenceIterator[R](it)
	return it
}

type asyncFlattenSequenceIterator[R any, T []R] struct {
	*asyncSequenceIterator[R]

	*AsyncFlattenSequence[R, T]
	bufferedData T // Should to optimze with linked list
}

func (it *asyncFlattenSequenceIterator[R, T]) preflight() bool {
	for op, err := it.previousIterator.next(); op.valid; op, err = it.previousIterator.next() {
		if err != nil {
			continue
		}

		if len(op.data) > 0 {
			it.bufferedData = append(it.bufferedData, op.data...)
			return true
		}
	}
	return false
}

func (it *asyncFlattenSequenceIterator[R, T]) next() (*Optional[R], error) {
	if len(it.bufferedData) == 0 {
		if !it.preflight() {
			return NewOptionalEmpty[R](), nil
		}
	}

	result, nextBuffer := it.bufferedData[0], it.bufferedData[1:]
	it.bufferedData = nextBuffer

	return OptionalOf(result), nil
}
